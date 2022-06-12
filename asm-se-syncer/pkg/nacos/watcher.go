package nacos

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cenkalti/backoff"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	log "github.com/sirupsen/logrus"
	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/common"
	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/provider"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"io/ioutil"
	mcp "istio.io/api/mcp/v1alpha1"
	"istio.io/api/mesh/v1alpha1"
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/serviceregistry/memory"
	v3 "istio.io/istio/pilot/pkg/xds/v3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/collections"
	"k8s.io/client-go/rest"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	grpcInitialWindowSize     = 1 << 30
	grpcInitialConnWindowSize = 1 << 30
)

// ADSC implements a basic client for ADS, for use in stress tests and tools
// or libraries that need to connect to Istio pilot or other ADS servers.
type ADSC struct {
	// Stream is the GRPC connection stream, allowing direct GRPC send operations.
	// Set after Dial is called.
	stream discovery.AggregatedDiscoveryService_StreamAggregatedResourcesClient
	// xds client used to create a stream
	client discovery.AggregatedDiscoveryServiceClient
	conn   *grpc.ClientConn

	// Indicates if the ADSC client is closed
	closed bool

	// NodeID is the node identity sent to Pilot.
	nodeID string

	url string

	toNamespace string

	watcherType string

	watchTime time.Time

	// InitialLoad tracks the time to receive the initial configuration.
	InitialLoad time.Duration

	// httpListeners contains received listeners with a http_connection_manager filter.
	httpListeners map[string]*listener.Listener

	// tcpListeners contains all listeners of type TCP (not-HTTP)
	tcpListeners map[string]*listener.Listener

	// All received clusters of type eds, keyed by name
	edsClusters map[string]*cluster.Cluster

	// All received clusters of no-eds type, keyed by name
	clusters map[string]*cluster.Cluster

	// All received routes, keyed by route name
	routes map[string]*route.RouteConfiguration

	// All received endpoints, keyed by cluster name
	eds map[string]*endpoint.ClusterLoadAssignment

	// Metadata has the node metadata to send to pilot.
	// If nil, the defaults will be used.
	Metadata *pstruct.Struct

	// Updates includes the type of the last update received from the server.
	Updates     chan string
	XDSUpdates  chan *discovery.DiscoveryResponse
	VersionInfo map[string]string

	// Last received message, by type
	Received map[string]*discovery.DiscoveryResponse

	mutex sync.RWMutex

	Mesh *v1alpha1.MeshConfig

	nacosNamespace string

	// Retrieved configurations can be stored using the common istio model interface.
	Store model.IstioConfigStore

	// Retrieved endpoints can be stored in the memory registry. This is used for CDS and EDS responses.
	Registry *memory.ServiceDiscovery

	// LocalCacheDir is set to a base name used to save fetched resources.
	// If set, each update will be saved.
	// TODO: also load at startup - so we can support warm up in init-container, and survive
	// restarts.
	LocalCacheDir string

	// RecvWg is for letting goroutines know when the goroutine handling the ADS stream finishes.
	RecvWg sync.WaitGroup

	cfg *Config

	ServiceEntry map[string]string

	// sendNodeMeta is set to true if the connection is new - and we need to send node meta.,
	sendNodeMeta                   bool
	sync                           map[string]time.Time
	syncCh                         chan string
	Locality                       *core.Locality
	IstioClient                    *IstioClient
	CreateOrUpdateServiceEntryChan chan *v1alpha3.ServiceEntry
	//UpdateServiceEntryChan chan *v1alpha3.ServiceEntry
	DeleteServiceEntryChan chan *v1alpha3.ServiceEntry
}

func NewWatcher(endpoint string, opts *Config, istioConfig *rest.Config, nacosNamespace, toNamespace string, store model.ConfigStoreCache) (*ADSC, error) {
	if opts == nil {
		opts = &Config{}
	}
	// We want to recreate stream
	if opts.BackoffPolicy == nil {
		opts.BackoffPolicy = backoff.NewExponentialBackOff()
	}
	istioClient, err := NewClient(istioConfig)
	if err != nil {
		return nil, err
	}
	log.Info("create nacos istio client success")
	adsc := &ADSC{
		Updates:                        make(chan string, 100),
		XDSUpdates:                     make(chan *discovery.DiscoveryResponse, 100),
		VersionInfo:                    map[string]string{},
		toNamespace:                    toNamespace,
		nacosNamespace:                 nacosNamespace,
		watcherType:                    string(common.Nacos),
		url:                            endpoint,
		Received:                       map[string]*discovery.DiscoveryResponse{},
		RecvWg:                         sync.WaitGroup{},
		cfg:                            opts,
		syncCh:                         make(chan string, len(collections.Pilot.All())),
		sync:                           map[string]time.Time{},
		IstioClient:                    istioClient,
		CreateOrUpdateServiceEntryChan: make(chan *v1alpha3.ServiceEntry, 50),
		Store:                          model.MakeIstioStore(store),
		ServiceEntry:                   make(map[string]string),
		//UpdateServiceEntryChan: make(chan *v1alpha3.ServiceEntry, 10),
		DeleteServiceEntryChan: make(chan *v1alpha3.ServiceEntry, 50),
	}

	if opts.Namespace == "" {
		opts.Namespace = "default"
	}
	if opts.NodeType == "" {
		opts.NodeType = "sidecar"
	}
	if opts.IP == "" {
		opts.IP = getPrivateIPIfAvailable().String()
	}
	if opts.Workload == "" {
		opts.Workload = "test-1"
	}
	adsc.Metadata = opts.Meta
	adsc.Locality = opts.Locality

	adsc.nodeID = fmt.Sprintf("%s~%s~%s.%s~%s.svc.cluster.local", opts.NodeType, opts.IP,
		opts.Workload, opts.Namespace, opts.Namespace)

	if err := adsc.Dial(); err != nil {
		return nil, err
	}

	adsc.client = discovery.NewAggregatedDiscoveryServiceClient(adsc.conn)
	adsc.stream, err = adsc.client.StreamAggregatedResources(context.Background())
	if err != nil {
		log.Error("can not get stream err is %s", err.Error())
		return nil, err
	}
	log.Info("new stream success")
	adsc.sendNodeMeta = true
	adsc.InitialLoad = 0
	if err := adsc.Send(&discovery.DiscoveryRequest{
		TypeUrl: collections.IstioNetworkingV1Alpha3Serviceentries.Resource().GroupVersionKind().String(),
	}); err != nil {
		return adsc, err
	}
	return adsc, nil
}

// Raw send of a request.
func (a *ADSC) Send(req *discovery.DiscoveryRequest) error {
	if a.sendNodeMeta {
		req.Node = a.node()
		a.sendNodeMeta = false
	}
	req.ResponseNonce = time.Now().String()
	return a.stream.Send(req)
}

func (a *ADSC) ToNamespace() string {
	return a.toNamespace
}
func (a *ADSC) WatcherType() string {
	return a.watcherType
}
func (a *ADSC) Cache() provider.Cache {
	return nil
}

func (a *ADSC) Prefix() string {
	return ""
}

func (a *ADSC) Run(ctx context.Context) {
	// Send the initial requests
	/*for _, r := range a.cfg.InitialDiscoveryRequests {
		if r.TypeUrl == v3.ClusterType {
			a.watchTime = time.Now()
		}
		_ = a.Send(r)
	}*/
	// by default, we assume 1 goroutine decrements the waitgroup (go a.handleRecv()).
	// for synchronizing when the goroutine finishes reading from the gRPC stream.
	a.RecvWg.Add(1)
	go a.handleRecv()
	a.RecvWg.Add(1)
	go a.handleServiceEntry()
}

func (a *ADSC) Dial() error {
	opts := a.cfg

	var err error
	grpcDialOptions := opts.GrpcOpts
	if len(grpcDialOptions) == 0 {
		// Only disable transport security if the user didn't supply custom dial options
		grpcDialOptions = append(grpcDialOptions, grpc.WithInsecure(),
			grpc.WithInitialWindowSize(grpcInitialWindowSize),
			grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
			grpc.WithReadBufferSize(30*1024*1024),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*30)),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(1024*1024*30)),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                30 * time.Second,
				Timeout:             10 * time.Second,
				PermitWithoutStream: true,
			}))
	}

	a.conn, err = grpc.Dial(a.url, grpcDialOptions...)
	if err != nil {
		return err
	}
	return nil
}

func (a *ADSC) handleRecv() {
	for {
		var err error
		msg, err := a.stream.Recv()
		if err != nil {
			a.RecvWg.Done()
			log.Info("Connection closed for node %v with err: %v", a.nodeID, err)
			// if 'reconnect' enabled - schedule a new Run
			if a.cfg.BackoffPolicy != nil {
				time.AfterFunc(a.cfg.BackoffPolicy.NextBackOff(), a.reconnect)
			} else {
				a.Close()
				a.WaitClear()
				a.Updates <- ""
				a.XDSUpdates <- nil
			}
			return
		}

		// Group-value-kind - used for high level api generator.
		gvk := strings.SplitN(msg.TypeUrl, "/", 3)

		log.Info("Received ", a.url, " type ", msg.TypeUrl,
			" cnt=", len(msg.Resources), " nonce=", msg.Nonce)
		if a.cfg.ResponseHandler != nil {
			a.cfg.ResponseHandler.HandleResponse(a, msg)
		}

		if msg.TypeUrl == collections.IstioMeshV1Alpha1MeshConfig.Resource().GroupVersionKind().String() &&
			len(msg.Resources) > 0 {
			rsc := msg.Resources[0]
			m := &v1alpha1.MeshConfig{}
			err = proto.Unmarshal(rsc.Value, m)
			if err != nil {
				log.Warn("Failed to unmarshal mesh config", err)
			}
			a.Mesh = m
			if a.LocalCacheDir != "" {
				// TODO: use jsonpb
				strResponse, err := json.MarshalIndent(m, "  ", "  ")
				if err != nil {
					continue
				}
				err = ioutil.WriteFile(a.LocalCacheDir+"_mesh.json", strResponse, 0644)
				if err != nil {
					continue
				}
			}
			continue
		}

		// Process the resources.
		a.VersionInfo[msg.TypeUrl] = msg.VersionInfo
		switch msg.TypeUrl {
		default:
			a.handleMCP(gvk, msg.Resources)
		}

		// If we got no resource - still save to the store with empty name/namespace, to notify sync
		// This scheme also allows us to chunk large responses !

		// TODO: add hook to inject nacks

		a.mutex.Lock()
		if len(gvk) == 3 {
			gt := config.GroupVersionKind{Group: gvk[0], Version: gvk[1], Kind: gvk[2]}
			if _, exist := a.sync[gt.String()]; !exist {
				a.sync[gt.String()] = time.Now()
				a.syncCh <- gt.String()
			}
		}
		a.Received[msg.TypeUrl] = msg
		a.ack(msg)
		a.mutex.Unlock()

		select {
		case a.XDSUpdates <- msg:
		default:
		}
	}
}

func (a *ADSC) WaitClear() {
	for {
		select {
		case <-a.Updates:
		default:
			return
		}
	}
}

func (a *ADSC) ack(msg *discovery.DiscoveryResponse) {
	var resources []string
	if msg.TypeUrl == v3.EndpointType {
		for c := range a.edsClusters {
			resources = append(resources, c)
		}
	}
	if msg.TypeUrl == v3.RouteType {
		for r := range a.routes {
			resources = append(resources, r)
		}
	}

	_ = a.stream.Send(&discovery.DiscoveryRequest{
		ResponseNonce: msg.Nonce,
		TypeUrl:       msg.TypeUrl,
		Node:          a.node(),
		VersionInfo:   msg.VersionInfo,
		ResourceNames: resources,
	})
}

func (a *ADSC) handleMCP(gvk []string, resources []*any.Any) {
	if len(gvk) != 3 || a.Store == nil {
		return // Not MCP Generic - fill up the store
	}

	groupVersionKind := config.GroupVersionKind{Group: gvk[0], Version: gvk[1], Kind: gvk[2]}
	for _, rsc := range resources {
		m := &mcp.Resource{}
		err := types.UnmarshalAny(&types.Any{
			TypeUrl: rsc.TypeUrl,
			Value:   rsc.Value,
		}, m)
		if err != nil {
			log.Error("Error unmarshalling received MCP config %v", err.Error())
			continue
		}
		xVersion, ok := a.ServiceEntry[m.Metadata.Name]
		if ok {
			if xVersion == m.Metadata.Version {
				log.Info("service entry xVersion is not change")
				continue
			}
			a.ServiceEntry[m.Metadata.Name] = m.Metadata.Version
		} else {
			a.ServiceEntry[m.Metadata.Name] = m.Metadata.Version
		}
		val, err := mcpToPilot(m)
		if err != nil {
			log.Error("Invalid data ", err.Error(), " ", string(rsc.Value))
			continue
		}
		//received[val.Namespace+"/"+val.Name] = val

		val.GroupVersionKind = groupVersionKind
		serviceEntry, err := getServiceEntry(val)
		if err != nil {
			continue
		}
		if len(serviceEntry.Spec.Endpoints) == 0 {
			a.DeleteServiceEntryChan <- serviceEntry
			delete(a.ServiceEntry, m.Metadata.Name)
		} else {
			a.CreateOrUpdateServiceEntryChan <- serviceEntry
		}

	}
}

func (a *ADSC) handleServiceEntry() {
	for {
		select {
		case serviceEntry := <-a.CreateOrUpdateServiceEntryChan:
			err := a.IstioClient.CreateOrUpdateServiceEntry(serviceEntry)
			if err != nil {
				log.Error("create service entry err %v", err.Error())
			}
		case serviceEntry := <-a.DeleteServiceEntryChan:
			err := a.IstioClient.DeleteServiceEntry(serviceEntry)
			if err != nil {
				log.Error("delete service entry err %v", err.Error())
			}
		}

	}
}

func (a *ADSC) node() *core.Node {
	n := &core.Node{
		Id:       a.nodeID,
		Locality: a.Locality,
	}
	if a.Metadata == nil {
		n.Metadata = &pstruct.Struct{
			Fields: map[string]*pstruct.Value{
				"ISTIO_VERSION": {Kind: &pstruct.Value_StringValue{StringValue: "65536.65536.65536"}},
			}}
	} else {
		n.Metadata = a.Metadata
		if a.Metadata.Fields["ISTIO_VERSION"] == nil {
			a.Metadata.Fields["ISTIO_VERSION"] = &pstruct.Value{Kind: &pstruct.Value_StringValue{StringValue: "65536.65536.65536"}}
		}
	}
	return n
}

func (a *ADSC) reconnect() {
	a.mutex.RLock()
	if a.closed {
		a.mutex.RUnlock()
		return
	}
	a.mutex.RUnlock()
	var err error
	a.client = discovery.NewAggregatedDiscoveryServiceClient(a.conn)
	a.stream, err = a.client.StreamAggregatedResources(context.Background())
	if err != nil {
		log.Error("can not get stream err is %s", err.Error())
		a.cfg.BackoffPolicy.Reset()
		time.AfterFunc(a.cfg.BackoffPolicy.NextBackOff(), a.reconnect)
	} else {
		a.Run(context.Background())
	}
	log.Info("new stream success")
}

// Close the stream.
func (a *ADSC) Close() {
	a.mutex.Lock()
	_ = a.conn.Close()
	a.closed = true
	close(a.CreateOrUpdateServiceEntryChan)
	//close(a.UpdateServiceEntryChan)
	//close(a.CreateServiceEntryChan)
	a.mutex.Unlock()
}

func mcpToPilot(m *mcp.Resource) (*config.Config, error) {
	if m == nil || m.Metadata == nil {
		return &config.Config{}, nil
	}
	c := &config.Config{
		Meta: config.Meta{
			ResourceVersion: m.Metadata.Version,
			Labels:          m.Metadata.Labels,
			Annotations:     m.Metadata.Annotations,
		},
	}
	nsn := strings.Split(m.Metadata.Name, "/")
	if len(nsn) != 2 {
		return nil, fmt.Errorf("invalid name %s", m.Metadata.Name)
	}
	c.Namespace = nsn[0]
	c.Name = nsn[1]
	var err error
	c.CreationTimestamp, err = types.TimestampFromProto(m.Metadata.CreateTime)
	if err != nil {
		return nil, err
	}

	pb, err := types.EmptyAny(m.Body)
	if err != nil {
		return nil, err
	}
	err = types.UnmarshalAny(m.Body, pb)
	if err != nil {
		return nil, err
	}
	c.Spec = pb
	return c, nil
}

func getServiceEntry(val *config.Config) (*v1alpha3.ServiceEntry, error) {
	serviceEntry := v1alpha3.ServiceEntry{}
	serviceEntry.Name = strings.ToLower(val.Name)
	serviceEntry.Namespace = val.Namespace
	serviceEntry.Annotations = val.Annotations
	serviceEntry.Kind = "ServiceEntry"
	serviceEntry.APIVersion = "networking.istio.io/v1alpha3"
	//serviceEntry.ResourceVersion = val.ResourceVersion
	r, err := config.ToJSON(val.Spec)
	if err != nil {
		log.Error("Error config to json %v", err.Error())
		return nil, err
	}
	spec := networkingv1alpha3.ServiceEntry{}
	err = json.Unmarshal(r, &spec)
	if err != nil {
		log.Error("Unmarshal config to json %v", err.Error())
		return nil, err
	}
	serviceEntry.Spec = spec
	var hosts []string
	for _, host := range serviceEntry.Spec.Hosts {
		hosts = append(hosts, strings.ToLower(host))
	}
	serviceEntry.Spec.Hosts = hosts
	for index, endpoint := range serviceEntry.Spec.Endpoints {
		label := make(map[string]string)
		if endpoint.Labels != nil {
			for key, value := range endpoint.Labels {
				if key == "app" || key == "version" {
					label[key] = value
				}
			}
			serviceEntry.Spec.Endpoints[index].Labels = label
		}
	}
	return &serviceEntry, nil
}

func getPrivateIPIfAvailable() net.IP {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		default:
			continue
		}
		if !ip.IsLoopback() {
			return ip
		}
	}
	return net.IPv4zero
}
