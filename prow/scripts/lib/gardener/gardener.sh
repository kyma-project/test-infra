#!/usr/bin/env bash

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"

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
  if [ -n "${CLEANUP_CLUSTER}" ]; then
    log::info "Deprovision cluster: \"${CLUSTER_NAME}\""
    GARDENER_PROJECT_NAME=$1
    GARDENER_CLUSTER_NAME=$2
    GARDENER_CREDENTIALS=$3

    local NAMESPACE="garden-${GARDENER_PROJECT_NAME}"

    kubectl --kubeconfig "${GARDENER_CREDENTIALS}" -n "${NAMESPACE}" annotate shoot "${GARDENER_CLUSTER_NAME}" confirmation.gardener.cloud/deletion=true --overwrite
    kubectl --kubeconfig "${GARDENER_CREDENTIALS}" -n "${NAMESPACE}" delete shoot "${GARDENER_CLUSTER_NAME}" --wait=false
  fi
}


# gardener::reprovision_cluster will generate new cluster name
# and start provisioning again
gardener::reprovision_cluster() {
  if [ "${reprovisionCount:-0}" -lt 1 ]; then
    log::info "cluster provisioning failed, trying provision new cluster"
    export reprovisionCount=1
    CLEANUP_CLUSTER="true" gardener::deprovision_cluster "${GARDENER_KYMA_PROW_PROJECT_NAME}" "${CLUSTER}" "${GARDENER_KYMA_PROW_KUBECONFIG}"
    utils::generate_commonName "${COMMON_NAME_PREFIX}"
    CLUSTER_NAME="${COMMON_NAME}"
    gardener::provision_cluster
  else
    log::info "cluster provisioning failed, already tried with new cluster, I give up"
  fi
}
