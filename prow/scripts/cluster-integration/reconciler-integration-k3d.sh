#!/usr/bin/env bash

set -o errexit
set -o pipefail

readonly RECONCILER_DIR="./reconciler"
readonly GO_VERSION=1.16.6

function prereq_test() {
  command -v node >/dev/null 2>&1 || { echo >&2 "node not found"; exit 1; }
  command -v npm >/dev/null 2>&1 || { echo >&2 "npm not found"; exit 1; }
  command -v jq >/dev/null 2>&1 || { echo >&2 "jq not found"; exit 1; }
  command -v helm >/dev/null 2>&1 || { echo >&2 "helm not found"; exit 1; }
  command -v kubectl >/dev/null 2>&1 || { echo >&2 "kubectl not found"; exit 1; }
  command -v k3d >/dev/null 2>&1 || { echo >&2 "k3d not found"; exit 1; }
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

function provision_k3d() {
  kyma alpha provision k3d -p 80:80@loadbalancer -p 443:443@loadbalancer
}

function run_tests() {
  echo "Install Go"
  wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && export PATH=$PATH:/usr/local/go/bin && go version

  export KUBECONFIG=~/.kube/config
  pushd "${RECONCILER_DIR}"
  make test-all
  popd
}

prereq_test
install_cli
provision_k3d
run_tests
