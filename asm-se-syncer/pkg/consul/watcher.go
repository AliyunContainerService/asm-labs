package consul

import (
	"context"
	log "github.com/sirupsen/logrus"
	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/common"
	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/serviceentry"
	"net/url"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"istio.io/api/networking/v1alpha3"

	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/provider"
)

var errIndexChangeTimeout = errors.New("blocking request timeout while waiting for index to change")

type watcher struct {
	client          *api.Client
	store           provider.Cache
	tickInterval    time.Duration
	lastIndex       uint64 // lastly synced index of Catalog
	consulNamespace string
	prefix          string
	toNamespace     string
	watcherType     string
}

const (
	// TODO: allow users to specify these
	defaultBlockingRequestWaitTimeDuration = 5 * time.Second
	defaultTickIntervalDuration            = 10 * time.Second
)

var _ provider.Watcher = &watcher{}

func NewWatcher(store provider.Cache, endpoint string, consulNamespace, prefix, toNamespace string) (provider.Watcher, error) {
	if len(endpoint) == 0 {
		return nil, errors.New("Consul endpoint not specified")
	}

	config := api.DefaultConfig()
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing endpoint: %s", endpoint)
	}

	// TODO: allow users to specify TOKEN
	config.Scheme = u.Scheme
	config.Address = u.Host
	config.WaitTime = defaultBlockingRequestWaitTimeDuration

	log.Infof("watching config: %+v", config)
	client, err := api.NewClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "error creating client")
	}
	return &watcher{client: client,
		store:           store,
		tickInterval:    defaultTickIntervalDuration,
		consulNamespace: consulNamespace,
		prefix:          prefix,
		watcherType:     string(common.Consul),
		toNamespace:     toNamespace,
	}, nil
}

func (w *watcher) Cache() provider.Cache {
	return w.store
}

func (w *watcher) Prefix() string {
	return w.prefix
}

func (w *watcher) ToNamespace() string {
	return w.toNamespace
}
func (w *watcher) WatcherType() string {
	return w.watcherType
}

// Run the watcher until the context is cancelled
func (w *watcher) Run(ctx context.Context) {
	ticker := time.NewTicker(w.tickInterval)
	defer ticker.Stop()

	w.refreshStore(w.Prefix()) // init
	for {
		select {
		case <-ticker.C:
			w.refreshStore(w.Prefix())
		case <-ctx.Done():
			return
		}
	}
}

// fetch services and endpoints from consul catalog and sync them with Store
func (w *watcher) refreshStore(prefix string) {
	names, err := w.listServices()
	if err == errIndexChangeTimeout {
		log.Infof("waiting for index to change: current index: %d", w.lastIndex)
		return
	} else if err != nil {
		log.Errorf("error listing services from Consul: %v", err)
		return
	}

	css := w.describeServices(names)
	data := make(map[string][]*v1alpha3.WorkloadEntry, len(css))
	for name, cs := range css {
		name = common.FormatedName(name)
		eps := make([]*v1alpha3.WorkloadEntry, 0, len(cs))
		for _, c := range cs {
			if ep := catalogServiceToEndpoints(c); ep != nil {
				eps = append(eps, ep)
			}
		}
		if len(eps) > 0 {
			data[prefix+name] = eps
		}
	}
	w.store.Set(data)
}

// listServices lists services
func (w *watcher) listServices() (map[string][]string, error) {
	data, metadata, err := w.client.Catalog().Services(
		&api.QueryOptions{WaitIndex: w.lastIndex, Namespace: w.consulNamespace},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list services")
	}

	if w.lastIndex == metadata.LastIndex {
		// this case indicates the request reaches timeout of blocking request
		return nil, errIndexChangeTimeout
	}

	w.lastIndex = metadata.LastIndex
	return data, nil
}

// describeServices gets catalog services for given service names
func (w *watcher) describeServices(names map[string][]string) map[string][]*api.CatalogService {
	ss := make(map[string][]*api.CatalogService, len(names))
	for name := range names { // ignore tags in value
		svcs, err := w.describeService(name)
		if err != nil {
			log.Errorf("error describing service catalog from Consul: %v ", err)
			continue
		}
		ss[name] = svcs
	}
	return ss
}

func (w *watcher) describeService(name string) ([]*api.CatalogService, error) {
	svcs, _, err := w.client.Catalog().Service(name, "", &api.QueryOptions{
		Namespace: w.consulNamespace,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to describe svc: %s", name)
	}
	return svcs, nil
}

// catalogServiceToEndpoints converts catalog service to service entry endpoint
func catalogServiceToEndpoints(c *api.CatalogService) *v1alpha3.WorkloadEntry {
	address := c.ServiceAddress
	if address == "" {
		address = c.Address
	}

	if address == "" {
		log.Infof("instance %s of %s.%v is of a type that is not currently supported",
			c.ServiceID, c.ServiceName, c.Namespace)
		return nil
	}

	port := c.ServicePort
	if port > 0 { // port is optional and defaults to zero
		return serviceentry.Endpoint(address, uint32(port))
	}

	log.Infof("no port found for address %v, assuming http (80) and https (443)", address)
	return &v1alpha3.WorkloadEntry{Address: address, Ports: map[string]uint32{"http": 80, "https": 443}}
}
