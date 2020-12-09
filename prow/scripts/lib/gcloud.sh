#!/usr/bin/env bash
LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}"/log.sh

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
# Required exported variables:
# GOOGLE_APPLICATION_CREDENTIALS - google login credentials
function gcloud::authenticate() {
    echo "Authenticating to gcloud"
    gcloud auth activate-service-account --key-file "${GOOGLE_APPLICATION_CREDENTIALS}" || exit 1
}

# gcloud::reserve_ip_address requests a new IP address from gcloud and prints this value to STDOUT
# Required exported variables:
# CLOUDSDK_CORE_PROJECT - gcp project
# CLOUDSDK_COMPUTE_REGION - gcp region
# IP_ADDRESS_NAME - name of the IP address to be set in gcp
function gcloud::reserve_ip_address {
  counter=0
  # Check if IP address reservation is present. Wait and retry for one minute to disappear. If IP reservation was removed just before it need few seconds to disappear.
  # Otherwise, creation will fail.
  IP_ADDRESS=$(gcloud compute addresses list --filter="name=${IP_ADDRESS_NAME}" --format="value(ADDRESS)")
  until [[ -z ${IP_ADDRESS} ]]; do
    sleep 15
    counter=$(( counter + 1 ))
    IP_ADDRESS=$(gcloud compute addresses list --filter="name=${IP_ADDRESS_NAME}" --format="value(ADDRESS)")
    if (( counter == 5 )); then
      # Fail after one minute wait.
      echo "${IP_ADDRESS_NAME} IP address is still present after one minute wait. Failing"
      return 1
    fi
  done

  gcloud compute addresses create "${IP_ADDRESS_NAME}" --project="${CLOUDSDK_CORE_PROJECT}" --region="${CLOUDSDK_COMPUTE_REGION}" --network-tier=PREMIUM
  # Print reserved IP address on stdout as it's consumed by calling process and used for next steps.
  gcloud compute addresses list --filter="name=${IP_ADDRESS_NAME}" --format="value(ADDRESS)"
}

function gcloud::delete_ip_address {
  # TODO (@Ressetkk): write some implementation that allows to safely remove IP address.
  true
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

      set +e
      log::info "Fetching DNS related debug logs and storing them in the artifacts folder..."
      {
        log::info "trace DNS response for ${dnsName}"
        dig +trace "${dnsName}"
        log::info "query authoritative servers directly"
        log::info "ns-cloud-b1.googledomains.com."
        dig "${dnsName}" @ns-cloud-b1.googledomains.com.
        log::info "ns-cloud-b2.googledomains.com."
        dig "${dnsName}" @ns-cloud-b2.googledomains.com.
        log::info "ns-cloud-b3.googledomains.com."
        dig "${dnsName}" @ns-cloud-b3.googledomains.com.
        log::info "ns-cloud-b4.googledomains.com."
        dig "${dnsName}" @ns-cloud-b4.googledomains.com.
        log::info "checking /etc/resolv.conf"
        cat /etc/resolv.conf
        log::info "checking kube-dns service IP"
        token=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
        curl -X GET -s --header "Authorization: Bearer $token" --insecure https://kubernetes.default.svc/api/v1/namespaces/kube-system/services?labelSelector=k8s-app=kube-dns | jq -r ".items[] | .spec.clusterIP"
        log::info "checking kube-dns endpoints addresses"
        endpoints=$(curl -X GET -s --header "Authorization: Bearer $token" --insecure https://kubernetes.default.svc/api/v1/namespaces/kube-system/endpoints?labelSelector=k8s-app=kube-dns | jq -r ".items[] | .subsets[] | .addresses[] | .ip")
        echo "$endpoints"
        log::info "query kube-dns pods directly"
        for srv in $endpoints; do log::info "querying $srv"; dig "${dnsName}" @"$srv";done
      } >> "${ARTIFACTS}/dns-debug.txt"
      set -e
  done

  echo "Cannot resolve ${dnsName} to expected IP_ADDRESS: ${ipAddress}."
  log::warn "WARNING! I will continue anyway... Kyma installation may fail!"
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
  gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction start --zone="${CLOUDSDK_DNS_ZONE_NAME}" && \
  gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction remove "${ipAddress}" --name="${dnsName}" --ttl=60 --type=A --zone="${CLOUDSDK_DNS_ZONE_NAME}" && \
  gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction execute --zone="${CLOUDSDK_DNS_ZONE_NAME}"
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
  readonly PLAIN_TEXT="$1"
  if [ -p "$PLAIN_TEXT" ]; then
      log::error "Plaintext variable is missing! Exiting..."
      exit 1
  fi
  readonly CIPHER_TEXT="$2"
  if [ -c "$CIPHER_TEXT" ]; then
      log::error "Ciphertext variable is missing! Exiting..."
      exit 1
  fi

  shout "Encrypting ${PLAIN_TEXT} as ${CIPHER_TEXT}"
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
  readonly PLAIN_TEXT="$1"
  if [ -p "$PLAIN_TEXT" ]; then
      echo "PLAIN_TEXT variable is missing!"
      exit 1
  fi
  readonly CIPHER_TEXT="$2"
  if [ -c "$CIPHER_TEXT" ]; then
      echo "CIPHER_TEXT variable is missing. Exiting..."
      exit 1
  fi

  shout "Decrypting ${CIPHER_TEXT} to ${PLAIN_TEXT}"

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

   log::info "successfully created network $GCLOUD_NETWORK_NAME"
}

function gcloud::cleanup {
  true
}

gcloud::verify_deps
