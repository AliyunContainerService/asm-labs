package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/common"
	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/consul"
	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/control"
	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/nacos"
	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/provider"
	"gitlab.alibaba-inc.com/cos/asm-se-syncer/pkg/serviceentry"
	"istio.io/api/networking/v1alpha3"
	ic "istio.io/client-go/pkg/clientset/versioned"
	icinformer "istio.io/client-go/pkg/informers/externalversions/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/config/memory"
	"istio.io/istio/pkg/config/schema/collections"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"os"
	"time"
)

const (
	apiGroup      = "networking.istio.io"
	apiVersion    = "v1alpha3"
	apiType       = apiGroup + "/" + apiVersion
	kind          = "ServiceEntry"
	allNamespaces = ""
	resyncPeriod  = 30
)

var (
	debug           bool
	kubeConfig      string
	namespace       string
	consulEndpoint  string
	consulNamespace string
	prefix          string
	meshId          string
	regionId        string
	accessKeyId     string
	accessKeySecret string
)

func serve() (serve *cobra.Command) {

	serve = &cobra.Command{
		Use:     "serve",
		Aliases: []string{"serve"},
		Short:   "Starts the ASM ServiceEntry Syncer server",
		Example: "asm-se-syncer serve",
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg *restclient.Config
			var serviceRegistryConfigList []map[string]interface{}

			if meshId == "" {
				return errors.New("meshId is required")
			}
			serviceRegistryConfigList, err := common.GetServiceRegistryConfig()
			if err != nil || len(serviceRegistryConfigList) == 0 {
				return errors.Wrap(err, "failed to get service registry config")
			}
			akId, err := common.GetAccessKeyID()
			if err != nil {
				return err
			}
			akSecret, err := common.GetAccessKeySecret()
			if err != nil {
				return err
			}
			accessKeyId = string(akId)
			accessKeySecret = string(akSecret)
			cfg, err = common.GetASMRestConfig(meshId, regionId, accessKeyId, accessKeySecret)
			if err != nil {
				return errors.Wrap(err, "failed to get asm rest config")
			}

			//viper.SetConfigName("config") // name of config file (without extension)
			//viper.AddConfigPath(serviceRegistryConfigDir)   // path to look for the config file in
			//err = viper.ReadInConfig() // Find and read the config file
			//if err != nil { // Handle errors reading the config file
			//	log.Errorf("Fatal error config file: %+v \n", err)
			//	return errors.Wrapf(err, "Fatal error config file")
			//}
			//viper.WatchConfig()
			//viper.OnConfigChange(func(e fsnotify.Event) {
			//	log.Infof("Config file changed: %s", e.Name)
			//	if err := viper.Unmarshal(&serviceRegistryConfig); err != nil {
			//		log.Errorf("error Unmarshal config file %q\n", serviceRegistryConfigFilePath)
			//	}
			//})

			ic, err := ic.NewForConfig(cfg)
			if err != nil {
				return errors.Wrap(err, "failed to create an istio client from the k8s rest config")
			}

			ctx := context.Background() // common context for cancellation across all loops/routines

			if len(namespace) == 0 {
				if ns, set := os.LookupEnv("PUBLISH_NAMESPACE"); set {
					namespace = ns
				}
			}

			// only support one Nacos service registry or multi consul service registries for now
			watchers, err := getWatcher(serviceRegistryConfigList, cfg)
			if err != nil {
				return err
			}
			//check if has two kind  service registry
			hasMulti := hasMultiKindServiceRegistry(watchers)
			if hasMulti {
				return errors.New("Multiple kind service registries is not supported.")
			}
			for _, watcher := range watchers {
				if watcher.WatcherType() == string(common.Nacos) {
					go watcher.Run(ctx)
					//nacos only support one cluster
					break
				} else {
					//Consul
					go watcher.Run(ctx)
				}
			}
			for _, watcher := range watchers {
				// we get the service entry for namespace `namespace` for the synchronizer to publish service entries in to
				// (if we use an `allNamespaces` client here we can't publish). Listening for ServiceEntries is done with
				// the informer, which uses allNamespace.
				toNamespace := findNamespace(watcher.ToNamespace())
				err = populateNamespace(cfg, toNamespace)
				if err != nil {
					return err
				}
				serviceRegistryType := watcher.WatcherType()
				if serviceRegistryType == string(common.Consul) {
					istio := serviceentry.New()
					if debug {
						istio = serviceentry.NewLoggingStore(istio, log.Infof)
					}
					log.Info("Starting Synchronizer control loop, prefix %s", watcher.Prefix())
					write := ic.NetworkingV1alpha3().ServiceEntries(toNamespace)
					location := v1alpha3.ServiceEntry_MESH_EXTERNAL
					interval := time.Second * 5
					sync := control.NewSynchronizer(namespace, istio, watcher.Cache(), watcher.Prefix(), location, interval, write)
					go sync.Run(ctx)

					informer := icinformer.NewServiceEntryInformer(ic, allNamespaces, 5*time.Second,
						cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
					serviceentry.AttachHandler(istio, informer)
					log.Infof("Watching %s.%s across all namespaces with resync period %d and %q", apiType, kind, resyncPeriod)
					go informer.Run(ctx.Done())
				}
			}

			<-ctx.Done()
			return nil
		},
	}

	serve.PersistentFlags().BoolVar(&debug, "debug", true, "if true, enables more logging")
	serve.PersistentFlags().StringVar(&meshId,
		"meshId", "", "the id of asm instance")
	serve.PersistentFlags().StringVar(&regionId,
		"regionId", "cn-hangzhou", "the id of the region where the asm instance is running")
	serve.PersistentFlags().StringVar(&accessKeyId,
		"accessKeyId", "xxx", "user accessKeyId")
	serve.PersistentFlags().StringVar(&accessKeySecret,
		"accessKeySecret", "xxx", "user accessKeySecret")
	serve.PersistentFlags().StringVar(&kubeConfig,
		"kubeconfig", "", "kubeconfig location; if empty the server will assume it's in a cluster; for local testing use ~/.kube/config")
	return serve
}

func getWatcher(serviceRegistryConfigList []map[string]interface{}, cfg *restclient.Config) ([]provider.Watcher, error) {
	if serviceRegistryConfigList == nil || len(serviceRegistryConfigList) == 0 {
		return nil, errors.New("failed to initialize watchers as serviceRegistryConfigList is empty")
	}
	log.Info("Initializing Watchers")
	var watchers []provider.Watcher
	for _, serviceRegistryInfo := range serviceRegistryConfigList {
		var watcher provider.Watcher
		serviceRegistryType := cast.ToString(serviceRegistryInfo["type"])
		if serviceRegistryType == string(common.Consul) {
			store := provider.NewCache()
			consulEndpoint := cast.ToString(serviceRegistryInfo["endpoint"])
			consulNamespace := cast.ToString(serviceRegistryInfo["consulNamespace"])
			prefix := cast.ToString(serviceRegistryInfo["prefix"])
			toNamespace := cast.ToString(serviceRegistryInfo["toNamespace"])
			consulWatcher, consulErr := consul.NewWatcher(store, consulEndpoint, consulNamespace, prefix, toNamespace)
			if consulErr != nil {
				log.Errorf("error setting up consul: %v", consulErr)
				continue
			} else {
				log.Infof("Consul Watcher initialized at %s", consulEndpoint)
				watcher = consulWatcher
			}
			watchers = append(watchers, watcher)
		} else if serviceRegistryType == string(common.Nacos) {
			nacosEndpoint := cast.ToString(serviceRegistryInfo["endpoint"])
			//nacosNamespace := cast.ToString(serviceRegistryInfo["nacosNamespace"])
			toNamespace := cast.ToString(serviceRegistryInfo["toNamespace"])
			store := memory.Make(collections.Pilot)
			configController := memory.NewController(store)
			log.Info("create configController success")
			nacosWatcher, nacosErr := nacos.NewWatcher(nacosEndpoint, nil, cfg, "", toNamespace, configController)
			if nacosErr != nil {
				log.Errorf("error setting up nacos: %v", nacosErr)
				return nil, errors.Wrapf(nacosErr, "failed to initialize nacos watchers")
			} else {
				log.Infof("nacos Watcher initialized at %s", consulEndpoint)
				watcher = nacosWatcher
			}
			watchers = append(watchers, watcher)
		} else {
			log.Errorf("the service registry type is not supported: %s", serviceRegistryType)
			return nil, errors.Errorf("the service registry type is not supported: %s", serviceRegistryType)
		}
	}
	log.Infof("watchers nums %d", len(watchers))

	return watchers, nil

}

func main() {
	root := &cobra.Command{
		Short:   "asm-se-syncer",
		Example: "",
	}
	root.AddCommand(serve())
	if err := root.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func findNamespace(namespace string) string {
	if len(namespace) > 0 {
		log.Infof("using namespace flag to publish service entries into %q", namespace)
		return namespace
	}
	// This way assumes you've set the POD_NAMESPACE environment variable using the downward API.
	// This check has to be done first for backwards compatibility with the way InClusterConfig was originally set up
	if ns, ok := os.LookupEnv("POD_NAMESPACE"); ok {
		log.Infof("using POD_NAMESPACE environment variable to publish service entries into %q", namespace)
		return ns
	}

	log.Infof("couldn't determine a namespace, falling back to %q", "default")
	return "default"
}

func populateNamespace(cfg *restclient.Config, namespace string) error {
	// check if namespace exists, create it if not
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}
	ns, err := clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		log.Info(fmt.Sprintf("checking if namespace %s exists with error - requeue", namespace), "error", err)
		return err
	}

	if k8serrors.IsNotFound(err) {
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		ns, err = clientset.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
		if err != nil {
			log.Info(fmt.Sprintf("namespace %s cannot be created with error - requeue", namespace), "error", err)
			return err
		}
		log.Info(fmt.Sprintf("namespace %s has been created", namespace))
		return nil
	} else {
		log.Info(fmt.Sprintf("namespace %s already exists", namespace))
	}

	return nil
}
func getASMServiceRegistry(cfg *restclient.Config) (types.UID, error) {
	crdConfig := cfg
	crdConfig.GroupVersion = &schema.GroupVersion{Group: "istio.alibabacloud.com", Version: "v1beta1"}
	crdConfig.APIPath = "/apis"
	crdConfig.ContentType = k8sruntime.ContentTypeJSON
	crdConfig.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	if crdConfig.UserAgent == "" {
		crdConfig.UserAgent = restclient.DefaultKubernetesUserAgent()
	}
	restClient, err := restclient.RESTClientFor(crdConfig)
	if err != nil {
		return "", err
	}
	var defaultSR unstructured.Unstructured
	err = restClient.Get().Resource("asmserviceregistrys").Name("default").Do(context.TODO()).Into(&defaultSR)
	if err != nil {
		return "", err
	}
	log.Infof(fmt.Sprintf("asmserviceregistrys:%+v", defaultSR))

	return defaultSR.GetUID(), nil
}

func hasMultiKindServiceRegistry(watchers []provider.Watcher) bool {
	previousType := ""
	if watchers == nil || len(watchers) <= 1 {
		return false
	}
	for i, watcher := range watchers {
		if i == 0 {
			previousType = watcher.WatcherType()
		} else {
			if previousType != watcher.WatcherType() {
				return true
			}
			previousType = watcher.WatcherType()
		}
	}
	return false
}
