#!/usr/bin/env bash
LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${LIBDIR}/utils.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${LIBDIR}/gcloud.sh"

# gcloud::provision_gke_cluster creates a GKE cluster
# For switch parameters look up the code below.
#
# Required exported variables:
# GCLOUD_COMPUTE_ZONE - zone in which the new cluster will be created
# GCLOUD_PROJECT_NAME - name of GCP project
# GKE_CLUSTER_VERSION - GKE cluster version
#
# Arguments:
# $1 - cluster name
# $2 - optional additional labels for the cluster
function gcloud::provision_gke_cluster {
  if [ -z "$1" ]; then
    log::error "Cluster name not provided. Provide proper cluster name."
    exit 1
  fi
  CLUSTER_NAME=$1
  ADDITIONAL_LABELS=$2

  readonly CURRENT_TIMESTAMP_READABLE_PARAM=$(date +%Y%m%d)
  readonly CURRENT_TIMESTAMP_PARAM=$(date +%s)
  TTL_HOURS_PARAM="${TTL_HOURS:-"3"}"
  MACHINE_TYPE_PARAM="n1-standard-4"
  NUM_NODES_PARAM="3"
  NETWORK_PARAM="--network=default"
  local params

    # mandatory labels
    CLEANER_LABELS="created-at=${CURRENT_TIMESTAMP_PARAM},created-at-readable=${CURRENT_TIMESTAMP_READABLE_PARAM},ttl=${TTL_HOURS_PARAM}"
    JOB_LABELS="job=${JOB_NAME},job-id=${PROW_JOB_ID}"
    CLUSTER_LABELS="cluster=${CLUSTER_NAME},volatile=true"

    # optional labels
    if [ "${ADDITIONAL_LABELS}" ]; then
        ADDITIONAL_LABELS=",${ADDITIONAL_LABELS}"
    fi

    LABELS="$JOB_LABELS,"
    LABELS+="$CLUSTER_LABELS,"
    LABELS+="$CLEANER_LABELS"
    LABELS+="$ADDITIONAL_LABELS"
  # Resolving parameters
    params+=("--cluster-version=${GKE_CLUSTER_VERSION}")
    params+=("--machine-type=${MACHINE_TYPE:-$MACHINE_TYPE_PARAM}")
    params+=("--labels=$LABELS")

    if [ "${PROVISION_REGIONAL_CLUSTER}" ] && [ "${CLOUDSDK_COMPUTE_REGION}" ] && [ "${NODES_PER_ZONE}" ]; then
        params+=("--region=${CLOUDSDK_COMPUTE_REGION}")
        params+=("--num-nodes=${NODES_PER_ZONE}")
    else
        params+=("--zone=${GCLOUD_COMPUTE_ZONE}")
        params+=("--num-nodes=${NUM_NODES:-$NUM_NODES_PARAM}")
    fi
    if [ "${GKE_RELEASE_CHANNEL}" ]; then
        params+=("--release-channel=${GKE_RELEASE_CHANNEL}")
    fi
    if [ "${IMAGE_TYPE}" ]; then
        params+=("--image-type=${IMAGE_TYPE}")
    fi
    if [ "${GCLOUD_NETWORK_NAME}" ] && [ "${GCLOUD_SUBNET_NAME}" ]; then
        params+=("--network=${GCLOUD_NETWORK_NAME}" "--subnetwork=${GCLOUD_SUBNET_NAME}")
    else
        params+=("${NETWORK_PARAM}")
    fi
    if [ "${STACKDRIVER_KUBERNETES}" ]; then
        params+=("--enable-stackdriver-kubernetes")
    fi
    if [ "${CLUSTER_USE_SSD}" ]; then
        params+=("--disk-type=pd-ssd")
    fi
    if [ "${GCLOUD_SECURITY_GROUP_DOMAIN}" ]; then
        params+=("--security-group=gke-security-groups@${GCLOUD_SECURITY_GROUP_DOMAIN}")
    fi
    if [ "${GKE_ENABLE_POD_SECURITY_POLICY}" ]; then
        params+=("--enable-pod-security-policy")
    fi

  log::info "Provisioning GKE cluster"
  gcloud --project="$GCLOUD_PROJECT_NAME" beta container clusters create "$CLUSTER_NAME" "${params[@]}"
  log::info "Successfully created cluster $CLUSTER_NAME!"

  log::info "Patching kube-dns with stub domains"
  counter=0
  until [[ $(kubectl get cm kube-dns -n kube-system > /dev/null 2>&1; echo $?) == 0 ]]; do
      if (( counter == 5 )); then
          echo -e "kube-dns configmap not available after 5 tries, exiting"
          exit 1
      fi
      echo -e "Waiting for kube-dns to be available. Try $(( counter + 1 )) of 5"
      counter=$(( counter + 1 ))
      sleep 15
  done

  kubectl -n kube-system patch cm kube-dns --type merge --patch \
    "$(cat "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/resources/kube-dns-stub-domains-patch.yaml)"

  # Schedule pod with oom finder.
  if [ "${DEBUG_COMMANDO_OOM}" = "true" ]; then
      # run oom debug pod
      utils::debug_oom
  fi
}
