#!/bin/bash
#
# Run fast-integration tests.
# Install k3d and Kyma CLI to provision Kyma on k3d cluster as prerequisite.

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

install_k3d() {
  # TODO pin version?
  curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash
  k3d --version 
}

install_cli() {
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

deploy_kyma() {
  # TODO pin version?
  kyma alpha provision k3s --ci
  kyma alpha deploy --ci --verbose --source=local --workspace "${KYMA_SOURCES_DIR}"
  kubectl get pods -n kyma-system
}

run_tests() {
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
install_k3d
install_cli
deploy_kyma
run_tests
