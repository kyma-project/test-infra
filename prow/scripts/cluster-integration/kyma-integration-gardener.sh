#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on a real Gardener cluster.
#
#Expected vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - GARDENER_REGION - Gardener compute region
# - GARDENER_ZONES - Gardener compute zones inside the region
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME Name of the secret configured in the gardener project to access the cloud provider
# - MACHINE_TYPE (optional): machine type
#
#Provider specific:
#
#AWS:
#Permissions: In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
# - Compute Admin
# - Service Account User
# - Service Account Admin
# - Service Account Token Creator
# - Make sure the service account is enabled for the Google Identity and Access Management API.
#
#
#Azure:
# - RS_GROUP - azure resource group
# - REGION - azure region
# - AZURE_SUBSCRIPTION_ID
# - AZURE_SUBSCRIPTION_APP_ID
# - AZURE_SUBSCRIPTION_SECRET
# - AZURE_SUBSCRIPTION_TENANT
# - CLOUDSDK_CORE_PROJECT - required for cleanup of resources
#Permissions: In order to run this script you need to use an AKS service account with the contributor role
#
#
#GCP:
#Permissions: In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
# - Compute Admin
# - Service Account User
# - Service Account Admin
# - Service Account Token Creator
# - Make sure the service account is enabled for the Google Identity and Access Management API.


# exit on error, and raise error when variable is not set when used
set -eu

ENABLE_TEST_LOG_COLLECTOR=false

export GARDENER_CLUSTER_VERSION="1.16"

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/cluster-integration/helpers/kyma-cli.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/kyma-cli.sh"

# All provides require these values, each of them may check for additional variables
requiredVars=(
    GARDENER_PROVIDER
    KYMA_PROJECT_DIR
    GARDENER_REGION
    GARDENER_ZONES
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
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
else
    ## TODO what should I put here? Is this a backend?
    log::error "GARDENER_PROVIDER ${GARDENER_PROVIDER} is not yet supported"
    exit 1
fi

# nice cleanup on exit, be it succesful or on fail
trap gardener::cleanup EXIT INT

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
readonly COMMON_NAME_PREFIX="grd"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
export COMMON_NAME

### Cluster name must be less than 10 characters!
export CLUSTER_NAME="${COMMON_NAME}"

# set KYMA_SOURCE used by gardener::install_kyma
# at the time of writing this comment, kyma-integration-gardener never sets BUILD_TYPE to "release"
if [[ -n ${PULL_NUMBER} ]]; then
    # In case of PR, operate on PR number
    KYMA_SOURCE="PR-${PULL_NUMBER}"
    export KYMA_SOURCE
    # TODO maybe can be replaced with PULL_BASE_REF?
elif [[ "$BUILD_TYPE" == "release" ]]; then
    readonly RELEASE_VERSION=$(cat "${TEST_INFRA_SOURCES_DIR}/prow/RELEASE_VERSION")
    log::info "Reading release version from RELEASE_VERSION file, got: ${RELEASE_VERSION}"
    KYMA_SOURCE="${RELEASE_VERSION}"
    export KYMA_SOURCE
else
    # Otherwise (master), operate on triggering commit id
    if [[ -n ${PULL_BASE_SHA} ]]; then
        readonly COMMIT_ID="${PULL_BASE_SHA::8}"
        KYMA_SOURCE="${COMMIT_ID}"
        export KYMA_SOURCE
    else
        # periodic job, so default to master
        KYMA_SOURCE="master"
        export KYMA_SOURCE
    fi
fi

# checks required vars and initializes gcloud/docker if necessary
gardener::init

# if MACHINE_TYPE is not set then use default one
gardener::set_machine_type

install::kyma_cli

# currently only Azure generates overrides, but this may chaneg in the future
gardener::generate_overrides

log::info "Provision cluster: \"${CLUSTER_NAME}\""
gardener::provision_cluster

# uses previously set KYMA_SOURCE
gardener::install_kyma


if [[ "$HIBERNATION_ENABLED" == "true" ]]; then
    gardener::hibernate_kyma
    sleep 120
    gardener::wake_up_kyma
fi

if [[ "$EXECUTION_PROFILE" == "evaluation" ]]; then
    gardener::test_fast_integration_kyma
else
    # enable test-log-collector before tests; if prowjob fails before test phase we do not have any reason to enable it earlier
    if [[ "${BUILD_TYPE}" == "master" && -n "${LOG_COLLECTOR_SLACK_TOKEN}" ]]; then
      export ENABLE_TEST_LOG_COLLECTOR=true
    fi
    gardener::test_kyma
fi

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
