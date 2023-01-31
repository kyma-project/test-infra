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
# - PREVIOUS_MINOR_VERSION_COUNT - Count of last Kyma2 minor versions to be upgraded from
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
    PREVIOUS_MINOR_VERSION_COUNT
)

function deploy_base() {
  #reverse array from oldest minor version -> latest minor version
  valid_minor_release_count="${#minor_release_versions[@]}"

  # base version
  export KYMA_SOURCE="${minor_release_versions[$((valid_minor_release_count -1))]}"
  log::info "### Installing Kyma $KYMA_SOURCE"

  # uses previously set KYMA_SOURCE
  kyma2_install_dir="$KYMA_SOURCES_DIR/$KYMA_SOURCE"
  kyma::deploy_kyma -s "$KYMA_SOURCE" -d "$KYMA_SOURCES_DIR" -u "true"

  log::info "### test directory: '$kyma2_install_dir/tests/fast-integration'"
  gardener::pre_upgrade_test_fast_integration_kyma -d "${kyma2_install_dir}/tests/fast-integration"
}

function upgrade() {
  for (( i=$((valid_minor_release_count -1)); i>0; i-- )) ; do
    #    e.g. for last 3 minor versions input (before reversal):
    #    0 - x.2.x
    #    1 - x.1.x
    #    2 - x.0.x
    log::info "### Testing upgrade from Kyma ${minor_release_versions[${i}]} to ${minor_release_versions[$((i - 1))]}"

    # upgrade version
    export KYMA_SOURCE="${minor_release_versions[$((i - 1))]}"
    log::info "### Installing Kyma $KYMA_SOURCE"

    kyma2_install_dir="$KYMA_SOURCES_DIR/$KYMA_SOURCE"
    kyma::deploy_kyma -s "$KYMA_SOURCE" -d "$KYMA_SOURCES_DIR" -u "true"

    log::info "### restart all functions in all namespaces to workaround https://github.com/kyma-project/kyma/issues/14757"
    kubectl delete pod -l=serverless.kyma-project.io/managed-by=function-controller -A
    
    # generate pod-security-policy list in json
    utils::save_psp_list "${ARTIFACTS}/kyma-psp.json"

    log::info "### Run post-upgrade tests"
    gardener::post_upgrade_test_fast_integration_kyma -d "${kyma2_install_dir}/tests/fast-integration"

    log::info "### waiting some time to finish cleanups"
    sleep 60
  done
}

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

# nice cleanup on exit, be it successful or on fail
trap gardener::cleanup EXIT INT

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
readonly COMMON_NAME_PREFIX="grd"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
export COMMON_NAME

### Cluster name must be less than 10 characters!
export CLUSTER_NAME="${COMMON_NAME}"

# set pipefail to handle right errors from tests
set -o pipefail

# Read latest release version
kyma::get_last_release_version -t "${BOT_GITHUB_TOKEN}"
export KYMA_SOURCE="${kyma_get_last_release_version_return_version:?}"
log::info "### Reading latest release version from RELEASE_VERSION file, got: ${KYMA_SOURCE}"

# Assess previous minor versions
declare -A minor_release_versions
kyma::get_offset_minor_releases -v "${KYMA_SOURCE}"

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

# deploy base minor version
deploy_base

# upgrade to next versions in a loop
upgrade

unset minor_release_versions

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"


