#!/usr/bin/env bash

#Description: decrypts files from gcloud
# The purpose of the script is to decrypt encrypted kyma files.
#
#Expected vars:
# - KYMA_KEYRING: kyma keyring name
# - KYMA_ENCRYPTION_KEY: encryption key name used to encrypt the files
# - TEST_INFRA_SOURCES_DIR: Path for log library
# - CLOUDSDK_KMS_PROJECT: Google Cloud Project ID with decryption key

set -o errexit

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"

readonly PLAIN_TEXT="$1"
if [ -p "$PLAIN_TEXT" ]; then
    log::error "Plaintext variable is missing!"
    exit 1
fi
readonly CIPHER_TEXT="$2"
if [ -c "$CIPHER_TEXT" ]; then
    log::error "Ciphertext variable is missing!"
    exit 1
fi

log::info "Decrypting ${CIPHER_TEXT} to  ${PLAIN_TEXT}"

gcloud kms decrypt --location global \
    --keyring "${KYMA_KEYRING}" \
    --key "${KYMA_ENCRYPTION_KEY}" \
    --ciphertext-file "${CIPHER_TEXT}" \
    --plaintext-file "${PLAIN_TEXT}" \
    --project "${CLOUDSDK_KMS_PROJECT}"