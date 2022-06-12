module gitlab.alibaba-inc.com/cos/asm-se-syncer

go 1.14

require (
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.870
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/envoyproxy/go-control-plane v0.9.9-0.20210115003313-31f9241a16e6
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.4.3
	github.com/hashicorp/consul/api v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.1
	google.golang.org/grpc v1.35.0
	istio.io/api v0.0.0-20210219010445-724943e9da20
	istio.io/client-go v0.0.0-20200908160912-f99162621a1a
	istio.io/istio v0.0.0-20210219021219-b4d95971757d
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.1
	sigs.k8s.io/controller-runtime v0.7.0
)

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.6.0
