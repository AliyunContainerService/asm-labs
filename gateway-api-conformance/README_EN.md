# Alibaba Cloud Service Mesh Gateway API Conformance Test Instructions
## About Alibaba Cloud Service Mesh
Alibaba Cloud Service Mesh (Service Mesh, abbreviated as ASM) provides a fully managed service mesh platform that is compatible with the community Istio open source service mesh. It simplifies service governance, including traffic routing and split management between service calls, authentication security for inter-service communication, and mesh observability capabilities, thereby greatly reducing the workload of development and operations.

## Before You Begin
1. Ensure you have created an ACK cluster, see:
   - https://www.alibabacloud.com/help/zh/ack/ack-managed-and-ack-dedicated/user-guide/create-an-ack-managed-cluster-2/?spm=a2c63.p38356.help-menu-85222.d_2_0_1.4c287474VErtwq
1. Ensure you have installed the servicemesh-operator component in your ACK cluster and have obtained the ASM ID. See:
   - https://help.aliyun.com/zh/ack/product-overview/servicemesh-operator
1. Have an Alibaba Cloud AccessKey that can operate all ACK and ASM resources and can obtain administrator-level KubeConfig.

## Start Testing
```sh
# Set test environment variables
export ALICLOUD_ACCESS_KEY='xxxxx'
export ALICLOUD_SECRET_KEY='xxxxx'
export ACK_ID='xxxxx'
export ASM_ID='xxxxx'
```
```sh
go test Test_AmbientGatewayAPIConformance ./test/ -v
```
