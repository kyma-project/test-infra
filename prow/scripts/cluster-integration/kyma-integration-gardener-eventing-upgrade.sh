#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps.
# The purpose is to install and test Kyma using the CLI on a real Gardener cluster, run the eventing fast integration tests, upgrade the cluster to the current PR version and the run the tests again
#
#Expected common vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - GARDENER_REGION - Gardener compute region
# - GARDENER_ZONES - Gardener compute zones inside the region
# - GARDENER_CLUSTER_VERSION - Version of the Kubernetes cluster
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME - Name of the gardener project where the cluster will be integrated
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME - Name of the secret configured in the gardener project to access the cloud provider
# - CREDENTIALS_DIR - Directory where is the EventMesh service key is mounted
# - MACHINE_TYPE - (optional) machine type
#
#Please look in each provider script for provider specific requirements

## ---------------------------------------------------------------------------------------
## Configurations and Variables
## ---------------------------------------------------------------------------------------

# exit on error, and raise error when variable is not set when used
set -o errexit

ENABLE_TEST_CLEANUP=false
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
# shellcheck source=prow/scripts/lib/gardener/gardener.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gardener.sh"

# All provides require these values, each of them may check for additional variables
requiredVars=(
    BOT_GITHUB_TOKEN
    GARDENER_PROVIDER
    KYMA_PROJECT_DIR
    GARDENER_REGION
    GARDENER_ZONES
    GARDENER_CLUSTER_VERSION
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
    CREDENTIALS_DIR
    EVENTMESH_SECRET_FILE
)

utils::check_required_vars "${requiredVars[@]}"
log::info "### Starting pipeline"

if [[ $GARDENER_PROVIDER == "azure" ]]; then
    # shellcheck source=prow/scripts/lib/gardener/azure.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/azure.sh"
elif [[ $GARDENER_PROVIDER == "aws" ]]; then
    # shellcheck source=prow/scripts/lib/gardener/aws.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/aws.sh"
elif [[ $GARDENER_PROVIDER == "gcp" ]]; then
    # shellcheck source=prow/scripts/lib/gardener/gcp.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gcp.sh"
else
    log::error "GARDENER_PROVIDER ${GARDENER_PROVIDER} is not yet supported"
    exit 1
fi

function cleanupJobAssets() {
    # Must be at the beginning
    EXIT_STATUS=$?

    set +e

    log::banner "Job Exit Status:: \"${EXIT_STATUS}\""

    if [[ $EXIT_STATUS != "0" ]]; then
        eventing::print_troubleshooting_logs
    fi

    log::banner "Cleaning job assets"
    if  [[ "${ENABLE_TEST_CLEANUP}" = true ]] ; then
        log::banner "Cleanup fast-integration assets"
        eventing::fast_integration_test_cleanup || log::info "Cleanup fast-integration assets failed"
    fi

    if  [[ "${CLEANUP_CLUSTER}" == "true" ]] ; then
        log::info "Deprovision cluster: \"${CLUSTER_NAME}\""
        gardener::deprovision_cluster \
            -p "${GARDENER_KYMA_PROW_PROJECT_NAME}" \
            -c "${CLUSTER_NAME}" \
            -f "${GARDENER_KYMA_PROW_KUBECONFIG}"
    fi

    set -e
    exit ${EXIT_STATUS}
}

# nice cleanup on exit, be it successful or on fail
trap cleanupJobAssets EXIT INT

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
readonly COMMON_NAME_PREFIX="grd"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
export COMMON_NAME

### Cluster name must be less than 10 characters!
export CLUSTER_NAME="${COMMON_NAME}"

# set pipefail to handle right errors from tests
set -o pipefail

# Install Kyma form latest release
kyma::get_last_release_version -t "${BOT_GITHUB_TOKEN}"

export KYMA_SOURCE="${kyma_get_last_release_version_return_version:?}"
log::info "### Reading release version from RELEASE_VERSION file, got: ${KYMA_SOURCE}"

# checks required vars and initializes gcloud/docker if necessary
gardener::init

# if MACHINE_TYPE is not set then use default one
gardener::set_machine_type

#TODO add an env flag to switch from provisioning using CLI to reconciler in local mode
kyma::install_unstable_cli

# currently only Azure generates overrides, but this may change in the future
gardener::generate_overrides

log::info "### Provisioning Gardener cluster"
export CLEANUP_CLUSTER="true"
gardener::provision_cluster

log::info "### Deploying Kyma $KYMA_SOURCE using $EXECUTION_PROFILE profile"
gardener::deploy_kyma --source "${KYMA_SOURCE}" -p "${EXECUTION_PROFILE}"

# generate pod-security-policy list in json
utils::save_psp_list "${ARTIFACTS}/kyma-psp.json"

# test the default Eventing backend which comes with Kyma
ENABLE_TEST_CLEANUP=true
eventing::pre_upgrade_test_fast_integration

# upgrade the kyma to the current PR/commit state
KYMA_SOURCE="PR-${PULL_NUMBER}"
export KYMA_SOURCE

# uses previously set KYMA_SOURCE
log::info "### Upgrading Kyma to $KYMA_SOURCE using $EXECUTION_PROFILE profile"
gardener::deploy_kyma --source "${KYMA_SOURCE}" -p "${EXECUTION_PROFILE}"

# test the eventing fi tests after the upgrade
eventing::fast_integration_tests

log::info "### Upgrading Kyma to $KYMA_SOURCE using $EXECUTION_PROFILE profile once again"
gardener::deploy_kyma --source "${KYMA_SOURCE}" -p "${EXECUTION_PROFILE}"

# test the eventing fi tests after the second upgrade and clean up
eventing::post_upgrade_test_fast_integration

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
