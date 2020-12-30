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
}

install_kyma() {
    mv "${KYMA_SOURCES_DIR}/resources" "${LOCAL_KYMA_DIR}/"

    pushd ${LOCAL_KYMA_DIR}
    ./create-cluster-k3s.sh

    REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /registry.localhost)
    echo "${REGISTRY_IP} registry.localhost" >> /etc/hosts
    
    ./install-istio.sh -f config-istio.yaml
    
    REGISTRY_VALUES="dockerRegistry.enableInternal=false,dockerRegistry.serverAddress=registry.localhost:5000,dockerRegistry.registryAddress=registry.localhost:5000,containers.manager.envs.functionBuildExecutorImage.value=eu.gcr.io/kyma-project/external/aerfio/kaniko-executor:v1.3.0" \
      ./install-kyma.sh
    
    popd
}

run_tests() {
    pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"

    npm install
    npm test
    
    popd
}

prereq_test
install_kyma
run_tests
