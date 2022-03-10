ASM_KUBECONFIG=./asm_kube_config.conf
USER_KUBECONFIG=./user_kube_config.conf

ssl:
	openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=myexample Inc./CN=aliyun.com' -keyout aliyun.root.key -out aliyun.root.crt
	openssl req -out aliyun.com.csr -newkey rsa:2048 -nodes -keyout aliyun.com.key -subj "/CN=aliyun.com/O=myexample organization"
	openssl x509 -req -days 365 -CA aliyun.root.crt -CAkey aliyun.root.key -set_serial 0 -in aliyun.com.csr -out aliyun.com.crt

install: ssl
  	curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
  	cp kubectl /usr/local/bin/kubectl
  	chmod +x /usr/local/bin/kubectl
	kubectl --kubeconfig  ${USER_KUBECONFIG} create -n istio-system secret tls istio-ingressgateway-certs --key ./aliyun.com.key --cert ./aliyun.com.crt
	kubectl --kubeconfig ${USER_KUBECONFIG} create configmap myexample-nginx-configmap --from-file=nginx.conf=./deploy/myexample-nginx.conf
	kubectl --kubeconfig ${USER_KUBECONFIG} apply -f ./deploy/backend.yaml
	kubectl --kubeconfig ${ASM_KUBECONFIG} apply -f ./deploy/istiogateway.yaml 
	kubectl --kubeconfig ${ASM_KUBECONFIG} label namespace default istio-injection=enabled
	kubectl --kubeconfig ${ASM_KUBECONFIG} apply -f ./deploy/gateway.yaml
	kubectl --kubeconfig ${ASM_KUBECONFIG} apply -f ./deploy/virtualservice.yaml

perf_disable:
	./demo/perf.sh disable

perf_enable:
	./demo/perf.sh enable

