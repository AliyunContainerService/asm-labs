#!/bin/bash

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

source config.sh

function setup_test() {
    kubectl create ns "${NAMESPACE}" || true
    kubectl label namespace "${NAMESPACE}" istio-injection=enabled --overwrite || true
    
    helm install ${TEST_NAME} -n "${NAMESPACE}" . \
        --set httpbinReplicas=$HTTPBIN_REPLICAS
}

function cleanup_test() {
    helm uninstall ${TEST_NAME} -n "${NAMESPACE}"
    kubectl delete ns "${NAMESPACE}" || true
}

function start_test() {
    kubectl exec -it deploy/sleep -n ${NAMESPACE} -- sh -c '
i=1
while [ $i -le 100 ]; do
    target=$((RANDOM % 400 + 1))
    echo "$i request to httpbin-v${target} at $(date)"
    curl httpbin-v${target}:8000 -s -o /dev/null
    i=$((i+1))
    sleep 0.2
done'
}

function restart() {
    i=1
    while [ $i -le $SCALEUP_HTTPBIN_REPLICAS ]; do
        kubectl rollout restart deploy/httpbin-v$i -n ${NAMESPACE} > /dev/null &
        i=$((i+1))
    done
}

function scaleup() {
    helm upgrade ${TEST_NAME} -n "${NAMESPACE}" . \
        --reuse-values \
        --set httpbinReplicas=$SCALEUP_HTTPBIN_REPLICAS
}

function scaledown() {
    helm upgrade ${TEST_NAME} -n "${NAMESPACE}" . \
        --reuse-values \
        --set httpbinReplicas=$HTTPBIN_REPLICAS

    kubectl rollout restart deploy/sleep -n ${NAMESPACE}
}

function usage() {
    echo "Usage: $0 [-s] [--clean]"
    echo "  -s | --setup      Setup test environment"
    echo "  -c | --clean      Cleanup test environment"
    echo "  -t | --test       Start test traffic"
    echo "  -su | --scaleup   Scale up httpbin replicas"
    echo "  -sd | --scaledown Scale down httpbin replicas"
    echo "  -r | --restart    Restart httpbin pods"
    exit 1
}

if [ "$#" -eq 0 ]; then
    usage
fi

while [[ "$#" -gt 0 ]]; do
    case $1 in
        -s|--setup)
            setup_test
            shift
            ;;
        -c|--clean)
            cleanup_test
            shift
            ;;
        -su|--scaleup)
            scaleup
            shift
            ;;
        -sd|--scaledown)
            scaledown
            shift
            ;;
        -t|--test)
            start_test
            shift
            ;;
        -r|--restart)
            restart
            shift
            ;;
    esac
done
