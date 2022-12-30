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
# - GARDENER_KYMA_PROW_PROJECT_NAME - Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME - Name of the secret configured in the gardener project to access the cloud provider
# - MACHINE_TYPE - (optional) machine type
#
#Please look in each provider script for provider specific requirements




# exit on error
set -o errexit

ENABLE_TEST_LOG_COLLECTOR=false

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
set -x
### Cluster name must be less than 10 characters!
export CLUSTER_NAME="${COMMON_NAME}"

# set pipefail to handle right errors from tests
set -o pipefail

utils::mask_debug_output
# Install kyma from latest 2.x release
kyma::get_last_release_version -t "${BOT_GITHUB_TOKEN}"
utils::unmask_debug_output

export KYMA_SOURCE="${kyma_get_last_release_version_return_version:?}"
log::info "### Reading release version from RELEASE_VERSION file, got: ${KYMA_SOURCE}"

# checks required vars and initializes gcloud/docker if necessary
gardener::init

# if MACHINE_TYPE is not set then use default one
gardener::set_machine_type

log::info "### Install latest unstable Kyma CLI"
kyma::install_unstable_cli

# currently only Azure generates overrides, but this may change in the future
gardener::generate_overrides

log::info "### Provisioning Gardener cluster"
export CLEANUP_CLUSTER="true"
gardener::provision_cluster

set +x

log::info "### Installing Kyma $KYMA_SOURCE"

kyma2_install_dir="$KYMA_SOURCES_DIR/kyma2"
kyma::deploy_kyma -s "$KYMA_SOURCE" -d "$kyma2_install_dir" -u "true"

# generate pod-security-policy list in json
utils::save_psp_list "${ARTIFACTS}/kyma-psp.json"

# Pre-Upgrade Tests
gardener::pre_upgrade_test_fast_integration_kyma -d "${kyma2_install_dir}/${KYMA_SOURCE}/tests/fast-integration"

# Upgrade kyma to main branch with latest stable cli
kyma::install_cli

export KYMA_SOURCE="main"
log::info "### Installing Kyma $KYMA_SOURCE"

kymaMain_install_dir="$KYMA_SOURCES_DIR/kymaMain"
kyma::deploy_kyma -s "$KYMA_SOURCE" -d "$kymaMain_install_dir" -u "true"

log::info "### Running post-upgrade Kyma Fast Integration tests"
pushd kymaMain/$KYMA_SOURCE/tests/fast-integration
make ci-post-upgrade
log::success "Tests completed"

log::info "### waiting some time to finish cleanups"
sleep 60

log::info "### Run pre-upgrade tests again to validate component removal"
gardener::pre_upgrade_test_fast_integration_kyma

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
