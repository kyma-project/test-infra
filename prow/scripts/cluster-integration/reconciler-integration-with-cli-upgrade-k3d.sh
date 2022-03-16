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
}

function load_env() {
  ENV_FILE=".env"
  if [ -f "${ENV_FILE}" ]; then
    # shellcheck disable=SC2046
    export $(xargs < "${ENV_FILE}")
  fi
}

function install_reconciler_pr_cli() {
  local install_dir
  declare -r install_dir="/usr/local/bin"
  mkdir -p "$install_dir"

  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  if [[ -z "$os" || ! "$os" =~ ^(linux)$ ]]; then
    echo >&2 -e "Unsupported host OS. Must be Linux."
    exit 1
  else
    readonly os
  fi

  kyma_cli_url="https://storage.googleapis.com/kyma-cli-pr/kyma-${os}-pr-${PULL_NUMBER}"

  pushd "$install_dir" || exit
  echo "Downloading Kyma CLI from: ${kyma_cli_url}"
  curl -Lo kyma "${kyma_cli_url}"
  chmod +x kyma
  popd

  kyma version --client
}

function provision_k3d() {
  k3d version

  if [[ -v K8S_VERSION ]]; then
    echo "Creating k3d with kubernetes version: ${K8S_VERSION}"
    kyma provision k3d --ci -k "${K8S_VERSION}"
  else
    kyma provision k3d --ci
  fi

  echo "Printing client and server version info"
  kubectl version
}

function deploy_kyma() {
  echo "Deploying Kyma version: ${KYMA_SOURCE} using Execution profile: ${EXECUTION_PROFILE}"
  kyma deploy --ci --timeout 20m -p "$EXECUTION_PROFILE" --source "${KYMA_SOURCE}"
}

function upgrade_kyma() {
  echo "Upgrading Kyma to version: ${KYMA_UPGRADE_VERSION} using Execution profile: ${EXECUTION_PROFILE}"
  kyma deploy --ci --timeout 20m -p "$EXECUTION_PROFILE" --source "${KYMA_UPGRADE_VERSION}"
}

function run_pre_upgrade_tests() {
  echo "Executing pre-upgrade fast-integration tests"
  pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
  make ci-pre-upgrade
  popd
}

function run_post_upgrade_tests() {
  echo "Executing post-upgrade fast-integration tests"
  pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
  make ci-post-upgrade
  popd
}

prereq_test
load_env
install_reconciler_pr_cli
provision_k3d
deploy_kyma
run_pre_upgrade_tests
upgrade_kyma
run_post_upgrade_tests
