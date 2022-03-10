#!/bin/bash

if [ $1 = "disable" ]; then
 echo "Perf ASM gateway with multibuffer disabled"
 SERVER_HOST=$(kubectl --kubeconfig=./user_kube_config.conf  -n istio-system get pod -l app=istio-ingressgateway -o=jsonpath="{range .items[*]}{.status.podIP}{end}") k6 run --vus 200 --duration 10s demo.js
elif [ $1 = "enable" ]; then
 echo "Perf ASM gateway with multibuffer enabled"
 SERVER_HOST=$(kubectl --kubeconfig=./user_kube_config.conf  -n istio-system get pod -l app=istio-ingressgateway -o=jsonpath="{range .items[*]}{.status.podIP}{end}") k6 run --vus 200 --duration 10s demo.js
else
 echo "no support !"
fi
