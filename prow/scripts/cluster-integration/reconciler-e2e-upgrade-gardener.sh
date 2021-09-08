#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Reconciler end-to-end upgrade flow on a real Gardener cluster.
#
#Expected common vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - GARDENER_REGION - Gardener compute region
# - GARDENER_ZONES - Gardener compute zones inside the region
# - GARDENER_CLUSTER_VERSION - Version of the Kubernetes cluster
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME - Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME - Name of the secret configured in the gardener project to access the cloud provider
# - MACHINE_TYPE - (optional) machine type
#
#Please look in each provider script for provider specific requirements

log::info "Starting pipeline..."

## ---------------------------------------------------------------------------------------
## Configurations and Variables
## ---------------------------------------------------------------------------------------

# exit on error, and raise error when variable is not set when used
set -e

ENABLE_TEST_LOG_COLLECTOR=false

# Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export RECONCILER_SOURCES_DIR="/home/prow/go/src/github.com/kyma-incubator/reconciler"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
## KYMA_SOURCE set to dummy value, required by gardener/gcp.sh
#export KYMA_SOURCE="main"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/gardener/gardener.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gardener.sh"

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

# Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD

readonly COMMON_NAME_PREFIX="grd"
utils::generate_commonName -n "${COMMON_NAME_PREFIX}"
COMMON_NAME="${utils_generate_commonName_return_commonName:?}"
export COMMON_NAME
export CLUSTER_NAME="${COMMON_NAME}"

# Get Kyma latest release version
kyma::get_last_release_version \
    -t "${BOT_GITHUB_TOKEN}"
LAST_RELEASE_VERSION="${kyma_get_last_release_version_return_version:?}"
log::info "### Reading release version from RELEASE_VERSION file, got: ${LAST_RELEASE_VERSION}"
KYMA_SOURCE="${LAST_RELEASE_VERSION}"

## ---------------------------------------------------------------------------------------
## Prow job execution steps
## ---------------------------------------------------------------------------------------

log::banner "Provisioning Gardener cluster"
# Checks required vars and initializes gcloud/docker if necessary
gardener::init

# If MACHINE_TYPE is not set then use default one
gardener::set_machine_type

# Install Kyma CLI
kyma::install_cli

# Provision garderner cluster
gardener::provision_cluster

# generate pod-security-policy list in json
utils::save_psp_list "${ARTIFACTS}/kyma-psp.json"

log::banner "Executing pre-upgrade tests"
gardener::pre_upgrade_test_fast_integration_kyma

### Once Kyma is installed run the fast integration test
#log::banner "Executing test"
#gardener::test_fast_integration_kyma
