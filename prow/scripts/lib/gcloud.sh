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

# gcloud::authenticate authenticates to gcloud.
# Arguments:
# $1 - google login credentials
function gcloud::authenticate() {
    if [[ -z "$1" ]]; then
      log::error "Missing account credentials, please provide proper credentials"
    fi
    log::info "Authenticating to gcloud"
    gcloud auth activate-service-account --key-file "${1}" || exit 1
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

# gcloud::reserve_ip_address requests a new IP address from gcloud and prints this value to STDOUT
# Required exported variables:
# CLOUDSDK_CORE_PROJECT - gcp project
# CLOUDSDK_COMPUTE_REGION - gcp region
# Arguments:
# $1 - name of the IP address to be set in gcp
# Returns:
# gcloud::reserve_ip_address_return_1 - reserved ip address
# TODO: add support for setting CLOUDSDK env vars from function args.
function gcloud::reserve_ip_address {
    utils::check_empty_arg "${1}" "IP address name is empty. Exiting..."
    local ipAddressName=$1
    log::info "Reserve IP Address for ${ipAddressName}"
    # TODO: setting this variable should be done outside function, it's to specific
    export CLEANUP_GATEWAY_IP_ADDRESS="true"
    local counter=0
    # Check if IP address reservation is present. Wait and retry for one minute to disappear.
    # If IP reservation was removed just before it need a few seconds to disappear.
    # Otherwise, creation will fail.
    local ipAddress
    ipAddress=$(gcloud compute addresses list --filter="name=${ipAddressName}" --format="value(ADDRESS)")
  until [[ -z ${ipAddress} ]]; do
    sleep 15
    counter=$(( counter + 1 ))
    ipAddress=$(gcloud compute addresses list --filter="name=${ipAddressName}" --format="value(ADDRESS)")
    if (( counter == 5 )); then
      # Fail after one minute wait.
      echo "${ipAddressName} IP address is still present after one minute wait. Failing"
      return 1
    fi
  done

  gcloud compute addresses create "${ipAddressName}" --project="${CLOUDSDK_CORE_PROJECT}" --region="${CLOUDSDK_COMPUTE_REGION}" --network-tier="PREMIUM"
  # Print reserved IP address on stdout as it's consumed by calling process and used for next steps.
  # TODO: export result as variable, change consuming command to use exported variable
  #gcloud compute addresses list --filter="name=${IP_ADDRESS_NAME}" --format="value(ADDRESS)"
  gcloud::reserve_ip_address_return_1="$(gcloud compute addresses list --filter="name=${ipAddressName}" --format="value(ADDRESS)")"
  log::info "Created IP Address for Ingressgateway: ${ipAddressName}"
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

# gcloud::create_dns_record creates an A dns record for corresponding ip address
# Required exported variables:
# CLOUDSDK_CORE_PROJECT - gcp project
# CLOUDSDK_COMPUTE_REGION - gcp region
#
# Arguments:
# $1 - ip address
# $2 - domain name
function gcloud::create_dns_record {
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

  set +e
  local attempts=10
  local retryTimeInSec="5"
  for ((i=1; i<=attempts; i++)); do
      gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction start --zone="${CLOUDSDK_DNS_ZONE_NAME}" && \
      gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction add "${ipAddress}" --name="${dnsName}" --ttl=60 --type=A --zone="${CLOUDSDK_DNS_ZONE_NAME}" && \
      gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction execute --zone="${CLOUDSDK_DNS_ZONE_NAME}"

      if [[ $? -eq 0 ]]; then
          break
      fi

      gcloud dns record-sets transaction abort --zone="${CLOUDSDK_DNS_ZONE_NAME}" --verbosity none

      if [[ "${i}" -lt "${attempts}" ]]; then
          echo "Unable to create DNS record, let's wait ${retryTimeInSec} seconds and retry. Attempts ${i} of ${attempts}."
      else
          echo "Unable to create DNS record after ${attempts} attempts, giving up."
          exit 1
      fi

      sleep ${retryTimeInSec}
  done

  set -e

  local SECONDS=0
  local END_TIME=$((SECONDS+600)) #600 seconds == 10 minutes

  while [ ${SECONDS} -lt ${END_TIME} ];do
      echo "Trying to resolve ${dnsName}"
      sleep 10

      RESOLVED_IP_ADDRESS=$(dig +short "${dnsName}")

      if [ "${RESOLVED_IP_ADDRESS}" = "${ipAddress}" ]; then
          echo "Successfully resolved ${dnsName} to ${RESOLVED_IP_ADDRESS}!"
          return 0
      fi
  done

  echo "Cannot resolve ${dnsName} to expected IP_ADDRESS: ${ipAddress}."
  log::warn "Continuing anyway... Kyma installation may fail!"
  #TODO: fix it
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

# gcloud::create_network creates a GCP network for a cluster
# Required exported variables:
# GCLOUD_NETWORK_NAME - name for the new GCP network
# GCLOUD_SUBNET_NAME - name for the subnet of the network
# GCLOUD_PROJECT_NAME - name of GCP project
function gcloud::create_network {
  if [ -z "$1" ]; then
    log::error "Network name is empty. Exiting..."
    exit 1
  fi
  if [ -z "$2" ]; then
    log::error "Subnet name is empty. Exiting..."
    exit 1
  fi
  GCLOUD_NETWORK_NAME=$1
  GCLOUD_SUBNET_NAME=$2

  if gcloud compute networks describe "$GCLOUD_NETWORK_NAME"; then
    log::warn "Network $GCLOUD_NETWORK_NAME already exists! Skipping network creation."
    return 0
  fi
  log::info "Creating network $GCLOUD_NETWORK_NAME"
  gcloud compute networks create "${GCLOUD_NETWORK_NAME}" \
 --project="${GCLOUD_PROJECT_NAME}" \
 --subnet-mode=custom

  gcloud compute networks subnets create "${GCLOUD_SUBNET_NAME}" \
   --network="${GCLOUD_NETWORK_NAME}" \
   --range=10.0.0.0/22

   log::info "Successfully created network $GCLOUD_NETWORK_NAME"
}

# gcloud::delete_network deletes a GCP network for a cluster
# Required exported variables:
# GCLOUD_NETWORK_NAME - name for the new GCP network
# GCLOUD_SUBNET_NAME - name for the subnet of the network
# GCLOUD_PROJECT_NAME - name of GCP project
function gcloud::delete_network {
  log::info "Deleting network $GCLOUD_NETWORK_NAME"
  gcloud compute networks subnets delete "${GCLOUD_SUBNET_NAME}" \
    --quiet

  gcloud compute networks delete "${GCLOUD_NETWORK_NAME}" \
    --project="${GCLOUD_PROJECT_NAME}" \
    --quiet

  log::info "Successfully deleted network $GCLOUD_NETWORK_NAME"
}

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
#TODO: remove after migration
function gcloud::provision_gke_cluster {
  if [ -z "$1" ]; then
    log::error "Cluster name not provided. Provide proper cluster name."
    exit 1
  fi
  CLUSTER_NAME=$1
  ADDITIONAL_LABELS=$2

  readonly CURRENT_TIMESTAMP_READABLE_PARAM=$(date +%Y%m%d)
  readonly CURRENT_TIMESTAMP_PARAM=$(date +%s)
  TTL_HOURS_PARAM="3"
  MACHINE_TYPE_PARAM="n1-standard-4"
  NUM_NODES_PARAM="3"
  NETWORK_PARAM="--network=default"

  local params

  if [ "${TTL_HOURS}" ]; then TTL_HOURS_PARAM="${TTL_HOURS}"; fi
  CLEANER_LABELS_PARAM="created-at=${CURRENT_TIMESTAMP_PARAM},created-at-readable=${CURRENT_TIMESTAMP_READABLE_PARAM},ttl=${TTL_HOURS_PARAM}"

#  gcloud config set project "$GCLOUD_PROJECT_NAME"
#  gcloud config set compute/zone "${GCLOUD_COMPUTE_ZONE}"
  # Resolving parameters
  params+=("--cluster-version=${GKE_CLUSTER_VERSION}")
  if [ "${GKE_RELEASE_CHANNEL}" ]; then params+=("--release-channel=${GKE_RELEASE_CHANNEL}"); fi
  params+=("--machine-type=${MACHINE_TYPE:-$MACHINE_TYPE_PARAM}")
  if [ "${IMAGE_TYPE}" ]; then params+=("--image-type=${IMAGE_TYPE}"); fi
  params+=("--num-nodes=${NUM_NODES:-$NUM_NODES_PARAM}")
  if [ "${GCLOUD_NETWORK_NAME}" ] && [ "${GCLOUD_SUBNET_NAME}" ]; then params+=("--network=${GCLOUD_NETWORK_NAME}" "--subnetwork=${GCLOUD_SUBNET_NAME}"); else params+=("${NETWORK_PARAM}"); fi
  if [ "${STACKDRIVER_KUBERNETES}" ];then params+=("--enable-stackdriver-kubernetes"); fi
  if [ "${CLUSTER_USE_SSD}" ];then params+=("--disk-type=pd-ssd"); fi
  # Must be added at the end to override --num-nodes parameter value set earlier.
  if [ "${PROVISION_REGIONAL_CLUSTER}" ] && [ "${CLOUDSDK_COMPUTE_REGION}" ] && [ "${NODES_PER_ZONE}" ];then
    params+=("--region=${CLOUDSDK_COMPUTE_REGION}" "--num-nodes=${NODES_PER_ZONE}")
  else
    params+=("--zone=${GCLOUD_COMPUTE_ZONE}")
  fi
  if [ "${GCLOUD_SECURITY_GROUP_DOMAIN}" ]; then params+=("--security-group=gke-security-groups@${GCLOUD_SECURITY_GROUP_DOMAIN}"); fi
  if [ "${GKE_ENABLE_POD_SECURITY_POLICY}" ]; then params+=("--enable-pod-security-policy"); fi

  APPENDED_LABELS=()
  if [ "${ADDITIONAL_LABELS}" ]; then APPENDED_LABELS=(",${ADDITIONAL_LABELS}") ; fi
  params+=("--labels=job=${JOB_NAME},job-id=${PROW_JOB_ID},cluster=${CLUSTER_NAME},volatile=true${APPENDED_LABELS[@]},${CLEANER_LABELS_PARAM}")

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
