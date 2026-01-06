package utils

import (
	"fmt"
	"os"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"istio.io/istio/pkg/kube"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

func FromEnvOr(key, defaultVal string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	return val
}

func Deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func GetKubeClientOptions() (client.Options, error) {
	opt := client.Options{}
	// Create a new scheme and add both Istio and Gateway API types
	scheme := runtime.NewScheme()

	// Add core Kubernetes types first
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return opt, fmt.Errorf("failed to add core Kubernetes types to scheme: %v", err)
	}

	// Add both Istio types and Gateway API types by copying from the existing Istio scheme
	// Add each known type from IstioScheme to our new scheme
	for gvk, objType := range kube.IstioScheme.AllKnownTypes() {
		// Create a new instance of the type to get a runtime.Object
		obj := reflect.New(objType).Interface()
		if runtimeObj, ok := obj.(runtime.Object); ok {
			scheme.AddKnownTypeWithName(gvk, runtimeObj)
		} else {
			return opt, fmt.Errorf("type %v is not a runtime.Object", gvk)
		}
	}

	// Add Gateway API types
	if err := v1.Install(scheme); err != nil {
		return opt, fmt.Errorf("failed to add Gateway API v1 types to scheme: %v", err)
	}
	opt.Scheme = scheme
	return opt, nil
}
