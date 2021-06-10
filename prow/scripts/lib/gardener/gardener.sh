#!/usr/bin/env bash

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

function gardener::deprovision_cluster() {
  if [ -z "$1" ]; then
    echo "Project name is empty. Exiting..."
    exit 1
  fi
  if [ -z "$2" ]; then
    echo "Cluster name is empty. Exiting..."
    exit 1
  fi
  if [ -z "$3" ]; then
    echo "Kubeconfig path is empty. Exiting..."
    exit 1
  fi
  log::info "Deprovision cluster: ${CLUSTER_NAME}"
  GARDENER_PROJECT_NAME=$1
  GARDENER_CLUSTER_NAME=$2
  GARDENER_CREDENTIALS=$3

  local NAMESPACE="garden-${GARDENER_PROJECT_NAME}"

  kubectl annotate shoot "${GARDENER_CLUSTER_NAME}" confirmation.gardener.cloud/deletion=true \
    --overwrite \
    -n "${NAMESPACE}" \
    --kubeconfig "${GARDENER_CREDENTIALS}"
  kubectl delete shoot "${GARDENER_CLUSTER_NAME}" \
    --wait=false \
    --kubeconfig "${GARDENER_CREDENTIALS}" \
    -n "${NAMESPACE}"
}


# gardener::reprovision_cluster will generate new cluster name
# and start provisioning again
gardener::reprovision_cluster() {
    log::info "cluster provisioning failed, trying provision new cluster"
    log::info "cleaning damaged cluster first"
    gardener::deprovision_cluster "${GARDENER_KYMA_PROW_PROJECT_NAME}" "${CLUSTER_NAME}" "${GARDENER_KYMA_PROW_KUBECONFIG}"
    log::info "building new cluster name"
    utils::generate_commonName "${COMMON_NAME_PREFIX}"
    CLUSTER_NAME="${COMMON_NAME}"
    gardener::provision_cluster
}
