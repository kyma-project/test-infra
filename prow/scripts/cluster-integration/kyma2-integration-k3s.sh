#!/usr/bin/env bash

set -o errexit
set -o pipefail

export KYMA_SOURCES_DIR="./kyma"

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

prepare_k3s() {
    echo "starting cluster"
    # TODO pass OIDC setup 
    curl -sfL https://get.k3s.io | K3S_KUBECONFIG_MODE=777 INSTALL_K3S_EXEC="server --disable traefik" sh -
    mkdir -p ~/.kube
    cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
    chmod 600 ~/.kube/config

    REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /registry.localhost)
    echo "${REGISTRY_IP} registry.localhost" >> /etc/hosts
}

install_kyma() {
    pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
    npm install
    ./kyma.js install 
    popd
}

prereq_test
load_env
prepare_k3s
install_kyma
