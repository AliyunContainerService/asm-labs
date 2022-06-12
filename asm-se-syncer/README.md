External Services Synchronizer是一个网格外服务信息同步组件，将服务的地址、端口、协议等信息同步为Istio中的ServiceEntry，提供网格内服务调用注册中心中注册的网格外部服务的能力。

此组件能够帮助您在微服务迁移服务网格的过程中，网格内的服务需要调用存量的注册在如Consul、Nacos中的外部服务。计划支持多种注册中心类型，支持自定义 MCP Server 和向 API Server 写入 ServiceEntry这2种同步方式。 

操作步骤： https://help.aliyun.com/document_detail/202143.html