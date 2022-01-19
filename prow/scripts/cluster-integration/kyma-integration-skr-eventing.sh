#!/usr/bin/env bash

#Description: This scripts implements a pipeline that consists of many steps. The purpose is to provision and test Kyma eventing on SKR on a real Gardener cluster.
#
#Expected common vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
#
#Please look in each provider script for provider specific requirements

## ---------------------------------------------------------------------------------------
## Configurations and Variables
## ---------------------------------------------------------------------------------------

# exit on error, and raise error when variable is not set when used
set -e

ENABLE_TEST_LOG_COLLECTOR=false

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export EVENTMESH_SECRET_FILE="${CREDENTIALS_DIR}/serviceKey" # For eventing E2E fast-integration tests

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/cluster-integration/helpers/eventing.sh
source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/eventing.sh"

# All provides require these values, each of them may check for additional variables
requiredVars=(
    KYMA_PROJECT_DIR
    CREDENTIALS_DIR
    EVENTMESH_SECRET_FILE
)

utils::check_required_vars "${requiredVars[@]}"

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD

# set COMMON_NAME for cluster
RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
readonly COMMON_NAME_PREFIX="evnt"
export COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")

# set ENVs to be used by KEB to provision SKR
export INSTANCE_ID=$(cat /proc/sys/kernel/random/uuid) # SKR Runtime Id
export RUNTIME_NAME="${COMMON_NAME}"
export KYMA_VERSION="PR-${PULL_NUMBER}"
# shellcheck disable=SC2002
export KYMA_OVERRIDES_VERSION=$(cat "${KYMA_SOURCES_DIR}/tests/fast-integration/eventing-test/prow/config/skr_config.json" | jq -r '.kymaOverridesVersion')
export KYMA_TYPE=SKR

# Runs cleanup for the job
function skr::cleanup() {
    log::banner "De-provision SKR"
    eventing::test_fast_integration_deprovision_skr
}

## ---------------------------------------------------------------------------------------
## Prow job execution steps
## ---------------------------------------------------------------------------------------

# cleanup (De-provision SKR) hook on exit, either on successful or on fail
trap skr::cleanup EXIT INT

log::info "### Note: If the job fails to provision SKR, then verify the kymaOverridesVersion is correctly specified in kyma/tests/fast-integration/eventing-test/prow/config/skr_config.json for your PR"

log::banner "Provision SKR"
eventing::test_fast_integration_provision_skr

# set KUBECONFIG to ~/.kube/config
eventing::set_default_kubeconfig_env

log::banner "Execute eventing E2E fast-integration tests"
eventing::test_fast_integration_eventing

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
