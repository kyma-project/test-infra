#!/usr/bin/env bash

set -o errexit
set -o pipefail

readonly RECONCILER_DIR="./reconciler"
readonly GO_VERSION=1.19
readonly PG_MIGRATE_VERSION=v4.15.1
readonly INSTALL_DIR="/usr/local/bin"

function check_pre_requirements_for_tests() {
  echo "Check pre-requirements for test"
  command -v node >/dev/null 2>&1 || { echo >&2 "node not found"; exit 1; }
  command -v npm >/dev/null 2>&1 || { echo >&2 "npm not found"; exit 1; }
  command -v jq >/dev/null 2>&1 || { echo >&2 "jq not found"; exit 1; }
  command -v helm >/dev/null 2>&1 || { echo >&2 "helm not found"; exit 1; }
  command -v kubectl >/dev/null 2>&1 || { echo >&2 "kubectl not found"; exit 1; }
  command -v k3d >/dev/null 2>&1 || { echo >&2 "k3d not found"; exit 1; }
}

function create_local_bin_folder() {
    echo "Create local bin folder"
    mkdir -p $INSTALL_DIR
    export PATH=$PATH:$INSTALL_DIR
}

function install_cli() {
  echo "Install CLI"
  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  if [[ -z "$os" || ! "$os" =~ ^(darwin|linux)$ ]]; then
    echo >&2 -e "Unsupported host OS. Must be Linux or Mac OS X."
    exit 1
  else
    readonly os
  fi

  pushd $INSTALL_DIR || exit
  curl -Lo kyma "https://storage.googleapis.com/kyma-cli-unstable/kyma-${os}"
  chmod +x kyma
  popd

  kyma version --client
}

function provision_k3d() {
  kyma provision k3d --ci
}

function run_tests() {
  echo "Install Go"
  wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && export PATH=$PATH:/usr/local/go/bin && go version

 echo "Run all tests: make test all"
  export KUBECONFIG=~/.kube/config
  pushd "${RECONCILER_DIR}"
  make test-all
  popd
}

function provision_postgres() {
  echo "Starting Postgres"
  pushd $RECONCILER_DIR
  ./scripts/postgres.sh start
  popd
}

check_pre_requirements_for_tests
create_local_bin_folder
install_cli
provision_postgres
provision_k3d
run_tests
