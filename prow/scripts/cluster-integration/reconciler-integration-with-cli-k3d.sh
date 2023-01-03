#!/usr/bin/env bash

set -o errexit
set -o pipefail

function prereq_init() {
  export KYMA_SOURCES_DIR="./kyma"
  export HOME_DIR="$PWD"
  export TEST_INFRA_SOURCES_DIR="${HOME_DIR}/test-infra"

  command -v node >/dev/null 2>&1 || { echo >&2 "node not found"; exit 1; }
  command -v npm >/dev/null 2>&1 || { echo >&2 "npm not found"; exit 1; }
  command -v jq >/dev/null 2>&1 || { echo >&2 "jq not found"; exit 1; }
  command -v helm >/dev/null 2>&1 || { echo >&2 "helm not found"; exit 1; }
  command -v kubectl >/dev/null 2>&1 || { echo >&2 "kubectl not found"; exit 1; }
  command -v k3d >/dev/null 2>&1 || { echo >&2 "k3d not found"; exit 1; }

  # shellcheck source=prow/scripts/lib/log.sh
  source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
  # shellcheck source=prow/scripts/lib/utils.sh
  source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
  # shellcheck source=prow/scripts/lib/kyma.sh
  source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
}

function load_env() {
  ENV_FILE=".env"
  if [ -f "${ENV_FILE}" ]; then
    # shellcheck disable=SC2046
    export $(xargs < "${ENV_FILE}")
  fi
}

function run_tests() {
  pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"

  log::info "KYMA_SOURCE ${KYMA_SOURCE}"
  git reset --hard
  if [[ ${KYMA_SOURCE} == "main" ]]
  then
    git checkout "${KYMA_SOURCE}"
  else
    git checkout tags/"${KYMA_SOURCE}"
  fi

  make ci
  popd
}

## ******* Execution Steps ********
# Initialize pre-requisites
prereq_init

# Load env file as environment variables
load_env

log::banner "Installing Kyma CLI from reconciler PR-${PULL_NUMBER}"
kyma::install_cli_from_reconciler_pr

log::banner "Provisioning K3d cluster"
kyma::provision_k3d

log::banner "Deploying Kyma version: ${KYMA_SOURCE} using Execution profile: ${EXECUTION_PROFILE}"
kyma::deploy_kyma -s "${KYMA_SOURCE}" -p "${EXECUTION_PROFILE}"

log::banner "Executing fast-integration tests"
run_tests
