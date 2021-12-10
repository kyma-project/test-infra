#!/usr/bin/env bash

set -e
set -o pipefail

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

function delete_cluster_if_exists(){
  for i in {1..5}
  do
    local name="${INPUT_CLUSTER_NAME}${i}"
    set +e
    existing_shoot=$(kubectl get shoot "${name}" -ojsonpath="{ .metadata.name }")
    if [ -n "${existing_shoot}" ]; then
      log::info "Cluster found and deleting '${name}'"
      gardener::deprovision_cluster \
            -p "${GARDENER_KYMA_PROW_PROJECT_NAME}" \
            -c "${name}" \
            -f "${GARDENER_KYMA_PROW_KUBECONFIG}" \
            -w "true"

      log::info "We wait 120s for Gardener Shoot to settle after cluster deletion"
      sleep 120
    else
      log::info "Cluster '${name}' does not exist"
    fi
    set -e
  done
}

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export CONTROL_PLANE_RECONCILER_DIR="/home/prow/go/src/github.com/kyma-project/control-plane/tools/reconciler"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/gardener/gardener.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gardener.sh"
# shellcheck source=prow/scripts/cluster-integration/helpers/reconciler.sh
source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/reconciler.sh"


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

export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"

# Delete cluster with reconciler if exists
delete_cluster_if_exists

numeric_day=$(date +%u)
export INPUT_CLUSTER_NAME="${INPUT_CLUSTER_NAME}${numeric_day}"

# Provisioning gardener long lasting cluster
reconciler::provision_cluster

# Connect to the newly created shoot cluster
reconciler::connect_to_shoot_cluster

# Deploy reconciler
reconciler::deploy

# Disable sidecar injection for reconciler namespace
reconciler::disable_sidecar_injection_reconciler_ns

# Wait until reconciler is ready
reconciler::wait_until_is_ready
