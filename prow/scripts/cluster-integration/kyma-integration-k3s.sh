#!/usr/bin/env bash

set -o errexit
set -o pipefail

export KYMA_SOURCES_DIR="./kyma"
export LOCAL_KYMA_DIR="./local-kyma"

install::prereq() {
    curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
    curl -sL https://deb.nodesource.com/setup_14.x | bash -
    apt-get -y install jq nodejs
}

install::kyma() {
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

run::tests() {
    pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
    
    npm install
    npm test
    
    popd
}

run::sleep() {
    node -e 'setTimeout(() => {}, 1000*60*60);'
}

install::prereq
install::kyma
#run::tests
run::sleep
