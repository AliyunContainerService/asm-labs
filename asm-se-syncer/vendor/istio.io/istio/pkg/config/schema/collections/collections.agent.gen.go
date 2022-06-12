// +build agent
// GENERATED FILE -- DO NOT EDIT
//

package collections

import (
	"reflect"

	istioioapimeshv1alpha1 "istio.io/api/mesh/v1alpha1"
	istioioapimetav1alpha1 "istio.io/api/meta/v1alpha1"
	istioioapinetworkingv1alpha3 "istio.io/api/networking/v1alpha3"
	istioioapisecurityv1beta1 "istio.io/api/security/v1beta1"
	"istio.io/istio/pkg/config/schema/collection"
	"istio.io/istio/pkg/config/schema/resource"
	"istio.io/istio/pkg/config/validation"
)

var (

	// IstioMeshV1Alpha1MeshConfig describes the collection
	// istio/mesh/v1alpha1/MeshConfig
	IstioMeshV1Alpha1MeshConfig = collection.Builder{
		Name:         "istio/mesh/v1alpha1/MeshConfig",
		VariableName: "IstioMeshV1Alpha1MeshConfig",
		Disabled:     false,
		Resource: resource.Builder{
			Group:         "",
			Kind:          "MeshConfig",
			Plural:        "meshconfigs",
			Version:       "v1alpha1",
			Proto:         "istio.mesh.v1alpha1.MeshConfig",
			ReflectType:   reflect.TypeOf(&istioioapimeshv1alpha1.MeshConfig{}).Elem(),
			ProtoPackage:  "istio.io/api/mesh/v1alpha1",
			ClusterScoped: false,
			ValidateProto: validation.EmptyValidate,
		}.MustBuild(),
	}.MustBuild()

	// IstioMeshV1Alpha1MeshNetworks describes the collection
	// istio/mesh/v1alpha1/MeshNetworks
	IstioMeshV1Alpha1MeshNetworks = collection.Builder{
		Name:         "istio/mesh/v1alpha1/MeshNetworks",
		VariableName: "IstioMeshV1Alpha1MeshNetworks",
		Disabled:     false,
		Resource: resource.Builder{
			Group:         "",
			Kind:          "MeshNetworks",
			Plural:        "meshnetworks",
			Version:       "v1alpha1",
			Proto:         "istio.mesh.v1alpha1.MeshNetworks",
			ReflectType:   reflect.TypeOf(&istioioapimeshv1alpha1.MeshNetworks{}).Elem(),
			ProtoPackage:  "istio.io/api/mesh/v1alpha1",
			ClusterScoped: false,
			ValidateProto: validation.EmptyValidate,
		}.MustBuild(),
	}.MustBuild()

	// IstioNetworkingV1Alpha3Destinationrules describes the collection
	// istio/networking/v1alpha3/destinationrules
	IstioNetworkingV1Alpha3Destinationrules = collection.Builder{
		Name:         "istio/networking/v1alpha3/destinationrules",
		VariableName: "IstioNetworkingV1Alpha3Destinationrules",
		Disabled:     false,
		Resource: resource.Builder{
			Group:   "networking.istio.io",
			Kind:    "DestinationRule",
			Plural:  "destinationrules",
			Version: "v1alpha3",
			Proto:   "istio.networking.v1alpha3.DestinationRule", StatusProto: "istio.meta.v1alpha1.IstioStatus",
			ReflectType: reflect.TypeOf(&istioioapinetworkingv1alpha3.DestinationRule{}).Elem(), StatusType: reflect.TypeOf(&istioioapimetav1alpha1.IstioStatus{}).Elem(),
			ProtoPackage: "istio.io/api/networking/v1alpha3", StatusPackage: "istio.io/api/meta/v1alpha1",
			ClusterScoped: false,
			ValidateProto: validation.ValidateDestinationRule,
		}.MustBuild(),
	}.MustBuild()

	// IstioNetworkingV1Alpha3Envoyfilters describes the collection
	// istio/networking/v1alpha3/envoyfilters
	IstioNetworkingV1Alpha3Envoyfilters = collection.Builder{
		Name:         "istio/networking/v1alpha3/envoyfilters",
		VariableName: "IstioNetworkingV1Alpha3Envoyfilters",
		Disabled:     false,
		Resource: resource.Builder{
			Group:   "networking.istio.io",
			Kind:    "EnvoyFilter",
			Plural:  "envoyfilters",
			Version: "v1alpha3",
			Proto:   "istio.networking.v1alpha3.EnvoyFilter", StatusProto: "istio.meta.v1alpha1.IstioStatus",
			ReflectType: reflect.TypeOf(&istioioapinetworkingv1alpha3.EnvoyFilter{}).Elem(), StatusType: reflect.TypeOf(&istioioapimetav1alpha1.IstioStatus{}).Elem(),
			ProtoPackage: "istio.io/api/networking/v1alpha3", StatusPackage: "istio.io/api/meta/v1alpha1",
			ClusterScoped: false,
			ValidateProto: validation.ValidateEnvoyFilter,
		}.MustBuild(),
	}.MustBuild()

	// IstioNetworkingV1Alpha3Gateways describes the collection
	// istio/networking/v1alpha3/gateways
	IstioNetworkingV1Alpha3Gateways = collection.Builder{
		Name:         "istio/networking/v1alpha3/gateways",
		VariableName: "IstioNetworkingV1Alpha3Gateways",
		Disabled:     false,
		Resource: resource.Builder{
			Group:   "networking.istio.io",
			Kind:    "Gateway",
			Plural:  "gateways",
			Version: "v1alpha3",
			Proto:   "istio.networking.v1alpha3.Gateway", StatusProto: "istio.meta.v1alpha1.IstioStatus",
			ReflectType: reflect.TypeOf(&istioioapinetworkingv1alpha3.Gateway{}).Elem(), StatusType: reflect.TypeOf(&istioioapimetav1alpha1.IstioStatus{}).Elem(),
			ProtoPackage: "istio.io/api/networking/v1alpha3", StatusPackage: "istio.io/api/meta/v1alpha1",
			ClusterScoped: false,
			ValidateProto: validation.ValidateGateway,
		}.MustBuild(),
	}.MustBuild()

	// IstioNetworkingV1Alpha3Serviceentries describes the collection
	// istio/networking/v1alpha3/serviceentries
	IstioNetworkingV1Alpha3Serviceentries = collection.Builder{
		Name:         "istio/networking/v1alpha3/serviceentries",
		VariableName: "IstioNetworkingV1Alpha3Serviceentries",
		Disabled:     false,
		Resource: resource.Builder{
			Group:   "networking.istio.io",
			Kind:    "ServiceEntry",
			Plural:  "serviceentries",
			Version: "v1alpha3",
			Proto:   "istio.networking.v1alpha3.ServiceEntry", StatusProto: "istio.meta.v1alpha1.IstioStatus",
			ReflectType: reflect.TypeOf(&istioioapinetworkingv1alpha3.ServiceEntry{}).Elem(), StatusType: reflect.TypeOf(&istioioapimetav1alpha1.IstioStatus{}).Elem(),
			ProtoPackage: "istio.io/api/networking/v1alpha3", StatusPackage: "istio.io/api/meta/v1alpha1",
			ClusterScoped: false,
			ValidateProto: validation.ValidateServiceEntry,
		}.MustBuild(),
	}.MustBuild()

	// IstioNetworkingV1Alpha3Sidecars describes the collection
	// istio/networking/v1alpha3/sidecars
	IstioNetworkingV1Alpha3Sidecars = collection.Builder{
		Name:         "istio/networking/v1alpha3/sidecars",
		VariableName: "IstioNetworkingV1Alpha3Sidecars",
		Disabled:     false,
		Resource: resource.Builder{
			Group:   "networking.istio.io",
			Kind:    "Sidecar",
			Plural:  "sidecars",
			Version: "v1alpha3",
			Proto:   "istio.networking.v1alpha3.Sidecar", StatusProto: "istio.meta.v1alpha1.IstioStatus",
			ReflectType: reflect.TypeOf(&istioioapinetworkingv1alpha3.Sidecar{}).Elem(), StatusType: reflect.TypeOf(&istioioapimetav1alpha1.IstioStatus{}).Elem(),
			ProtoPackage: "istio.io/api/networking/v1alpha3", StatusPackage: "istio.io/api/meta/v1alpha1",
			ClusterScoped: false,
			ValidateProto: validation.ValidateSidecar,
		}.MustBuild(),
	}.MustBuild()

	// IstioNetworkingV1Alpha3Virtualservices describes the collection
	// istio/networking/v1alpha3/virtualservices
	IstioNetworkingV1Alpha3Virtualservices = collection.Builder{
		Name:         "istio/networking/v1alpha3/virtualservices",
		VariableName: "IstioNetworkingV1Alpha3Virtualservices",
		Disabled:     false,
		Resource: resource.Builder{
			Group:   "networking.istio.io",
			Kind:    "VirtualService",
			Plural:  "virtualservices",
			Version: "v1alpha3",
			Proto:   "istio.networking.v1alpha3.VirtualService", StatusProto: "istio.meta.v1alpha1.IstioStatus",
			ReflectType: reflect.TypeOf(&istioioapinetworkingv1alpha3.VirtualService{}).Elem(), StatusType: reflect.TypeOf(&istioioapimetav1alpha1.IstioStatus{}).Elem(),
			ProtoPackage: "istio.io/api/networking/v1alpha3", StatusPackage: "istio.io/api/meta/v1alpha1",
			ClusterScoped: false,
			ValidateProto: validation.ValidateVirtualService,
		}.MustBuild(),
	}.MustBuild()

	// IstioNetworkingV1Alpha3Workloadentries describes the collection
	// istio/networking/v1alpha3/workloadentries
	IstioNetworkingV1Alpha3Workloadentries = collection.Builder{
		Name:         "istio/networking/v1alpha3/workloadentries",
		VariableName: "IstioNetworkingV1Alpha3Workloadentries",
		Disabled:     false,
		Resource: resource.Builder{
			Group:   "networking.istio.io",
			Kind:    "WorkloadEntry",
			Plural:  "workloadentries",
			Version: "v1alpha3",
			Proto:   "istio.networking.v1alpha3.WorkloadEntry", StatusProto: "istio.meta.v1alpha1.IstioStatus",
			ReflectType: reflect.TypeOf(&istioioapinetworkingv1alpha3.WorkloadEntry{}).Elem(), StatusType: reflect.TypeOf(&istioioapimetav1alpha1.IstioStatus{}).Elem(),
			ProtoPackage: "istio.io/api/networking/v1alpha3", StatusPackage: "istio.io/api/meta/v1alpha1",
			ClusterScoped: false,
			ValidateProto: validation.ValidateWorkloadEntry,
		}.MustBuild(),
	}.MustBuild()

	// IstioNetworkingV1Alpha3Workloadgroups describes the collection
	// istio/networking/v1alpha3/workloadgroups
	IstioNetworkingV1Alpha3Workloadgroups = collection.Builder{
		Name:         "istio/networking/v1alpha3/workloadgroups",
		VariableName: "IstioNetworkingV1Alpha3Workloadgroups",
		Disabled:     false,
		Resource: resource.Builder{
			Group:   "networking.istio.io",
			Kind:    "WorkloadGroup",
			Plural:  "workloadgroups",
			Version: "v1alpha3",
			Proto:   "istio.networking.v1alpha3.WorkloadGroup", StatusProto: "istio.meta.v1alpha1.IstioStatus",
			ReflectType: reflect.TypeOf(&istioioapinetworkingv1alpha3.WorkloadGroup{}).Elem(), StatusType: reflect.TypeOf(&istioioapimetav1alpha1.IstioStatus{}).Elem(),
			ProtoPackage: "istio.io/api/networking/v1alpha3", StatusPackage: "istio.io/api/meta/v1alpha1",
			ClusterScoped: false,
			ValidateProto: validation.ValidateWorkloadGroup,
		}.MustBuild(),
	}.MustBuild()

	// IstioSecurityV1Beta1Authorizationpolicies describes the collection
	// istio/security/v1beta1/authorizationpolicies
	IstioSecurityV1Beta1Authorizationpolicies = collection.Builder{
		Name:         "istio/security/v1beta1/authorizationpolicies",
		VariableName: "IstioSecurityV1Beta1Authorizationpolicies",
		Disabled:     false,
		Resource: resource.Builder{
			Group:   "security.istio.io",
			Kind:    "AuthorizationPolicy",
			Plural:  "authorizationpolicies",
			Version: "v1beta1",
			Proto:   "istio.security.v1beta1.AuthorizationPolicy", StatusProto: "istio.meta.v1alpha1.IstioStatus",
			ReflectType: reflect.TypeOf(&istioioapisecurityv1beta1.AuthorizationPolicy{}).Elem(), StatusType: reflect.TypeOf(&istioioapimetav1alpha1.IstioStatus{}).Elem(),
			ProtoPackage: "istio.io/api/security/v1beta1", StatusPackage: "istio.io/api/meta/v1alpha1",
			ClusterScoped: false,
			ValidateProto: validation.ValidateAuthorizationPolicy,
		}.MustBuild(),
	}.MustBuild()

	// IstioSecurityV1Beta1Peerauthentications describes the collection
	// istio/security/v1beta1/peerauthentications
	IstioSecurityV1Beta1Peerauthentications = collection.Builder{
		Name:         "istio/security/v1beta1/peerauthentications",
		VariableName: "IstioSecurityV1Beta1Peerauthentications",
		Disabled:     false,
		Resource: resource.Builder{
			Group:   "security.istio.io",
			Kind:    "PeerAuthentication",
			Plural:  "peerauthentications",
			Version: "v1beta1",
			Proto:   "istio.security.v1beta1.PeerAuthentication", StatusProto: "istio.meta.v1alpha1.IstioStatus",
			ReflectType: reflect.TypeOf(&istioioapisecurityv1beta1.PeerAuthentication{}).Elem(), StatusType: reflect.TypeOf(&istioioapimetav1alpha1.IstioStatus{}).Elem(),
			ProtoPackage: "istio.io/api/security/v1beta1", StatusPackage: "istio.io/api/meta/v1alpha1",
			ClusterScoped: false,
			ValidateProto: validation.ValidatePeerAuthentication,
		}.MustBuild(),
	}.MustBuild()

	// IstioSecurityV1Beta1Requestauthentications describes the collection
	// istio/security/v1beta1/requestauthentications
	IstioSecurityV1Beta1Requestauthentications = collection.Builder{
		Name:         "istio/security/v1beta1/requestauthentications",
		VariableName: "IstioSecurityV1Beta1Requestauthentications",
		Disabled:     false,
		Resource: resource.Builder{
			Group:   "security.istio.io",
			Kind:    "RequestAuthentication",
			Plural:  "requestauthentications",
			Version: "v1beta1",
			Proto:   "istio.security.v1beta1.RequestAuthentication", StatusProto: "istio.meta.v1alpha1.IstioStatus",
			ReflectType: reflect.TypeOf(&istioioapisecurityv1beta1.RequestAuthentication{}).Elem(), StatusType: reflect.TypeOf(&istioioapimetav1alpha1.IstioStatus{}).Elem(),
			ProtoPackage: "istio.io/api/security/v1beta1", StatusPackage: "istio.io/api/meta/v1alpha1",
			ClusterScoped: false,
			ValidateProto: validation.ValidateRequestAuthentication,
		}.MustBuild(),
	}.MustBuild()

	// All contains all collections in the system.
	All = collection.NewSchemasBuilder().
		MustAdd(IstioMeshV1Alpha1MeshConfig).
		MustAdd(IstioMeshV1Alpha1MeshNetworks).
		MustAdd(IstioNetworkingV1Alpha3Destinationrules).
		MustAdd(IstioNetworkingV1Alpha3Envoyfilters).
		MustAdd(IstioNetworkingV1Alpha3Gateways).
		MustAdd(IstioNetworkingV1Alpha3Serviceentries).
		MustAdd(IstioNetworkingV1Alpha3Sidecars).
		MustAdd(IstioNetworkingV1Alpha3Virtualservices).
		MustAdd(IstioNetworkingV1Alpha3Workloadentries).
		MustAdd(IstioNetworkingV1Alpha3Workloadgroups).
		MustAdd(IstioSecurityV1Beta1Authorizationpolicies).
		MustAdd(IstioSecurityV1Beta1Peerauthentications).
		MustAdd(IstioSecurityV1Beta1Requestauthentications).
		Build()

	// Istio contains only Istio collections.
	Istio = collection.NewSchemasBuilder().
		MustAdd(IstioMeshV1Alpha1MeshConfig).
		MustAdd(IstioMeshV1Alpha1MeshNetworks).
		MustAdd(IstioNetworkingV1Alpha3Destinationrules).
		MustAdd(IstioNetworkingV1Alpha3Envoyfilters).
		MustAdd(IstioNetworkingV1Alpha3Gateways).
		MustAdd(IstioNetworkingV1Alpha3Serviceentries).
		MustAdd(IstioNetworkingV1Alpha3Sidecars).
		MustAdd(IstioNetworkingV1Alpha3Virtualservices).
		MustAdd(IstioNetworkingV1Alpha3Workloadentries).
		MustAdd(IstioNetworkingV1Alpha3Workloadgroups).
		MustAdd(IstioSecurityV1Beta1Authorizationpolicies).
		MustAdd(IstioSecurityV1Beta1Peerauthentications).
		MustAdd(IstioSecurityV1Beta1Requestauthentications).
		Build()

	// Kube contains only kubernetes collections.
	Kube = collection.NewSchemasBuilder().
		Build()

	// Pilot contains only collections used by Pilot.
	Pilot = collection.NewSchemasBuilder().
		MustAdd(IstioNetworkingV1Alpha3Destinationrules).
		MustAdd(IstioNetworkingV1Alpha3Envoyfilters).
		MustAdd(IstioNetworkingV1Alpha3Gateways).
		MustAdd(IstioNetworkingV1Alpha3Serviceentries).
		MustAdd(IstioNetworkingV1Alpha3Sidecars).
		MustAdd(IstioNetworkingV1Alpha3Virtualservices).
		MustAdd(IstioNetworkingV1Alpha3Workloadentries).
		MustAdd(IstioNetworkingV1Alpha3Workloadgroups).
		MustAdd(IstioSecurityV1Beta1Authorizationpolicies).
		MustAdd(IstioSecurityV1Beta1Peerauthentications).
		MustAdd(IstioSecurityV1Beta1Requestauthentications).
		Build()

	// PilotServiceApi contains only collections used by Pilot, including experimental Service Api.
	PilotServiceApi = collection.NewSchemasBuilder().
			MustAdd(IstioNetworkingV1Alpha3Destinationrules).
			MustAdd(IstioNetworkingV1Alpha3Envoyfilters).
			MustAdd(IstioNetworkingV1Alpha3Gateways).
			MustAdd(IstioNetworkingV1Alpha3Serviceentries).
			MustAdd(IstioNetworkingV1Alpha3Sidecars).
			MustAdd(IstioNetworkingV1Alpha3Virtualservices).
			MustAdd(IstioNetworkingV1Alpha3Workloadentries).
			MustAdd(IstioNetworkingV1Alpha3Workloadgroups).
			MustAdd(IstioSecurityV1Beta1Authorizationpolicies).
			MustAdd(IstioSecurityV1Beta1Peerauthentications).
			MustAdd(IstioSecurityV1Beta1Requestauthentications).
			Build()

	// Deprecated contains only collections used by that will soon be used by nothing.
	Deprecated = collection.NewSchemasBuilder().
			Build()
)
