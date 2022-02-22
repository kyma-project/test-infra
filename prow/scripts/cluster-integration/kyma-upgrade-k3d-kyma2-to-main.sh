#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on k3s. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on a local k3d cluster.
#
#Expected common vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - KYMA_MAJOR_VERSION - major version of the first installation
#
#Please look in each provider script for provider specific requirements




set -o errexit
set -o pipefail

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

    git reset --hard "${KYMA_SOURCE}"
    make -C "${KYMA_SOURCES_DIR}/tests/fast-integration" "${1}"

    if [[ $? -eq 0 ]];then
        log::success "Tests completed"
    else
        log::error "Tests failed"
        exit 1
    fi
}

function install_kyma() {
    # Fetch latest Kyma2 release
    kyma::get_last_release_version -t "${BOT_GITHUB_TOKEN}"
    export KYMA_SOURCE="${kyma_get_last_release_version_return_version:?}"
    log::info "### Reading release version from RELEASE_VERSION file, got: ${KYMA_SOURCE}"

    # Install kyma from latest 2.x release
    log::info "### Installing Kyma $KYMA_SOURCE"
    kyma deploy --ci --source "${KYMA_SOURCE}" --timeout 60m

    if [[ $? -eq 0 ]];then
        log::success "Installation completed"
    else
        log::error "Installation failed"
        exit 1
    fi
}

function upgrade_kyma() {
    # Upgrade kyma to main
    export KYMA_SOURCE="main"

    log::info "### Upgrade Kyma to ${KYMA_SOURCE}"
    kyma deploy --ci --source "${KYMA_SOURCE}" --timeout 20m

    # fixes for upgrade
    kubectl patch service monitoring-alertmanager --type=json -p='[{"op": "remove", "path": "/spec/selector/app"}]'
    kubectl delete servicemonitors.monitoring.coreos.com monitoring-node-exporter

    if [[ $? -eq 0 ]];then
        log::success "Upgrade completed"
    else
        log::error "Upgrade failed"
        exit 1
    fi
}

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

make_fast_integration "ci-pre-upgrade"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
