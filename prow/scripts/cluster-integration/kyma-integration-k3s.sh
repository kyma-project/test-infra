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
    ./install-istio.sh -f config-istio.yaml
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
