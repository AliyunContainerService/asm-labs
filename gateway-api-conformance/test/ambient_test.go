package test

import (
	"context"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"istio.io/istio/pkg/maps"
	"k8s.io/apimachinery/pkg/types"
	k8ssets "k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/gateway-api/conformance"
	confv1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/yaml"

	"istio.io/api/label"
	"istio.io/istio/pkg/test/scopes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"alibabacloud.com/servicemesh/test/ack"
	"alibabacloud.com/servicemesh/test/asm"
	"alibabacloud.com/servicemesh/test/utils"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	conformanceconfig "sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	gwfeatures "sigs.k8s.io/gateway-api/pkg/features"
)

// defined in sigs.k8s.io/gateway-api/conformance/base/manifests.yaml
var conformanceNamespaces = []string{
	"gateway-conformance-infra",
	"gateway-conformance-app-backend",
	"gateway-conformance-web-backend",
	"gateway-conformance-mesh",
}

var skippedTests = map[string]string{}

func Test_AmbientGatewayAPIConformance(t *testing.T) {
	ackOperator, err := ack.NewACK()
	require.NoError(t, err, "error in NewACK")
	asmOperator, err := asm.NewASM()
	require.NoError(t, err, "error in NewASM")
	hostnameType := v1.AddressType("Hostname")
	supported := gwfeatures.AllFeatures.Clone().Delete(gwfeatures.MeshConsumerRouteFeature)
	kubeOpts, err := utils.GetKubeClientOptions()
	require.NoError(t, err, "failed to get kubeclient options")
	opts := suite.ConformanceOptions{
		// RunTest:                  "TLSRouteInvalidReferenceGrant",
		Client:                   ackOperator.KubeClient,
		ClientOptions:            kubeOpts,
		Clientset:                ackOperator.ClientSet,
		RestConfig:               ackOperator.RestConfig,
		GatewayClassName:         "istio",
		Debug:                    scopes.Framework.DebugEnabled(),
		CleanupBaseResources:     true,
		ManifestFS:               []fs.FS{&conformance.Manifests},
		SupportedFeatures:        gwfeatures.SetsToNamesSet(supported),
		SkipTests:                maps.Keys(skippedTests),
		UsableNetworkAddresses:   []v1.GatewaySpecAddress{{Value: "infra-backend-v1.gateway-conformance-infra.svc.cluster.local", Type: &hostnameType}},
		UnusableNetworkAddresses: []v1.GatewaySpecAddress{{Value: "foo", Type: &hostnameType}},
		AllowCRDsMismatch:        true,
		ConformanceProfiles: k8ssets.New(
			suite.GatewayHTTPConformanceProfileName,
			suite.GatewayTLSConformanceProfileName,
			suite.GatewayGRPCConformanceProfileName,
			suite.MeshHTTPConformanceProfileName,
		),
		Implementation: confv1.Implementation{
			Organization: "Alibaba Cloud",
			Project:      "Alibaba Cloud Service Mesh",
			URL:          "https://www.alibabacloud.com/help/zh/asm/",
			Version:      "1.27",
			Contact:      []string{"https://www.alibabacloud.com/help/zh/asm/"},
		},
		NamespaceLabels: map[string]string{
			label.IoIstioDataplaneMode.Name: "ambient",
		},
		TimeoutConfig: conformanceconfig.DefaultTimeoutConfig(),
	}
	csuite, err := suite.NewConformanceTestSuite(opts)
	require.NoError(t, err, "error in NewConformanceTestSuit")
	csuite.Setup(t, tests.ConformanceTests)
	asmPostSetup(t, asmOperator)
	ackPostSetup(t, ackOperator)
	require.NoError(t, csuite.Run(t, tests.ConformanceTests))
	report, err := csuite.Report()
	require.NoError(t, err)
	reportb, err := yaml.Marshal(report)
	require.NoError(t, err)
	t.Logf("writing conformance test report")
	assert.NoError(t, os.WriteFile("conformance.yaml", reportb, 0o644))
}

func asmPostSetup(t *testing.T, asmOperator *asm.ASM) {
	// Precreate the GatewayConformance namespaces in asm
	for _, ns := range conformanceNamespaces {
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}
		if ns != "gateway-conformance-infra" {
			namespace.Labels = map[string]string{
				"istio.io/dataplane-mode": "ambient",
			}
		}
		// Try to create the namespace, and ignore AlreadyExists errors
		err := asmOperator.KubeClient.Create(context.TODO(), namespace)
		if client.IgnoreAlreadyExists(err) != nil {
			// If it's not an AlreadyExists error, it's a real error
			require.NoError(t, err, "error in create ns in ASM")
		}
	}
}

func ackPostSetup(t *testing.T, ackOperator *ack.ACK) {
	// remove dataplane-mode label in ack gateway-conformance-infra namespace using patch
	patch := []byte(`{"metadata":{"labels":{"istio.io/dataplane-mode": null}}}`)
	_, err := ackOperator.ClientSet.CoreV1().Namespaces().Patch(t.Context(), "gateway-conformance-infra",
		types.MergePatchType, patch, metav1.PatchOptions{FieldManager: "ack-operator"})
	require.NoError(t, err, "error in patch ns gateway-conformance-infra")

	// create waypoint for ns gateway-conformance-mesh
	gateway := &v1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "waypoint",
			Namespace: "gateway-conformance-mesh",
		},
		Spec: v1.GatewaySpec{
			GatewayClassName: "istio-waypoint",
			Listeners: []v1.Listener{{
				Name:     "mesh",
				Port:     15008,
				Protocol: "HBONE",
			}},
		},
	}
	err = ackOperator.KubeClient.Create(context.TODO(), gateway)
	require.NoError(t, client.IgnoreAlreadyExists(err), "error in create waypoint gateway")
	patch = []byte(`{"metadata":{"labels":{"istio.io/use-waypoint": "waypoint"}}}`)
	_, err = ackOperator.ClientSet.CoreV1().Namespaces().Patch(t.Context(), "gateway-conformance-mesh",
		types.MergePatchType, patch, metav1.PatchOptions{FieldManager: "ack-operator"})
	require.NoError(t, err, "error in patch ns gateway-conformance-mesh")
}
