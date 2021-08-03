#!/usr/bin/env bash

set -o errexit
set -o pipefail

export KYMA_SOURCES_DIR="./kyma"
export LOCAL_KYMA_DIR="./local-kyma"
export K3S_DOMAIN="local.kyma.dev"

function prereq_test() {
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

function prepare_k3s() {
    pushd ${LOCAL_KYMA_DIR}
    ./create-cluster-k3s.sh

    REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /registry.localhost)
    echo "${REGISTRY_IP} registry.localhost" >> /etc/hosts
    
    popd
}

function run_tests() {
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
function install_cli() {
  local install_dir
  declare -r install_dir="/usr/local/bin"
  mkdir -p "$install_dir"

  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  if [[ -z "$os" || ! "$os" =~ ^(darwin|linux)$ ]]; then
    echo >&2 -e "Unsupported host OS. Must be Linux or Mac OS X."
    exit 1
  else
    readonly os
  fi

  pushd "$install_dir" || exit
  curl -Lo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}"
  chmod +x kyma
  popd

  kyma version --client
}

function deploy_kyma() {

  if [[ -v CENTRAL_APPLICATION_GATEWAY_ENABLED ]]; then
      kyma deploy -d "local.kyma.dev" -p evaluation --ci --verbose --source=local --workspace "${KYMA_SOURCES_DIR}" --value application-connector.central_application_gateway.enabled=true
  else
      kyma deploy -d "local.kyma.dev" -p evaluation --ci --verbose --source=local --workspace "${KYMA_SOURCES_DIR}" 
  fi

  kubectl get pods -n kyma-system

#   if [[ -v COMPASS_INTEGRATION_ENABLED ]]; then
#     kubectl create namespace compass-system
#     kubectl label namespace compass-system istio-injection=enabled --overwrite
#     kubectl get namespace -L istio-injection
#   fi
kyma deploy --ci \
--component istio \
--value "global.ingress.domainName=local.kyma.dev"

  kubectl get pods -n kyma-system
}

prereq_test
load_env
prepare_k3s
install_cli
deploy_kyma
run_tests
