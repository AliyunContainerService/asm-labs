package control

import (
	"context"
	log "github.com/sirupsen/logrus"
	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/common"
	"reflect"
	"strings"
	"time"

	"istio.io/api/networking/v1alpha3"
	icapi "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1alpha3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/provider"
	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/serviceentry"
)

type synchronizer struct {
	namespace          string
	serviceEntry       serviceentry.ServiceEntryModel
	store              provider.Cache
	serviceEntryPrefix string
	location           v1alpha3.ServiceEntry_Location
	client             icapi.ServiceEntryInterface
	interval           time.Duration
}

func NewSynchronizer(namespace string,
	serviceEntry serviceentry.ServiceEntryModel, store provider.Cache, serviceEntryPrefix string, location v1alpha3.ServiceEntry_Location, interval time.Duration, client icapi.ServiceEntryInterface) *synchronizer {
	return &synchronizer{
		namespace:          namespace,
		serviceEntry:       serviceEntry,
		store:              store,
		serviceEntryPrefix: serviceEntryPrefix,
		location:           location,
		client:             client,
		interval:           interval,
	}
}

// Run the synchronizer until the context is cancelled
func (s *synchronizer) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.sync()
		case <-ctx.Done():
			return
		}
	}
}

func (s *synchronizer) sync() {
	// Entries are generated per host; entirely from information in the slice of endpoints;
	// so we only actually need to compare the current endpoints with the new endpoints.
	for host, endpoints := range s.store.Hosts() {
		s.createOrUpdate(host, endpoints)
	}
	s.garbageCollect()
}

func (s *synchronizer) createOrUpdate(host string, endpoints []*v1alpha3.WorkloadEntry) {
	newServiceEntry := serviceentry.Builder(s.namespace, s.serviceEntryPrefix, host, s.location, endpoints)
	name := common.FormatedName(host)
	if _, ok := s.serviceEntry.Ours()[host]; ok {
		// If we have already created an identical service entry, return.
		if reflect.DeepEqual(s.serviceEntry.Ours()[host].Spec.Endpoints, endpoints) {
			return
		}
		// Otherwise, endpoints have changed so update existing Service Entry
		oldServiceEntry, err := s.client.Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			log.Errorf("failed to get existing service entry %q for host %q, errMsg %v", name, host, err)
			return
		}
		newServiceEntry.ResourceVersion = oldServiceEntry.ResourceVersion
		rv, err := s.client.Update(context.TODO(), newServiceEntry, v1.UpdateOptions{})
		if err != nil {
			log.Errorf("error updating Service Entry %q: %v", name, err)
			return
		}
		log.Infof("updated Service Entry %q, ResourceVersion is now %q, host: %s, prefix: %s", name, rv.ResourceVersion, host, s.serviceEntryPrefix)
		return
	}
	// Otherwise, create a new Service Entry
	rv, err := s.client.Create(context.TODO(), newServiceEntry, v1.CreateOptions{})
	if err != nil {
		log.Errorf("error creating Service Entry %q: %v\n%v", name, err, newServiceEntry)
	}
	log.Infof("created Service Entry %q, ResourceVersion is %q, host: %s, prefix: %s", name, rv.ResourceVersion, host, s.serviceEntryPrefix)
}

func (s *synchronizer) garbageCollect() {
	for host := range s.serviceEntry.Ours() {
		//skip host not belong to synchronizer prefix
		if !strings.HasPrefix(host, s.serviceEntryPrefix) {
			continue
		}
		// If host no longer exists, delete service entry
		allHost := []string{}
		for k := range s.store.Hosts() {
			allHost = append(allHost, k)
		}
		if _, ok := s.store.Hosts()[host]; !ok {
			name := host
			//name := serviceentry.ServiceEntryName(s.serviceEntryPrefix, host)
			oldServiceEntry, err := s.client.Get(context.TODO(), name, v1.GetOptions{})
			// check the ASM label
			if err != nil {
				log.Error(err, "get oldServiceEntry get failed")
			}
			if oldServiceEntry.Labels[common.AsmSyncerLabel] != string(common.Consul) {
				continue
			}
			if err := s.client.Delete(context.TODO(), name, v1.DeleteOptions{}); err != nil {
				log.Errorf("error deleting Service Entry %q: %v", name, err)
			}
			log.Infof("successfully deleted Service Entry %q", name)
		}
	}
}
