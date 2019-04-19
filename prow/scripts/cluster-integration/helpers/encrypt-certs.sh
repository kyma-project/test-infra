#!/usr/bin/env bash

#Description: encrypts cert files from gcloud
# The purpose of the script is to encrypt the private key and cert for HTTPS in nightly builds.
#
#Expected vars:
# - KYMA_KEYRING: kyma keyring name
# - KYMA_ENCRYPTION_KEY: encryption key name used to encrypt the files
# - TEST_INFRA_SOURCES_DIR: Path for shout library

set -o errexit

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

shout "Encrypting ${KYMA_NIGHTLY_KEY}"
gcloud kms encrypt --location global \
	--keyring "${KYMA_KEYRING}" \
	--key "${KYMA_ENCRYPTION_KEY}" \
	--plaintext-file "./letsencrypt/live/${DOMAIN}/privkey.pem" \
 	--ciphertext-file "./letsencrypt/live/${DOMAIN}/${KYMA_NIGHTLY_KEY}"

shout "Encrypting ${KYMA_NIGHTLY_CERT}"
gcloud kms encrypt --location global \
	--keyring "${KYMA_KEYRING}" \
	--key "${KYMA_ENCRYPTION_KEY}" \
	--plaintext-file "./letsencrypt/live/${DOMAIN}/fullchain.pem" \
	--ciphertext-file "./letsencrypt/live/${DOMAIN}/${KYMA_NIGHTLY_CERT}"