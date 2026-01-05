package asm

import (
	"fmt"
	"os"

	"alibabacloud.com/servicemesh/test/utils"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	servicemesh20200111 "github.com/alibabacloud-go/servicemesh-20200111/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials"
)

type ASM struct {
	asmID      string
	client     *servicemesh20200111.Client
	KubeClient client.Client
}

func NewASM() (*ASM, error) {
	endpoint := utils.FromEnvOr("ASM_CLIENT_ENDPOINT", "servicemesh-share.aliyuncs.com")
	credentialConfig := new(credentials.Config).
		SetType("access_key").
		SetAccessKeyId(os.Getenv("ALICLOUD_ACCESS_KEY")).
		SetAccessKeySecret(os.Getenv("ALICLOUD_SECRET_KEY"))
	credential, _err := credentials.NewCredential(credentialConfig)
	if _err != nil {
		return nil, _err
	}

	config := &openapi.Config{
		Credential: credential,
		Endpoint:   &endpoint,
	}
	// Endpoint 请参考 https://api.aliyun.com/product/servicemesh
	config.Endpoint = tea.String(endpoint)
	client, err := servicemesh20200111.NewClient(config)
	if err != nil {
		return nil, err
	}

	a := &ASM{
		client: client,
	}
	if err := a.initKubeClient(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *ASM) initKubeClient() error {
	// Get the ASM cluster ID from environment variable
	asmID := utils.FromEnvOr("ASM_ID", "")
	if asmID == "" {
		return fmt.Errorf("ASM_ID env must be set.")
	}
	a.asmID = asmID

	// Get the kubeconfig for the ASM cluster
	describeClusterKubeconfigRequest := &servicemesh20200111.RevokeKubeconfigRequest{
		ServiceMeshId: tea.String(asmID),
	}

	resp, err := a.client.RevokeKubeconfig(describeClusterKubeconfigRequest)
	if err != nil {
		return fmt.Errorf("failed to get ASM cluster kubeconfig: %v", err)
	}

	kubeconfig := utils.Deref(resp.Body.Kubeconfig)

	// Create a Kubernetes client from the kubeconfig
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return fmt.Errorf("failed to create REST config from kubeconfig: %v", err)
	}

	kubeClient, err := client.New(restConfig, client.Options{})
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	a.KubeClient = kubeClient
	return nil
}
