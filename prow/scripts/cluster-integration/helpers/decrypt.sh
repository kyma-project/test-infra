#!/usr/bin/env bash

#Description: decrypts files from gcloud
# The purpose of the script is to decrypt encrypted kyma files.
#
#Expected vars:
# - KYMA_KEYRING: kyma keyring name
# - KYMA_ENCRYPTION_KEY: encryption key name used to encrypt the files
# - TEST_INFRA_SOURCES_DIR: Path for shout library
# - CLOUDSDK_CORE_PROJECT: Google Cloud Project ID with decryption key

set -o errexit

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"


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

shout "Decrypting ${CIPHER_TEXT} to  ${PLAIN_TEXT}"

gcloud kms decrypt --location global \
    --keyring "${KYMA_KEYRING}" \
    --key "${KYMA_ENCRYPTION_KEY}" \
    --ciphertext-file "${CIPHER_TEXT}" \
    --plaintext-file "${PLAIN_TEXT}" \
    --project "${CLOUDSDK_CORE_PROJECT}"