#!/usr/bin/env bash
export KYMA_SOURCES_DIR=~/go/src/github.com/kyma-project/kyma
export DEBUG=1
# export OIDC_ISSUER=$1
# export OIDC_CLIENT_ID=$2

# ./create-cluster-k3d-oidc.sh https://apskyxzcl.accounts400.ondemand.com 92dc3801-576b-465b-82ad-f69ed244d1e7
set -o errexit
set -o pipefail

# export KYMA_SOURCES_DIR="./kyma"

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
        export $(xargs < "${ENV_FILE}")
    fi
}

prepare_cli(){
    log::info "Building Kyma CLI"
    cd "${KYMA_PROJECT_DIR}/cli"
    make build-linux
    mv "${KYMA_PROJECT_DIR}/cli/bin/kyma-linux" "${KYMA_PROJECT_DIR}/cli/bin/kyma"
    export PATH="${KYMA_PROJECT_DIR}/cli/bin:${PATH}"
}

prepare_k3s() {
    pushd "${KYMA_SOURCES_DIR}/tests/fast-integration/hack"
    ./create-cluster-k3d-oidc.sh https://apskyxzcl.accounts400.ondemand.com 92dc3801-576b-465b-82ad-f69ed244d1e7
    popd
    # kyma alpha provision k3s --name "kyma" --server-args "--kube-apiserver-arg=oidc-issuer-url=${OIDC_ISSUER}" --server-args "--kube-apiserver-arg=oidc-client-id=${OIDC_CLIENT_ID}" --server-args "--kube-apiserver-arg=oidc-username-claim=sub" --server-args "--kube-apiserver-arg=oidc-username-prefix=-" --server-args "--kube-apiserver-arg=oidc-groups-claim=groups"
}

install_kyma() {
    pushd "${PWD}"
    kyma alpha deploy -w ${KYMA_SOURCES_DIR} -s local -p evaluation -f ./values.yaml --non-interactive
    popd
}

patch_coredns(){
    export REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /registry.localhost) 
    kubectl -n kube-system patch cm coredns --patch "$(cat k3d-coredns-patch.yaml | envsubst )"
}

test_kyma() {
    pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
    npm run test-no-install
    popd
}

prereq_test
load_env
prepare_cli
prepare_k3s
install_kyma
patch_coredns
# test_kyma
