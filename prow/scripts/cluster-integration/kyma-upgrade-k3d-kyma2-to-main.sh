#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on k3s. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on a local k3d cluster.
#
#Expected common vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - KYMA_MAJOR_VERSION - major version of the first installation
#
#Please look in each provider script for provider specific requirements




# exit on error, and raise error when variable is not set when used
set -e

ENABLE_TEST_LOG_COLLECTOR=false

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
log::info "### Starting pipeline"

ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD

kyma::install_cli_last_release

log::info "### Provision k3s cluster"
kyma provision k3d --ci

# Install kyma from latest 2.x release
kyma::get_last_release_version -t "${BOT_GITHUB_TOKEN}"

export KYMA_SOURCE="${kyma_get_last_release_version_return_version:?}"
log::info "### Reading release version from RELEASE_VERSION file, got: ${KYMA_SOURCE}"

log::info "### Installing Kyma $KYMA_SOURCE"
# uses previously set KYMA_SOURCE
kyma deploy -ci --source "${KYMA_SOURCE}" --timeout 90m

# generate pod-security-policy list in json
utils::save_psp_list "${ARTIFACTS}/kyma-psp.json"

log::info "### Run pre-upgrade tests"

pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
make ci-pre-upgrade
popd

log::success "Tests completed"

export KYMA_SOURCE="main"
log::info "### Upgrade Kyma to ${KYMA_SOURCE}"
kyma deploy --ci --source "${KYMA_SOURCE}" --timeout 90m

log::info "### Run post-upgrade tests"

pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
make ci-post-upgrade
popd

log::success "Tests completed"

log::info "### waiting some time to finish cleanups"
sleep 60

log::info "### Run pre-upgrade tests again to validate component removal"

pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
make ci-post-upgrade
popd

log::success "Tests completed"

log::info "### Remove old components"
helm delete core -n kyma-system
helm delete console -n kyma-system
helm delete dex -n kyma-system
helm delete apiserver-proxy -n kyma-system
helm delete iam-kubeconfig-service -n kyma-system
helm delete testing -n kyma-system
helm delete xip-patch -n kyma-installer
helm delete permission-controller -n kyma-system

kubectl delete ns kyma-installer --ignore-not-found=true

log::info "### Run post-upgrade tests again to validate component removal"

pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
make ci-post-upgrade
popd

log::success "Tests completed"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
