package ack

import (
	"fmt"
	"os"

	"alibabacloud.com/servicemesh/test/utils"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cs20151215 "github.com/alibabacloud-go/cs-20151215/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials"
	"istio.io/istio/pkg/kube"
)

type ACK struct {
	ackID      string
	openApiCli *cs20151215.Client

	RestConfig *rest.Config
	KubeClient client.Client
	ClientSet  *clientset.Clientset
}

func NewACK() (*ACK, error) {
	endpoint := utils.FromEnvOr("ACK_CLIENT_ENDPOINT", "cs.cn-hangzhou.aliyuncs.com")
	ackID := utils.FromEnvOr("ACK_ID", "")
	if ackID == "" {
		return nil, fmt.Errorf("ACK_ID env must be set.")
	}

	credentialConfig := new(credentials.Config).
		SetType("access_key").
		SetAccessKeyId(os.Getenv("ALICLOUD_ACCESS_KEY")).
		SetAccessKeySecret(os.Getenv("ALICLOUD_SECRET_KEY"))
	credential, err := credentials.NewCredential(credentialConfig)
	if err != nil {
		return nil, err
	}

	config := &openapi.Config{
		Credential: credential,
		Endpoint:   &endpoint,
	}
	client, err := cs20151215.NewClient(config)
	if err != nil {
		return nil, err
	}
	a := &ACK{
		ackID:      ackID,
		openApiCli: client,
	}
	if err := a.initKubeClient(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *ACK) initKubeClient() error {
	describeClusterUserKubeconfigRequest := &cs20151215.DescribeClusterUserKubeconfigRequest{}
	resp, _err := a.openApiCli.DescribeClusterUserKubeconfig(tea.String(a.ackID), describeClusterUserKubeconfigRequest)
	if _err != nil {
		return _err
	}
	kubeconfig := utils.Deref(resp.Body.Config)

	// Create a Kubernetes client from the kubeconfig
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return fmt.Errorf("failed to create REST config from kubeconfig: %v", err)
	}
	a.RestConfig = restConfig

	kubeClient, err := client.New(restConfig, client.Options{Scheme: kube.IstioScheme})
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	a.KubeClient = kubeClient

	clientset, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	a.ClientSet = clientset

	return nil
}
