#!/usr/bin/env bash
export DEBUG=1
# export OIDC_ISSUER=$1
# export OIDC_CLIENT_ID=$2

export KYMA_SOURCES_DIR="./kyma"

set -o errexit
set -o pipefail

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
    export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
}

prepare_cli(){
    export PATH="$PWD/bin:${PATH}"
    kyma alpha deploy --help
}

prepare_k3s() {
    pushd "${KYMA_SOURCES_DIR}/tests/fast-integration/hack"
    ./create-cluster-k3s.sh
    popd
}

install_kyma() {
    envsubst < "./k3s-overrides.tpl.yaml" > "./k3s-overrides.yaml"
    kyma alpha deploy -w ${KYMA_SOURCES_DIR} -s local -p evaluation -f ./k3s-overrides.yaml --non-interactive
}

# patch_coredns(){
#     export REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /registry.localhost) 
#     kubectl -n kube-system patch cm coredns --patch "$(cat k3d-coredns-patch.yaml | envsubst )"
# }

test_kyma() {
    pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
    nom install
    npm run test-no-install
    popd
}

prereq_test
load_env
prepare_cli
prepare_k3s
install_kyma
# patch_coredns
test_kyma
