#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on k3s. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on a local k3d cluster.
#
#Expected common vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - KYMA_MAJOR_VERSION - major version of the first installation
#
#Please look in each provider script for provider specific requirements



function prereq() {
    # Unpack given envs 
    ENV_FILE=".env"
    if [ -f "${ENV_FILE}" ]; then
    # shellcheck disable=SC2046
    export $(xargs < "${ENV_FILE}")
    fi

    export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
    export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
    export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

    # shellcheck source=prow/scripts/lib/log.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
    # shellcheck source=prow/scripts/lib/utils.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
    # shellcheck source=prow/scripts/lib/kyma.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"

    # All provides require these values, each of them may check for additional variables
    requiredVars=(
        KYMA_PROJECT_DIR
        KYMA_MAJOR_VERSION
    )
    utils::check_required_vars "${requiredVars[@]}"

    # install kymaCLI from the last release
    kyma::install_cli_last_release
} 

function provision_cluster() {
    log::info "### Provision k3s cluster"
    kyma provision k3d --ci
}

function make_fast_integration() {
    log::info "### Run ${1} tests"

    pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
    git reset --hard "${KYMA_SOURCE}"
    make "${1}"
    popd

    log::success "Tests completed"
}

function install_kyma() {
    export KYMA_SOURCE=$(curl --silent --fail --show-error -H "Authorization: token $BOT_GITHUB_TOKEN" \
        "https://api.github.com/repos/kyma-project/kyma/releases" \
        | jq -r '[.[] | select(.tag_name | startswith("2."))] | first | .tag_name')
    log::info "### Reading release version from RELEASE_VERSION file, got: ${KYMA_SOURCE}"

    log::info "### Installing Kyma $KYMA_SOURCE"
    kyma deploy --ci --source "${KYMA_SOURCE}" --timeout 90m
}

function upgrade_kyma() {
    # Upgrade kyma to main
    export KYMA_SOURCE="main"

    log::info "### Upgrade Kyma to ${KYMA_SOURCE}"
    kyma deploy --ci --source "${KYMA_SOURCE}" --timeout 90m
}

# exit on error, handle right errors from tests
set -e

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD
ENABLE_TEST_LOG_COLLECTOR=false

prereq

log::info "### Starting pipeline"
provision_cluster

install_kyma

make_fast_integration "ci-pre-upgrade"

upgrade_kyma

make_fast_integration "ci-post-upgrade"

log::info "### waiting some time to finish cleanups"
sleep 60

make_fast_integration "ci-post-upgrade"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
