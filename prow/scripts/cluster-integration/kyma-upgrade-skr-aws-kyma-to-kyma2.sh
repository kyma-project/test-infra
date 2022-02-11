#!/usr/bin/env bash

#Description: Kyma Integration plan on SKR AWS. This scripts implements a pipeline that consists of many steps. The purpose is to trigger the ci-skr-kyma-to-kyma2-upgrade fast-integration test in Kyma repository
#Expected common vars:
#
#
#Please look in each provider script for provider specific requirements



set -o errexit
set -o pipefail

function prereq() {
    export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
    # shellcheck source=prow/scripts/lib/utils.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
    # shellcheck source=prow/scripts/lib/kyma.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"

    # All provides require these values, each of them may check for additional variables
    requiredVars=(
        KYMA_VERSION
    )
    utils::check_required_vars "${requiredVars[@]}"
}

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD

prereq

log::info "### Starting pipeline"

# Fetch latest Kyma2 release
kyma::get_last_release_version -t "${BOT_GITHUB_TOKEN}"
# KYMA_UPGRADE_VERSION will be used as a source in the fast-integration test
export KYMA_UPGRADE_VERSION="${kyma_get_last_release_version_return_version:?}"
log::info "### Reading release version from GitHub release API, got: ${KYMA_UPGRADE_VERSION}"

log::info "### Run make ci-skr-kyma-to-kyma2-upgrade"
make -C /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration ci-skr-kyma-to-kyma2-upgrade

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
