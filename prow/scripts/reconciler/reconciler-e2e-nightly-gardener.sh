#!/usr/bin/env bash

#Description: Reconciler E2E test plan on gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Reconciler end-to-end flow on a real Gardener cluster.
#Please look in each provider script for provider specific requirements

## ---------------------------------------------------------------------------------------
## Configurations and Variables
## ---------------------------------------------------------------------------------------

# exit on error, and raise error when variable is not set when used
set -e

pwd
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
    INPUT_CLUSTER_NAME
)

utils::check_required_vars "${requiredVars[@]}"

reconciler::export_nightly_cluster_name

echo "Connecting to nightly cluster"

reconciler::export_shoot_cluster_kubeconfig

reconciler::deploy_test_pod

reconciler::wait_until_test_pod_is_ready

reconciler::initialize_test_pod

reconciler::trigger_kyma_reconcile

reconciler::wait_until_kyma_reconciled

reconciler::break_kyma
