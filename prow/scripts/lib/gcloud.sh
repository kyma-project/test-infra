#!/usr/bin/env bash
LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}"/log.sh

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

function gcloud::authenticate() {
    echo "Authenticating to gcloud"
    gcloud auth activate-service-account --key-file "${GOOGLE_APPLICATION_CREDENTIALS}" || exit 1
}

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

function gcloud::create_dns_record {
  set +e
  local attempts=10
  local retryTimeInSec="5"
  for ((i=1; i<=attempts; i++)); do
      gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction start --zone="${CLOUDSDK_DNS_ZONE_NAME}" && \
      gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction add "${IP_ADDRESS}" --name="${DNS_FULL_NAME}" --ttl=60 --type=A --zone="${CLOUDSDK_DNS_ZONE_NAME}" && \
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
      echo "Trying to resolve ${DNS_FULL_NAME}"
      sleep 10

      RESOLVED_IP_ADDRESS=$(dig +short "${DNS_FULL_NAME}")

      if [ "${RESOLVED_IP_ADDRESS}" = "${IP_ADDRESS}" ]; then
          echo "Successfully resolved ${DNS_FULL_NAME} to ${RESOLVED_IP_ADDRESS}!"
          return 0
      fi

      set +e
      log::info "Fetching DNS related debug logs and storing them in the artifacts folder..."
      {
        log::info "trace DNS response for ${DNS_FULL_NAME}"
        dig +trace "${DNS_FULL_NAME}"
        log::info "query authoritative servers directly"
        log::info "ns-cloud-b1.googledomains.com."
        dig "${DNS_FULL_NAME}" @ns-cloud-b1.googledomains.com.
        log::info "ns-cloud-b2.googledomains.com."
        dig "${DNS_FULL_NAME}" @ns-cloud-b2.googledomains.com.
        log::info "ns-cloud-b3.googledomains.com."
        dig "${DNS_FULL_NAME}" @ns-cloud-b3.googledomains.com.
        log::info "ns-cloud-b4.googledomains.com."
        dig "${DNS_FULL_NAME}" @ns-cloud-b4.googledomains.com.
        log::info "checking /etc/resolv.conf"
        cat /etc/resolv.conf
        log::info "checking kube-dns service IP"
        token=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
        curl -X GET -s --header "Authorization: Bearer $token" --insecure https://kubernetes.default.svc/api/v1/namespaces/kube-system/services?labelSelector=k8s-app=kube-dns | jq -r ".items[] | .spec.clusterIP"
        log::info "checking kube-dns endpoints addresses"
        endpoints=$(curl -X GET -s --header "Authorization: Bearer $token" --insecure https://kubernetes.default.svc/api/v1/namespaces/kube-system/endpoints?labelSelector=k8s-app=kube-dns | jq -r ".items[] | .subsets[] | .addresses[] | .ip")
        echo "$endpoints"
        log::info "query kube-dns pods directly"
        for srv in $endpoints; do log::info "querying $srv"; dig "${DNS_FULL_NAME}" @"$srv";done
      } >> "${ARTIFACTS}/dns-debug.txt"
      set -e
  done

  echo "Cannot resolve ${DNS_FULL_NAME} to expected IP_ADDRESS: ${IP_ADDRESS}."
  log::warn "WARNING! I will continue anyway... Kyma installation may fail!"
  #TODO: fix it
}

function gcloud::delete_dns_record {
  log::info "Deleting DNS record $DNS_FULL_NAME"
  set +e
  gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction start --zone="${CLOUDSDK_DNS_ZONE_NAME}" && \
  gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction remove "${IP_ADDRESS}" --name="${DNS_FULL_NAME}" --ttl=60 --type=A --zone="${CLOUDSDK_DNS_ZONE_NAME}" && \
  gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction execute --zone="${CLOUDSDK_DNS_ZONE_NAME}"
  log::info "DNS Record deleted, but it can be visible for some time due to DNS caches"
  set -e
}

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

function gcloud::decrypt {
  readonly PLAIN_TEXT="$1"
  if [ -p "$PLAIN_TEXT" ]; then
      echo "Plaintext variable is missing!"
      exit 1
  fi
  readonly CIPHER_TEXT="$2"
  if [ -c "$CIPHER_TEXT" ]; then
      echo "Ciphertext variable is missing!"
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

function gcloud::cleanup {
  true
}

gcloud::verify_deps
