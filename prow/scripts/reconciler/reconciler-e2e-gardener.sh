#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Reconciler end-to-end flow on a real Gardener cluster.
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

## ---------------------------------------------------------------------------------------
## Configurations and Variables
## ---------------------------------------------------------------------------------------

# exit on error, and raise error when variable is not set when used
set -e
pwd

ENABLE_TEST_LOG_COLLECTOR=false

# Exported variables
# KYMA_SOURCE set to dummy value, required by gardener/gcp.sh
export KYMA_SOURCE="main"

# shellcheck source=prow/scripts/reconciler/common.sh
source "../../kyma-project/test-infra/prow/scripts/reconciler/common.sh"

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

trap gardener::cleanup EXIT INT

# Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD

readonly COMMON_NAME_PREFIX="grd"
utils::generate_commonName -n "${COMMON_NAME_PREFIX}"

export INPUT_CLUSTER_NAME="${utils_generate_commonName_return_commonName:?}"
# This is needed for the gardener::cleanup function
export CLUSTER_NAME="${INPUT_CLUSTER_NAME}"

# set Kyma version to reconcile
if [[ $KYMA_TEST_SOURCE == "latest-release" ]]; then
  # Fetch latest Kyma2 release

  pushd ../../kyma-project/kyma
  git remote add origin https://github.com/kyma-project/kyma.git
  git reset --hard && git remote update && git fetch --tags --all >/dev/null 2>&1
  kyma_get_last_release_version_return_version=$(git tag -l '[0-9]*.[0-9]*.[0-9]*'| sort -r -V | grep '^[^-rc]*$'| head -n1)
  export KYMA_UPGRADE_SOURCE="${kyma_get_last_release_version_return_version:?}"
  git reset --hard && git checkout tags/"${KYMA_UPGRADE_SOURCE}"
  popd
fi

## ---------------------------------------------------------------------------------------
## Prow job execution steps
## ---------------------------------------------------------------------------------------

echo ">>> Provisioning Gardener cluster"

# Provision garderner cluster
export CLEANUP_CLUSTER="true"
reconciler::provision_cluster

reconciler::export_shoot_cluster_kubeconfig

# Deploy reconciler
reconciler::deploy

# Wait until reconciler is ready
reconciler::wait_until_is_ready

# Deploy test pod which will trigger reconciliation
reconciler::deploy_test_pod

# Wait until test-pod is ready
reconciler::wait_until_test_pod_is_ready

# Set up test pod environment
reconciler::initialize_test_pod

# Trigger the reconciliation through test pod
reconciler::trigger_kyma_reconcile

# Wait until reconciliation is complete
reconciler::wait_until_kyma_reconciled

### Once Kyma is installed run the fast integration test
echo "Executing test"
if [[ $KYMA_TEST_SOURCE == "latest-release" ]]; then
  kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-operator.yaml
  kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/default_serverless_cr.yaml -n kyma-system
fi
make -C ../../kyma-project/kyma/tests/fast-integration ci

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"

gardener::cleanup
