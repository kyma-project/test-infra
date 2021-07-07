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

gcloud::verify_deps
