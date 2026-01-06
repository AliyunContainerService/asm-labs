# 阿里云服务网格Gateway API一致性测试说明
## 关于阿里云服务网格
阿里云服务网格（Service Mesh，简称ASM）提供了一个全托管式的服务网格平台，兼容于社区Istio开源服务网格，用于简化服务的治理，包括服务调用之间的流量路由与拆分管理、服务间通信的认证安全以及网格可观测性能力，从而极大地减轻开发与运维的工作负担。
## 开始之前
1. 请确保您已经创建了一个ACK集群，参考：https://www.alibabacloud.com/help/zh/ack/ack-managed-and-ack-dedicated/user-guide/create-an-ack-managed-cluster-2/?spm=a2c63.p38356.help-menu-85222.d_2_0_1.4c287474VErtwq
1. 请确保您已经在ACK集群中安装了servicemesh-operator组件，并且已经获取了ASM ID。参考：https://help.aliyun.com/zh/ack/product-overview/servicemesh-operator
1. 已经有一个阿里云AccessKey，可以操作所有ACK和ASM的资源，并且可以获取到管理员权限的KubeConfig。
## 开始测试
```sh
# 设置测试环境变量
export ALICLOUD_ACCESS_KEY='xxxxx'
export ALICLOUD_SECRET_KEY='xxxxx'
export ACK_ID='xxxxx'
export ASM_ID='xxxxx'
```
```sh
go test Test_AmbientGatewayAPIConformance ./test/ -v
```
