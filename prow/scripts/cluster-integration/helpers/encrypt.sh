#!/usr/bin/env bash

#Description: encrypts files from gcloud
# The purpose of the script is to encrypt plaintext files for kyma.
#
#Expected vars:
# - KYMA_KEYRING: kyma keyring name
# - KYMA_ENCRYPTION_KEY: encryption key name used to encrypt the files
# - TEST_INFRA_SOURCES_DIR: Path for log library
# - CLOUDSDK_KMS_PROJECT: Google Cloud Project ID with encryption key

set -o errexit

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"


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


log::info "Encrypting ${PLAIN_TEXT} as  ${CIPHER_TEXT}"
gcloud kms encrypt --location global \
    --keyring "${KYMA_KEYRING}" \
    --key "${KYMA_ENCRYPTION_KEY}" \
    --plaintext-file "${PLAIN_TEXT}" \
    --ciphertext-file "${CIPHER_TEXT}" \
    --project "${CLOUDSDK_KMS_PROJECT}"

