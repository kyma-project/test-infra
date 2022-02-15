#!/usr/bin/env bash

set -o errexit
set -o pipefail

readonly RECONCILER_DIR="./reconciler"
readonly GO_VERSION=1.17.5
readonly PG_MIGRATE_VERSION=v4.15.1
readonly INSTALL_DIR="/usr/local/bin"

function prereq_test() {
  command -v node >/dev/null 2>&1 || { echo >&2 "node not found"; exit 1; }
  command -v npm >/dev/null 2>&1 || { echo >&2 "npm not found"; exit 1; }
  command -v jq >/dev/null 2>&1 || { echo >&2 "jq not found"; exit 1; }
  command -v helm >/dev/null 2>&1 || { echo >&2 "helm not found"; exit 1; }
  command -v kubectl >/dev/null 2>&1 || { echo >&2 "kubectl not found"; exit 1; }
  command -v k3d >/dev/null 2>&1 || { echo >&2 "k3d not found"; exit 1; }

  # All provides require these values, each of them may check for additional variables
  requiredVars=(
      KYMA_UPGRADE_SOURCE
      KYMA_PROJECT_DIR
  )
  utils::check_required_vars "${requiredVars[@]}"

  echo "KYMA_UPGRADE_SOURCE: ${KYMA_UPGRADE_SOURCE}"
  echo "KYMA_PROJECT_DIR: ${KYMA_PROJECT_DIR}"

  export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
  export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
  export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
  export CONTROL_PLANE_RECONCILER_DIR="${KYMA_PROJECT_DIR}/control-plane/tools/reconciler"

  # shellcheck source=prow/scripts/lib/log.sh
  source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
  # shellcheck source=prow/scripts/lib/utils.sh
  source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
  # shellcheck source=prow/scripts/lib/kyma.sh
  source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"

  # install kyma CLI from the last release
  kyma::install_cli_last_release
}

function provision_k3d() {
  kyma provision k3d --ci
}

function run_fast_integration() {
    log::info "### Run ${1} tests"

    git reset --hard "${KYMA_SOURCE}"
    make -C "${KYMA_SOURCES_DIR}/tests/fast-integration" "${1}"

    if [[ $? -eq 0 ]];then
        log::success "Tests completed"
    else
        log::error "Tests failed"
        exit 1
    fi
}

# Initialize pre-requisites
prereq_test

# Provision k3d cluster
provision_k3d

# Deploy reconciler
reconciler::deploy

# Wait until reconciler is ready
reconciler::wait_until_is_ready

# Deploy test pod which will trigger reconciliation
reconciler::deploy_test_pod

# Wait until test-pod is ready
reconciler::wait_until_test_pod_is_ready

# Set up test pod environment
reconciler::initialize_test_pod

# Run a test pod from where the reconciliation will be triggered
reconciler::reconcile_kyma

log::banner "Executing fast-integration tests"
make_fast_integration "ci"
