#!/usr/bin/env bash

#Description: decrypts cert files from gcloud
# The purpose of the script is to decrypt the private key and cert for HTTPS in nightly builds.
#
#Expected vars:
# - DOMAIN: Domain name
# - KYMA_KEYRING: kyma keyring name
# - KYMA_ENCRYPTION_KEY: encryption key name used to encrypt the files
# - TEST_INFRA_SOURCES_DIR: Path for shout library

set -o errexit

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

shout "Decrypting ${DOMAIN}.build.kyma-project.io.key.encrypted"
gcloud kms decrypt --location global \
	--keyring "${KYMA_KEYRING}" \
	--key "${KYMA_ENCRYPTION_KEY}" \
	--ciphertext-file "./letsencrypt/live/${DOMAIN}/${DOMAIN}.build.kyma-project.io.key.encrypted" \
  --plaintext-file "./letsencrypt/live/${DOMAIN}/privkey.pem"

shout "Decrypting ${DOMAIN}.build.kyma-project.io.key.encrypted"
gcloud kms decrypt --location global \
	--keyring "${KYMA_KEYRING}" \
	--key "${KYMA_ENCRYPTION_KEY}" \
	--ciphertext-file "./letsencrypt/live/${DOMAIN}/${DOMAIN}.build.kyma-project.io.cert.encrypted" \
  --plaintext-file "./letsencrypt/live/${DOMAIN}/fullchain.pem"