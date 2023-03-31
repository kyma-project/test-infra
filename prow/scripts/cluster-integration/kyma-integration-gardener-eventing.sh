#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on a real Gardener cluster.
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

if [[ $GARDENER_PROVIDER == "azure" ]]; then
    # shellcheck source=prow/scripts/lib/gardener/azure.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/azure.sh"
elif [[ $GARDENER_PROVIDER == "aws" ]]; then
    # shellcheck source=prow/scripts/lib/gardener/aws.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/aws.sh"
elif [[ $GARDENER_PROVIDER == "gcp" ]]; then
    # shellcheck source=prow/scripts/lib/gardener/gcp.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gcp.sh"
    # shellcheck source=prow/scripts/lib/gardener/gardener.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gardener.sh"
else
    ## TODO what should I put here? Is this a backend?
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

    log::banner "Cleanup fast-integration assets"
    eventing::fast_integration_test_cleanup

    log::banner "Cleaning job assets"
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

export CLUSTER_NAME="${COMMON_NAME}"

# set KYMA_SOURCE used by gardener::deploy_kyma
utils::generate_vars_for_build \
    -b "$BUILD_TYPE" \
    -p "$PULL_NUMBER" \
    -s "$PULL_BASE_SHA" \
    -n "$JOB_NAME"
export KYMA_SOURCE=${utils_generate_vars_for_build_return_kymaSource:?}

## ---------------------------------------------------------------------------------------
## Prow job execution steps
## ---------------------------------------------------------------------------------------

# checks required vars and initializes gcloud/docker if necessary
gardener::init

# if MACHINE_TYPE is not set then use default one
gardener::set_machine_type

kyma::install_cli

# currently only Azure generates overrides, but this may change in the future
gardener::generate_overrides

export CLEANUP_CLUSTER="true"
gardener::provision_cluster

# deploy Kyma
log::info "Deploying Kyma ${KYMA_SOURCE}"
gardener::deploy_kyma -p "$EXECUTION_PROFILE" --source "${KYMA_SOURCE}"

# generate pod-security-policy list in json
utils::save_psp_list "${ARTIFACTS}/kyma-psp.json"


if [[ "${HIBERNATION_ENABLED}" == "true" ]]; then
    gardener::hibernate_kyma
    # TODO make the sleep value configurable (if it makes sense)
    sleep 120
    gardener::wake_up_kyma
fi

eventing::print_subscription_crd_version

if [[ "${EXECUTION_PROFILE}" == "evaluation" ]] || [[ "${EXECUTION_PROFILE}" == "production" ]]; then
    # test the default Eventing backend which comes with Kyma
    log::banner "Execute eventing E2E fast-integration tests"
    # eventing test assets cleanup will done later in cleanup script
    eventing::test_fast_integration_eventing_prep
    eventing::fast_integration_tests
else
    # enable test-log-collector before tests; if prowjob fails before test phase we do not have any reason to enable it earlier
    if [[ "${BUILD_TYPE}" == "master" && -n "${LOG_COLLECTOR_SLACK_TOKEN}" ]]; then
      export ENABLE_TEST_LOG_COLLECTOR=true
    fi
    gardener::test_kyma
fi


#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
