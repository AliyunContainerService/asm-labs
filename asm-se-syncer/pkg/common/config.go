package common

import (
	"encoding/json"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/servicemesh"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io/ioutil"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strconv"
)

const (
	serviceRegistryConfigDir = "/etc/istio/serviceregistry/"
	AccessKeyIDPath          = serviceRegistryConfigDir + "AccessKeyID"
	AccessKeySecretPath      = serviceRegistryConfigDir + "AccessKeySecret"
	RegistryConfigPath       = serviceRegistryConfigDir + "RegistryConfig"
	// 售卖区VPC内网访问OpenAPI https://yuque.antfin-inc.com/alibabacloud-openapi/pop-doc/eg5a6o
	privateOpenApiEndpoint = "servicemesh.vpc-proxy.aliyuncs.com"
	publicOpenApiEndpoint  = "servicemesh.aliyuncs.com"
)

func loadFileContent(path string) ([]byte, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Infof("error opening file %s\n", path)
		return nil, errors.Wrapf(err, "error opening file %s\n", path)
	}
	resBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Infof("error reading file %s\n", path)
		return nil, errors.Wrapf(err, "error reading file %s\n", path)
	}
	return resBytes, nil
}
func GetAccessKeyID() ([]byte, error) {
	return loadFileContent(AccessKeyIDPath)
}
func GetAccessKeySecret() ([]byte, error) {
	return loadFileContent(AccessKeySecretPath)
}
func GetRegistryConfig() ([]byte, error) {
	return loadFileContent(RegistryConfigPath)
}

func GetServiceRegistryConfig() ([]map[string]interface{}, error) {
	serviceRegistryConfigBytes, err := GetRegistryConfig()
	if err != nil {
		return nil, err
	}
	var serviceRegistryConfig []map[string]interface{}
	err = json.Unmarshal(serviceRegistryConfigBytes, &serviceRegistryConfig)
	if err != nil {
		log.Infof("error Unmarshal config file %q\n", RegistryConfigPath)
		return nil, errors.Wrapf(err, "error Unmarshal config file %q\n", RegistryConfigPath)
	}
	if serviceRegistryConfig == nil || len(serviceRegistryConfig) == 0 {
		log.Infof("error no service registry info within config file %q\n", RegistryConfigPath)
		return nil, errors.Wrapf(err, "error no service registry info within config file %q\n", RegistryConfigPath)
	}

	var serviceRegistryInfoList []map[string]interface{}
	for _, serviceRegistryInfo := range serviceRegistryConfig {
		serviceRegistryType := cast.ToString(serviceRegistryInfo["type"])
		if serviceRegistryType == string(Consul) || serviceRegistryType == string(Nacos) {
			serviceRegistryInfoList = append(serviceRegistryInfoList, serviceRegistryInfo)
		}
	}

	if len(serviceRegistryConfig) > 0 {
		return serviceRegistryInfoList, nil
	}
	return nil, errors.New("no available service registry config")
}

func GetASMRestConfig(meshId, regionId, accessKeyId, accessKeySecret string) (*restclient.Config, error) {
	kubeconfigString, err := getASMKubeConfig(meshId, regionId, accessKeyId, accessKeySecret)
	if err != nil {
		return nil, errors.Wrapf(err, "getASMKubeConfig failed to get the kubeconfig of asm %s", meshId)
	}
	clusterConfig, err := clientcmd.Load([]byte(kubeconfigString))
	if err != nil {
		return nil, errors.Wrapf(err, "could not load kubeconfig")
	}

	cfg, err := clientcmd.NewDefaultClientConfig(*clusterConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "could not create k8s rest config")
	}
	return cfg, nil
}

func getASMKubeConfig(meshId, regionId, accessKeyId, accessKeySecret string) (string, error) {
	client, err := servicemesh.NewClientWithAccessKey(regionId, accessKeyId, accessKeySecret)
	if err != nil {
		return "", err
	}
	request := servicemesh.CreateDescribeServiceMeshKubeconfigRequest()
	request.Scheme = "https"

	usePrivateIpAdddress := false
	mode, found := os.LookupEnv(IsSameVpc)
	if found {
		usePrivateIpAdddress, _ = strconv.ParseBool(mode)
	}

	request.PrivateIpAddress = requests.NewBoolean(usePrivateIpAdddress)
	request.ServiceMeshId = meshId

	response, err := client.DescribeServiceMeshKubeconfig(request)
	if err != nil || response == nil || response.Kubeconfig == "" {
		// 外网获取失败后，尝试使用内网地址
		client.Domain = privateOpenApiEndpoint
		request.Domain = privateOpenApiEndpoint
		request.PrivateIpAddress = requests.NewBoolean(true)
		responseInter, err := client.DescribeServiceMeshKubeconfig(request)
		if err != nil || responseInter == nil || responseInter.Kubeconfig == "" {
			log.Errorf("failed to get kubeconfig: %v", err)
			return "", err
		}
		return responseInter.Kubeconfig, nil
	}
	return response.Kubeconfig, nil

}
