#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/../log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${LIBDIR}/../utils.sh"

# gardener::deprovision_cluster removes a Gardener cluster
#
# Arguments:
#
# required:
# p - project name
# c - lcuster name
# f - kubeconfig file path
function gardener::deprovision_cluster() {
  local OPTIND
  local projectName
  local clusterName
  local kubeconfigFile
  local namespace
  local wait="false"

  while getopts ":p:c:f:w:" opt; do
      case $opt in
          p)
            projectName="$OPTARG" ;;
          c)
            clusterName="$OPTARG" ;;
          f)
            kubeconfigFile="$OPTARG" ;;
          w)
            wait=${OPTARG:-$wait} ;;
          \?)
              echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
          :)
              echo "Option -$OPTARG argument not provided" >&2 ;;
      esac
  done


  utils::check_empty_arg "$projectName" "Project name is empty. Exiting..."
  utils::check_empty_arg "$clusterName" "Cluster name is empty. Exiting..."
  utils::check_empty_arg "$kubeconfigFile" "Kubeconfig file path is empty. Exiting..."

  log::info "Deprovision cluster: ${clusterName}"

  namespace="garden-${projectName}"

  kubectl annotate shoot "${clusterName}" confirmation.gardener.cloud/deletion=true \
    --overwrite \
    -n "${namespace}" \
    --kubeconfig "${kubeconfigFile}"
  kubectl delete shoot "${clusterName}" \
    --wait="${wait}" \
    --kubeconfig "${kubeconfigFile}" \
    -n "${namespace}"
}


# gardener::reprovision_cluster will generate new cluster name
# and start provisioning again
gardener::reprovision_cluster() {
    # Save bash options to restore them later
    bashSettings="$-"
    # disable pipefile to let function regenerate cluster name
    set +o pipefile
    log::info "cluster provisioning failed, trying provision new cluster"
    log::info "cleaning damaged cluster first"
    gardener::deprovision_cluster \
      -p "${GARDENER_KYMA_PROW_PROJECT_NAME}" \
      -c "${CLUSTER_NAME}" \
      -f "${GARDENER_KYMA_PROW_KUBECONFIG}"
    log::info "building new cluster name"
    utils::generate_commonName -n "${COMMON_NAME_PREFIX}"
    COMMON_NAME=${utils_generate_commonName_return_commonName:?}
    export COMMON_NAME
    CLUSTER_NAME="${COMMON_NAME}"
    export CLUSTER_NAME
    gardener::provision_cluster
    # restore bash options
    set -"${bashSettings}"
}
