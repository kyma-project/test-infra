#!/usr/bin/env bash

set -o errexit
set -o pipefail

export KYMA_SOURCES_DIR="./kyma"
export LOCAL_KYMA_DIR="./local-kyma"

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
    node -e 'setTimeout(()=>{},1000*60*10)'
    npm install
    npm test
    
    popd
}

install_kyma
run_tests
