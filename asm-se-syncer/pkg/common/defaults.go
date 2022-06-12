package common

import "strings"

const (
	OwnerRefAPIVersion = "istio.alibabacloud.com/v1beta1"
	OwnerRefKind       = "ASMServiceRegistry"

	IsSameVpc      = "IS_SAME_VPC"
	AsmSyncerLabel = "ASM_Syncer"
)

type ServiceRegistryType string

const (
	Consul ServiceRegistryType = "consul"
	Nacos  ServiceRegistryType = "nacos"
)

func FormatedName(hostName string) string {
	cleanHostName := strings.Replace(hostName, "_", "-", -1)
	return cleanHostName
}
