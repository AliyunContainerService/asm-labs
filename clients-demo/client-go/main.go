package main

import (
	"context"
	"os"

	"github.com/ghodss/yaml"
	asmv1 "istio.io/api/alibabacloudservicemesh/v1"
	asmversionedclient "istio.io/client-go/asm/pkg/clientset"
	networkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	istioversionedclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func main() {
	kubeconfigPath := os.Getenv("HOME") + "/.kube/config"
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		klog.Errorf("failed to create rest config, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully create rest config")

	clientset, err := asmversionedclient.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("failed to create clientset, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully create asm clientset")

	klog.Info("run cluster scope clients demo")
	runClusterScopeDemo(clientset)

	klog.Info("run namespaced scope clients demo")
	runNamespacedScopeDemo(clientset)

	istioclientset, err := istioversionedclient.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("failed to create istio clientset, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully create istio clientset")

	klog.Info("run istio resource demo")
	runIstioResourceDemo(istioclientset)
}

// runClusterScopeDemo runs a demo for asm cluster scope resources
func runClusterScopeDemo(clientset *asmversionedclient.Clientset) {
	asmswimlanegroup := &asmv1.ASMSwimLaneGroup{}
	err := yaml.Unmarshal([]byte(swimlanegroupYaml), &asmswimlanegroup)
	if err != nil {
		klog.Errorf("failed to unmarshal yaml bytes, err: %+v", err)
		os.Exit(1)
	}

	_, err = clientset.IstioV1().ASMSwimLaneGroups().Create(context.TODO(), asmswimlanegroup, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("failed to create asmswimlanegroup, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully create asmswimlanegroup")

	_, err = clientset.IstioV1().ASMSwimLaneGroups().Get(context.TODO(), asmswimlanegroup.Name, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("failed to get asmswimlanegroup, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully get asmswimlanegroup")

	_, err = clientset.IstioV1().ASMSwimLaneGroups().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("failed to list asmswimlanegroups, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully list asmswimlanegroups")

	err = clientset.IstioV1().ASMSwimLaneGroups().Delete(context.TODO(), asmswimlanegroup.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("failed to delete asmswimlanegroup, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully delete asmswimlanegroup")
}

// runNamespacedScopeDemo runs a demo for asm namespaced scope resources
func runNamespacedScopeDemo(clientset *asmversionedclient.Clientset) {
	asmlocalratelimiter := &asmv1.ASMLocalRateLimiter{}
	err := yaml.Unmarshal([]byte(localratelimiterYaml), &asmlocalratelimiter)
	if err != nil {
		klog.Errorf("failed to unmarshal yaml bytes, err: %+v", err)
		os.Exit(1)
	}

	_, err = clientset.IstioV1().ASMLocalRateLimiters(asmlocalratelimiter.Namespace).Create(context.TODO(), asmlocalratelimiter, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("failed to create asmlocalratelimiter, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully create asmlocalratelimiter")

	_, err = clientset.IstioV1().ASMLocalRateLimiters(asmlocalratelimiter.Namespace).Get(context.TODO(), asmlocalratelimiter.Name, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("failed to get asmlocalratelimiter, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully get asmlocalratelimiter")

	_, err = clientset.IstioV1().ASMLocalRateLimiters(asmlocalratelimiter.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("failed to list asmlocalratelimiters, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully list asmlocalratelimiters")

	err = clientset.IstioV1().ASMLocalRateLimiters(asmlocalratelimiter.Namespace).Delete(context.TODO(), asmlocalratelimiter.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("failed to delete asmlocalratelimiter, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully delete asmlocalratelimiter")
}

// runIstioResourceDemo runs a demo for istio resources
func runIstioResourceDemo(clientset *istioversionedclient.Clientset) {
	vs := &networkingv1.VirtualService{}
	err := yaml.Unmarshal([]byte(virtualserviceYaml), &vs)
	if err != nil {
		klog.Errorf("failed to unmarshal yaml bytes, err: %+v", err)
		os.Exit(1)
	}

	_, err = clientset.NetworkingV1().VirtualServices(vs.Namespace).Create(context.TODO(), vs, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("failed to create virtualservice, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully create virtualservice")

	_, err = clientset.NetworkingV1().VirtualServices(vs.Namespace).Get(context.TODO(), vs.Name, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("failed to get virtualservice, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully get virtualservice")

	_, err = clientset.NetworkingV1().VirtualServices(vs.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("failed to list virtualservices, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully list virtualservices")

	err = clientset.NetworkingV1().VirtualServices(vs.Namespace).Delete(context.TODO(), vs.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("failed to delete virtualservice, err: %+v", err)
		os.Exit(1)
	}
	klog.Info("successfully delete virtualservice")
}

const (
	swimlanegroupYaml = `apiVersion: istio.alibabacloud.com/v1
kind: ASMSwimLaneGroup
metadata:
  name: demo
spec:
  ingress:
    gateway:
      name: ingressgateway
      namespace: istio-system
      type: ASM
  services:
  - name: mocka
    namespace: default
  - name: mockb
    namespace: default
  - name: mockc
    namespace: default`

	localratelimiterYaml = `apiVersion: istio.alibabacloud.com/v1beta1
kind: ASMLocalRateLimiter
metadata:
  name: demo
  namespace: default
spec:
  configs:
    - match:
        vhost:
          name: "www.example2.com"
          port: 80
          route:
            name_match: "test1"
      limit:
         fill_interval:
            seconds: 1
         quota: 100`

	virtualserviceYaml = `apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: demo
  namespace: default
spec:
  hosts:
  - "*"
  gateways:
  - istio-system/ingressgateway
  http:
  - match:
    route:
    - destination:
        host: demo
        port:
          number: 9080`
)
