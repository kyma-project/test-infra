#!/usr/bin/env bash

set -o errexit
set -o pipefail


export KYMA_SOURCES_DIR="./kyma"

function prereq_test() {
  command -v node >/dev/null 2>&1 || { echo >&2 "node not found"; exit 1; }
  command -v npm >/dev/null 2>&1 || { echo >&2 "npm not found"; exit 1; }
  command -v jq >/dev/null 2>&1 || { echo >&2 "jq not found"; exit 1; }
  command -v helm >/dev/null 2>&1 || { echo >&2 "helm not found"; exit 1; }
  command -v kubectl >/dev/null 2>&1 || { echo >&2 "kubectl not found"; exit 1; }
  command -v k3d >/dev/null 2>&1 || { echo >&2 "k3d not found"; exit 1; }
  command -v go >/dev/null 2>&1 || { echo >&2 "go not found"; exit 1; }
}

function load_env() {
  ENV_FILE=".env"
  if [ -f "${ENV_FILE}" ]; then
    # shellcheck disable=SC2046
    export $(xargs < "${ENV_FILE}")
  fi
  export PATH="${PATH}:/usr/local/go/bin"
  export PATH="${PATH}:~/go/bin"
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

}

function provision_cluster() {
    # provision k3d clusters with exposed 80, 443 and 32000-32767
    sudo kyma provision k3d  --k3d-arg "--api-port=6443" --k3d-arg "--port=32000-32767:32000-32767@loadbalancer" --k3s-arg "--tls-san=$MACHINE_IP@servers:*"
}

function get_kubeconfig() {
    sudo k3d kubeconfig get kyma |sudo tee kubeconfig.yaml > /dev/null
    sed -e "s/0.0.0.0/$MACHINE_IP/g" kubeconfig.yaml
}

load_env
prereq_test
install_cli
provision_cluster
get_kubeconfig
