package nacos

import (
	"context"
	log "github.com/sirupsen/logrus"
	metaV3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	metaV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IstioClient struct {
	Config         *rest.Config
	Client         *versionedclient.Clientset
	IstioK8sClient *kubernetes.Clientset
}

func NewClient(config *rest.Config) (*IstioClient, error) {
	ic, err := versionedclient.NewForConfig(config)
	if err != nil {
		log.Error(err, "can not new istio client")
		return nil, err
	}
	istioK8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &IstioClient{
		Config:         config,
		Client:         ic,
		IstioK8sClient: istioK8sClient,
	}, nil
}

func (k *IstioClient) CreateNamespace(namespace string) error {
	_, err := k.IstioK8sClient.CoreV1().Namespaces().Get(context.Background(), namespace, v1.GetOptions{})
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		ns := &metaV1.Namespace{}
		ns.Namespace = namespace
		ns.Name = namespace
		_, err := k.IstioK8sClient.CoreV1().Namespaces().Create(context.Background(), ns, v1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *IstioClient) CreateOrUpdateServiceEntry(serviceEntry *v1alpha3.ServiceEntry) error {
	err := k.CreateNamespace(serviceEntry.Namespace)
	if err != nil {
		log.Error(err, "create namespace error")
		return err
	}
	existServiceEntry, err := k.Client.NetworkingV1alpha3().ServiceEntries(serviceEntry.Namespace).Get(context.Background(), serviceEntry.Name, v1.GetOptions{})
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error("get service entry err %v", err.Error())
			return err
		}
		_, err = k.Client.NetworkingV1alpha3().ServiceEntries(serviceEntry.Namespace).Create(context.Background(), serviceEntry, v1.CreateOptions{})
		if err != nil {
			log.Error("create service entry err %v", err.Error())
			return err
		}

		return nil
	}
	if existServiceEntry.Annotations != nil {
		_, ok := existServiceEntry.Annotations["update"]
		if ok {
			existServiceEntry.Annotations["update"] = "nacos-mesh"
			var endpoints []*metaV3.WorkloadEntry
			for _, newEndpoint := range serviceEntry.Spec.Endpoints {
				isHave := false
				for _, endpoint := range existServiceEntry.Spec.Endpoints {
					if endpoint.Weight != 1 && endpoint.Address == newEndpoint.Address {
						isHave = true
					}
				}
				if !isHave {
					endpoints = append(endpoints, newEndpoint)
				}
			}
			serviceEntry.Spec.Endpoints = endpoints
		}
	}
	if reflect.DeepEqual(existServiceEntry.Spec, serviceEntry.Spec) {
		log.Info("service entry is not change")
		return nil
	}
	existServiceEntry.Spec = serviceEntry.Spec
	log.Info("service entry is ", serviceEntry)
	_, err = k.Client.NetworkingV1alpha3().ServiceEntries(serviceEntry.Namespace).Update(context.Background(), existServiceEntry, v1.UpdateOptions{})
	if err != nil {
		log.Error("update service entry err %v", err.Error())
		return err
	}
	return nil
}

func (k *IstioClient) DeleteServiceEntry(serviceEntry *v1alpha3.ServiceEntry) error {
	existServiceEntry, err := k.Client.NetworkingV1alpha3().ServiceEntries(serviceEntry.Namespace).Get(context.Background(), serviceEntry.Name, v1.GetOptions{})
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}

		return nil
	}

	err = k.Client.NetworkingV1alpha3().ServiceEntries(existServiceEntry.Namespace).Delete(context.Background(), existServiceEntry.Name, v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}
