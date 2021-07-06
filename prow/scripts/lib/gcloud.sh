#!/usr/bin/env bash
LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${LIBDIR}/utils.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${LIBDIR}/kyma.sh"

# gcloud::verify_deps checks if the needed preconditions are met to use this library
function gcloud::verify_deps {
  commands=(
  gcloud
  )
  for cmd in "${commands[@]}"; do
    if ! [ -x "$(command -v "$cmd")" ]; then
      log::error "'$cmd' command not found in \$PATH. Exiting..."
    fi
  done
}


# gcloud::set_account activates already authenticated account
# Arguments:
# $1 - credentials to Google application
function gcloud::set_account() {
  if [[ -z $1 ]]; then
    log::error "Missing account credentials, please provide proper credentials"
    exit 1
  fi
  client_email=$(jq -r '.client_email' < "$1")
  log::info "Activating account $client_email"
  gcloud config set account "${client_email}" || exit 1
}

function gcloud::delete_ip_address {
  if [ -z "$1" ]; then
    log::error "IP address name is empty. Exiting..."
    exit 1
  fi

  IP_ADDRESS_NAME=$1
  log::info "Removing IP address $IP_ADDRESS_NAME."
  gcloud compute addresses delete "$IP_ADDRESS_NAME" --project="${CLOUDSDK_CORE_PROJECT}" --region="${CLOUDSDK_COMPUTE_REGION}"
  log::info "Successfully removed IP $IP_ADDRESS_NAME!"
}


# gcloud::delete_dns_record
# Required exported variables:
# CLOUDSDK_CORE_PROJECT - gcp project
# CLOUDSDK_COMPUTE_REGION - gcp region
#
# Arguments:
# $1 - ip address
# $2 - domain name
function gcloud::delete_dns_record {
  if [ -z "$1" ]; then
    log::error "IP address is empty. Exiting..."
    exit 1
  fi
  if [ -z "$2" ]; then
    log::error "Domain name is empty. Exiting..."
    exit 1
  fi
  ipAddress=$1
  dnsName=$2

  log::info "Deleting DNS record $DNS_FULL_NAME"
  set +e

  local attempts=10
  local retryTimeInSec="5"
  for ((i=1; i<=attempts; i++)); do
    gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction start --zone="${CLOUDSDK_DNS_ZONE_NAME}" && \
    gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction remove "${ipAddress}" --name="${dnsName}" --ttl=60 --type=A --zone="${CLOUDSDK_DNS_ZONE_NAME}" && \
    gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction execute --zone="${CLOUDSDK_DNS_ZONE_NAME}"

    if [[ $? -eq 0 ]]; then
      break
    fi

    gcloud dns record-sets transaction abort --zone="${CLOUDSDK_DNS_ZONE_NAME}" --verbosity none

    if [[ "${i}" -lt "${attempts}" ]]; then
      echo "Unable to delete DNS record, Retrying after $retryTimeInSec. Attempts ${i} of ${attempts}."
    else
      echo "Unable to delete DNS record after ${attempts} attempts, giving up."
      exit 1
    fi
    sleep ${retryTimeInSec}
  done

  log::info "DNS Record deleted, but it can be visible for some time due to DNS caches"
  set -e
}

# gcloud::encrypt encrypts text using Google KMS
# Required exported variables:
# CLOUDSDK_KMS_PROJECT
# KYMA_KEYRING
# KYMA_ENCRYPTION_KEY
#
# Arguments:
# $1 - plain text to encrypt
# $2 - cipher
function gcloud::encrypt {
  local PLAIN_TEXT="$1"
  if [ -p "$PLAIN_TEXT" ]; then
      log::error "Plaintext variable is missing! Exiting..."
      exit 1
  fi
  local CIPHER_TEXT="$2"
  if [ -c "$CIPHER_TEXT" ]; then
      log::error "Ciphertext variable is missing! Exiting..."
      exit 1
  fi

  log::info "Encrypting ${PLAIN_TEXT} as ${CIPHER_TEXT}"
  gcloud kms encrypt --location global \
      --keyring "${KYMA_KEYRING}" \
      --key "${KYMA_ENCRYPTION_KEY}" \
      --plaintext-file "${PLAIN_TEXT}" \
      --ciphertext-file "${CIPHER_TEXT}" \
      --project "${CLOUDSDK_KMS_PROJECT}"
}

# gcloud::encrypt encrypts text using Google KMS
# Required exported variables:
# CLOUDSDK_KMS_PROJECT
# KYMA_KEYRING
# KYMA_ENCRYPTION_KEY
#
# Arguments:
# $1 - plain text to encrypt
# $2 - cipher
function gcloud::decrypt {
  local PLAIN_TEXT="$1"
  if [ -p "$PLAIN_TEXT" ]; then
      echo "PLAIN_TEXT variable is missing!"
      exit 1
  fi
  local CIPHER_TEXT="$2"
  if [ -c "$CIPHER_TEXT" ]; then
      echo "CIPHER_TEXT variable is missing. Exiting..."
      exit 1
  fi

  log::info "Decrypting ${CIPHER_TEXT} to ${PLAIN_TEXT}"
  log::info "Decrypting ${CIPHER_TEXT} to ${PLAIN_TEXT}"

  gcloud kms decrypt --location global \
      --keyring "${KYMA_KEYRING}" \
      --key "${KYMA_ENCRYPTION_KEY}" \
      --ciphertext-file "${CIPHER_TEXT}" \
      --plaintext-file "${PLAIN_TEXT}" \
      --project "${CLOUDSDK_KMS_PROJECT}"
}


# gcloud::deprovision_gke_cluster removes a GKE cluster
# Required exported variables:
# GCLOUD_COMPUTE_ZONE - zone in which the new cluster will be removed
# GCLOUD_PROJECT_NAME - name of GCP project
#
# Arguments:
# $1 - cluster name
#TODO: remove after migration
function gcloud::deprovision_gke_cluster {
  if [ -z "$1" ]; then
    log::error "Cluster name not provided. Provide proper cluster name."
    exit 1
  fi
  CLUSTER_NAME=$1

#  gcloud config set project "${GCLOUD_PROJECT_NAME}"
#  gcloud config set compute/zone "${GCLOUD_COMPUTE_ZONE}"

  if [ "${DEBUG_COMMANDO_OOM}" = "true" ]; then
      # copy output from debug container to artifacts directory
      utils::oom_get_output
  fi

  local params
  params+=("--quiet")
  if [ "${PROVISION_REGIONAL_CLUSTER}" ] && [ "${CLOUDSDK_COMPUTE_REGION}" ]; then
    #Pass gke region name to delete command.
    params+=("--region=${CLOUDSDK_COMPUTE_REGION}")
  else
    params+=("--zone=${GCLOUD_COMPUTE_ZONE}")
  fi

  if [ -z "${DISABLE_ASYNC_DEPROVISION+x}" ]; then
      params+=("--async")
  fi

  log::info "Deprovisioning cluster $CLUSTER_NAME."
  gcloud --project="$GCLOUD_PROJECT_NAME" beta container clusters delete "$CLUSTER_NAME" "${params[@]}"
  log::info "Successfully removed cluster $CLUSTER_NAME!"
}

# gcloud::delete_docker_image deletes Docker image
# Arguments:
# $1 - name of the Docker image
function gcloud::delete_docker_image() {
  if [[ -z "$1" ]]; then
    log::error "Name of the Docker image to delete is missing, please provide proper name"
    exit 1
  fi
  gcloud container images delete "$1" || \
  (
    log::error "Could not remove Docker image" && \
    exit 1
  )
}

# gcloud::cleanup is a meta-function that removes all resources that were allocated for specific job.
# Required exported variables:
# CLOUDSDK_CORE_PROJECT
# CLOUDSDK_DNS_ZONE_NAME
# CLOUDSDK_COMPUTE_REGION
# GATEWAY_DNS_FULL_NAME
# GATEWAY_IP_ADDRESS
# APISERVER_DNS_FULL_NAME
# APISERVER_IP_ADDRESS
function gcloud::cleanup {
  utils::oom_get_output
  if [ -n "$CLEANUP_CLUSTER" ]; then
    log::info "Removing cluster $CLUSTER_NAME"
    gcloud::deprovision_gke_cluster "$CLUSTER_NAME"
  fi
  if [ -n "${CLEANUP_GATEWAY_DNS_RECORD}" ]; then
    log::info "Removing DNS record for $GATEWAY_DNS_FULL_NAME"
    gcloud::delete_dns_record "$GATEWAY_IP_ADDRESS" "$GATEWAY_DNS_FULL_NAME"
  fi
  if [ -n "${CLEANUP_GATEWAY_IP_ADDRESS}" ]; then
    log::info "Removing IP address $GATEWAY_IP_ADDRESS_NAME"
    gcloud::delete_ip_address "$GATEWAY_IP_ADDRESS_NAME"
  fi
  if [ -n "${CLEANUP_APISERVER_DNS_RECORD}" ]; then
    log::info "Removing DNS record for $APISERVER_DNS_FULL_NAME"
    gcloud::delete_dns_record "$APISERVER_IP_ADDRESS" "$APISERVER_DNS_FULL_NAME"
  fi
}

# gcloud::set_latest_cluster_version_for_channel checks for latest possible version in GKE_RELEASE_CHANNEL and updates GKE_CLUSTER_VERSION accordingly
# Required exported variables:
# GKE_RELEASE_CHANNEL
function gcloud::set_latest_cluster_version_for_channel() {
  if [ "${GKE_RELEASE_CHANNEL}" ]; then
    GKE_CLUSTER_VERSION=$(gcloud container get-server-config --zone europe-west4 --format="json" | jq -r '.channels|.[]|select(.channel | contains("'"${GKE_RELEASE_CHANNEL}"'"|ascii_upcase))|.validVersions|.[0]')
    log::info "Updating GKE_CLUSTER_VERSION to newest available in ${GKE_RELEASE_CHANNEL}: ${GKE_CLUSTER_VERSION}"
  fi
}

gcloud::verify_deps
