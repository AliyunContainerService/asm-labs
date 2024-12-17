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
    if [[ "$ISTIO_INJECT" == "true" ]]
    then
        kubectl label namespace "${NAMESPACE}" istio-injection=enabled --overwrite || true
    fi

    mkdir results || true
    
    helm install massive-subsets -n "${NAMESPACE}" . \
        --set deployReplicas=$DEPLOY_REPLICAS \
        --set namespace=${NAMESPACE}
    
    kubectl rollout status deployment fortio -n "${NAMESPACE}" --timeout=5m

    end_time=$((SECONDS + $TIMEOUT))

    while [ $SECONDS -lt $end_time ]; do
        ready_replicas=$(kubectl get pod -n ${NAMESPACE} | grep httpbin | grep 2/2 | wc -l)
        if [ "$ready_replicas" -eq "$DEPLOY_REPLICAS" ]; then
            echo "all $DEPLOY_REPLICAS replicas are ready."
            break
        else
            echo "waiting for replicas to be ready. current ready replicas: $ready_replicas/$DEPLOY_REPLICAS"
            sleep 5
        fi
    done
}

function cleanup_test() {
    helm uninstall massive-subsets -n "${NAMESPACE}"
    kubectl delete ns "${NAMESPACE}" || true
}

function inject() {
    kubectl label namespace "${NAMESPACE}" istio-injection=enabled --overwrite || true
    kubectl get pods | grep '1/1' | awk '{print$1}' | xargs kubectl delete pod
}

function uninject() {
    kubectl label namespace "${NAMESPACE}" istio-injection- || true
    kubectl get pods | grep '2/2' | awk '{print$1}' | xargs kubectl delete pod
}

function start_test() {
    timestamp=$(date +%Y%m%d%H%M%S)
    current_replicas=$(kubectl get pod -n ${NAMESPACE} | grep httpbin | wc -l | xargs)
    filename="fortio_load_${current_replicas}_${timestamp}_CON${CONNECTIONS}_QPS${QPS}_injection_${ISTIO_INJECT}.json"

    kubectl exec -it deploy/fortio -n ${NAMESPACE} -- fortio load -c ${CONNECTIONS} -t ${DURATION} -qps ${QPS} -json ${filename} http://httpbin:8000
    kubectl exec -it deploy/sleep -n ${NAMESPACE} -- curl http://fortio:8080/fortio/data/${filename} > results/${filename}
}

function usage() {
    echo "Usage: $0 [-s] [--clean]"
    echo "  -s | --setup      Setup test environment"
    echo "  -c | --clean      Cleanup test environment"
    echo "  -t | --test       Run the test"
    echo "  -i | --inject     Inject/uninject sidecar into pods. e.g. --inject true"
    echo "  -d | --delete     Delete resources, options: vs, dr, acb. e.g. --delete vs"
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
        -t|--test)
            start_test
            shift
            ;;
        -i|--inject)
            if [[ "$#" -lt 2 ]]; then
                echo "Error: --inject requires a value (e.g., --inject true)"
                usage
            fi
            inject_option="$2"
            case $inject_option in
                true)
                    inject
                    ;;
                false)
                    uninject
                    ;;
                *)
                    echo "Error: Unknown clean option '$inject_option'"
                    usage
                    ;;
            esac
            shift 2
            ;;
        -d|--delete)
            if [[ "$#" -lt 2 ]]; then
                echo "Error: --clean requires a value (e.g., --delete all)"
                usage
            fi
            clean_option="$2"
            case $clean_option in
                vs)
                    kubectl delete virtualservice -n ${NAMESPACE} --all
                    ;;
                dr)
                    kubectl delete destinationrule -n ${NAMESPACE} --all
                    ;;
                acb)
                    kubectl delete asmcircuitbreakers -n ${NAMESPACE} --all
                    ;;
                *)
                    echo "Error: Unknown delete option '$clean_option'"
                    usage
                    ;;
            esac
            shift 2
            ;;
        *)
            usage
            ;;
    esac
done
