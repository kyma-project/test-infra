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
