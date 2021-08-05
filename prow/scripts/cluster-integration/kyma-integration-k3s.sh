#!/usr/bin/env bash

set -o errexit
set -o pipefail

export KYMA_SOURCES_DIR="./kyma"
export LOCAL_KYMA_DIR="./local-kyma"

function prereq_test() {
  command -v node >/dev/null 2>&1 || { echo >&2 "node not found"; exit 1; }
  command -v npm >/dev/null 2>&1 || { echo >&2 "npm not found"; exit 1; }
  command -v jq >/dev/null 2>&1 || { echo >&2 "jq not found"; exit 1; }
  command -v helm >/dev/null 2>&1 || { echo >&2 "helm not found"; exit 1; }
  command -v kubectl >/dev/null 2>&1 || { echo >&2 "kubectl not found"; exit 1; }
}

function load_env() {
  ENV_FILE=".env"
  if [ -f "${ENV_FILE}" ]; then
    export $(xargs < "${ENV_FILE}")
  fi
}

# function prepare_k3s() {
#   pushd ${LOCAL_KYMA_DIR}
#   ./create-cluster-k3s.sh

#   REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /registry.localhost)
#   echo "${REGISTRY_IP} registry.localhost" >> /etc/hosts

#   popd
# }

function prepare_k3d(){
    wget -q -O - https://raw.githubusercontent.com/rancher/k3d/main/install.sh | TAG=v4.4.7 bash
    k3d version
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
  kyma provision k3d -p 80:80@loadbalancer -p 443:443@loadbalancer

  local kyma_deploy_cmd
  kyma_deploy_cmd="kyma deploy -p evaluation --ci --verbose --source=local --workspace ${KYMA_SOURCES_DIR}"

  if [[ -v CENTRAL_APPLICATION_GATEWAY_ENABLED ]]; then
    kyma_deploy_cmd+=" --value application-connector.central_application_gateway.enabled=true"
    kyma_deploy_cmd+=" --value application-connector.central_application_connectivity_validator.enabled=true"
  fi

  if [[ -v COMPASS_INTEGRATION_ENABLED ]]; then
    kubectl create namespace compass-system
    kubectl label namespace compass-system istio-injection=enabled --overwrite
    kubectl get namespace -L istio-injection
    kyma_deploy_cmd+=" --value global.disableLegacyConnectivity=true"
  fi

  $kyma_deploy_cmd

  kubectl get pods -n kyma-system
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

prereq_test
load_env
# prepare_k3s
prepare_k3d
install_cli
deploy_kyma
run_tests
