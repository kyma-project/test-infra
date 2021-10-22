#!/usr/bin/env bash

set -o errexit
set -o pipefail

export KYMA_SOURCES_DIR="./kyma"
export LOCAL_KYMA_DIR="./local-kyma"

prereq_test() {
    command -v node >/dev/null 2>&1 || { echo >&2 "node not found"; exit 1; }
    command -v npm >/dev/null 2>&1 || { echo >&2 "npm not found"; exit 1; }
    command -v jq >/dev/null 2>&1 || { echo >&2 "jq not found"; exit 1; }
    command -v helm >/dev/null 2>&1 || { echo >&2 "helm not found"; exit 1; }
    command -v kubectl >/dev/null 2>&1 || { echo >&2 "kubectl not found"; exit 1; }
}

load_env() {
    ENV_FILE=".env"
    if [ -f "${ENV_FILE}" ]; then
        # shellcheck disable=SC2046
        export $(xargs < "${ENV_FILE}")
    fi
}

prepare_k3s() {
    pushd ${LOCAL_KYMA_DIR}
    ./create-cluster-k3s.sh

    REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /registry.localhost)
    echo "${REGISTRY_IP} registry.localhost" >> /etc/hosts
    
    popd
}

run_tests() {
    pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
    if [[ -v COMPASS_INTEGRATION_ENABLED && -v CENTRAL_APPLICATION_GATEWAY_ENABLED ]]; then
        make ci-application-connectivity-2-compass
    elif [[ -v COMPASS_INTEGRATION_ENABLED ]]; then
        make ci-compass
    elif [[ -v CENTRAL_APPLICATION_GATEWAY_ENABLED ]]; then
        make ci-application-connectivity-2
    else
        make ci
    fi
    popd
}

prereq_test
load_env
prepare_k3s
run_tests
