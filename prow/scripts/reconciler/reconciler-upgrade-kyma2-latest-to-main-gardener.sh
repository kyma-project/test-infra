#!/usr/bin/env bash

# Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps.
# The script does the following steps in order:
# 1. Provision a gardener cluster.
# 2. Deploy the Kyma reconciler from the control-plane pull-request.
# 3. Reconcile Kyma2 latest release using the deployed Kyma reconciler.
# 4. Execute pre-upgrade fast-integration tests.
# 5. Reconcile Kyma2 main using the deployed Kyma reconciler (to simulate Kyma2 version upgrade).
# 6. Execute post-upgrade fast-integration tests.
#
# Expected common vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - GARDENER_REGION - Gardener compute region
# - GARDENER_ZONES - Gardener compute zones inside the region
# - GARDENER_CLUSTER_VERSION - Version of the Kubernetes cluster
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME - Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME - Name of the secret configured in the gardener project to access the cloud provider
# - BOT_GITHUB_TOKEN: Bot github token used for API queries
# - MACHINE_TYPE - (optional) machine type
#
# Please look in each provider script for provider specific requirements.

## ---------------------------------------------------------------------------------------
## Configurations and Variables
## ---------------------------------------------------------------------------------------

# Exit on error, and raise error when variable is not set when used
set -e
pwd

# shellcheck source=prow/scripts/reconciler/common.sh
source "../../kyma-project/test-infra/prow/scripts/reconciler/common.sh"

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
)

utils::check_required_vars "${requiredVars[@]}"
# Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD

if  ! command -v helm &> /dev/null ; then
  echo "helm not found"
  mkdir -p bin
  HELM_VERSION=v3.12.1
  curl -Lo helm.tar.gz "https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz" && \
  tar -xzOf helm.tar.gz linux-amd64/helm > bin/helm && \
  chmod +x bin/helm
  export PATH="$PWD/bin:$PATH"
fi

## ---------------------------------------------------------------------------------------
## Provision Gardener cluster
## ---------------------------------------------------------------------------------------

# Generate cluster name
readonly COMMON_NAME_PREFIX="grd"
utils::generate_commonName -n "${COMMON_NAME_PREFIX}"

# Exported variables
export CLEANUP_CLUSTER="true"
export INPUT_CLUSTER_NAME="${utils_generate_commonName_return_commonName:?}"
export CLUSTER_NAME="${INPUT_CLUSTER_NAME}"

echo ">>> Provision Gardener cluster"
trap gardener::cleanup EXIT INT
reconciler::provision_cluster
reconciler::export_shoot_cluster_kubeconfig

## ---------------------------------------------------------------------------------------
## Deploy Kyma reconciler
## ---------------------------------------------------------------------------------------

# Exported variables
export CONTROL_PLANE_RECONCILER_DIR="/home/prow/go/src/github.com/kyma-project/control-plane/tools/reconciler"

# Deploy reconciler
reconciler::deploy

# Disable sidecar injection for reconciler namespace
reconciler::disable_sidecar_injection_reconciler_ns

# Wait until reconciler is ready
reconciler::wait_until_is_ready

## ---------------------------------------------------------------------------------------
## Reconcile and test Kyma2 latest release
## ---------------------------------------------------------------------------------------

# Get Kyma2 latest release version

pushd ../../kyma-project/kyma
git remote add origin https://github.com/kyma-project/kyma.git
git reset --hard && git remote update && git fetch --tags --all >/dev/null 2>&1
kyma_get_last_release_version_return_version=$(git tag -l '[0-9]*.[0-9]*.[0-9]*'| sort -r -V | grep '^[^-rc]*$'| head -n1)
export KYMA_UPGRADE_SOURCE="${kyma_get_last_release_version_return_version:?}"
git reset --hard && git checkout tags/"${KYMA_UPGRADE_SOURCE}"
popd

# Set up test pod environment
reconciler::deploy_test_pod
reconciler::wait_until_test_pod_is_ready
reconciler::initialize_test_pod

# Trigger the reconciliation through test pod
echo ">>> Reconcile Kyma2 version: ${KYMA_UPGRADE_SOURCE}"
reconciler::trigger_kyma_reconcile

# Wait until reconciliation is complete
reconciler::wait_until_kyma_reconciled

if [[ $KYMA_UPGRADE_SOURCE != "main" ]]; then
  kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-operator.yaml
  kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/default_serverless_cr.yaml -n kyma-system
fi
make -C "../../kyma-project/kyma/tests/fast-integration" ci-pre-upgrade

## ---------------------------------------------------------------------------------------
## Reconcile and test Kyma2 main
## ---------------------------------------------------------------------------------------

# Exported variables
export KYMA_UPGRADE_SOURCE="main"

# Set up test pod environment
reconciler::deploy_test_pod
reconciler::wait_until_test_pod_is_ready
reconciler::initialize_test_pod

# Trigger the reconciliation through test pod
echo ">>> Reconcile Kyma2 version: ${KYMA_UPGRADE_SOURCE}"
reconciler::trigger_kyma_reconcile

# Wait until reconciliation is complete
reconciler::wait_until_kyma_reconciled

# run the fast integration test after reconciliation
echo ">>> Executing post-upgrade test"
echo "switching local Kyma sources to the ${KYMA_UPGRADE_SOURCE}"
pushd "${KYMA_PROJECT_DIR}/kyma"
git reset --hard
git checkout "${KYMA_UPGRADE_SOURCE}"
popd

make -C "../../kyma-project/kyma/tests/fast-integration" ci-post-upgrade

# Must be at the end of the script
ERROR_LOGGING_GUARD="false"
