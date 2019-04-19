#!/usr/bin/env bash

#Description: decrypts cert files from gcloud
# The purpose of the script is to decrypt the private key and cert for HTTPS in nightly builds.
#
#Expected vars:
# - KYMA_KEYRING: kyma keyring name
# - KYMA_ENCRYPTION_KEY: encryption key name used to encrypt the files
#
printf "decrypting nightly-gke-tls-integration-app-client-key.encrypted"
    gcloud kms decrypt --location global \
	--keyring "${KYMA_KEYRING}" \
	--key "${KYMA_ENCRYPTION_KEY}" \
	--ciphertext-file "letsencrypt/live/${DOMAIN}/nightly-gke-tls-integration-app-client-key.encrypted" \
  --plaintext-file "./letsencrypt/live/${DOMAIN}/privkey.pem "

	printf "decrypting nightly-gke-tls-integration-app-client-cert.encrypted"
   gcloud kms decrypt --location global \
	--keyring "${KYMA_KEYRING}" \
	--key "${KYMA_ENCRYPTION_KEY}" \
	--ciphertext-file "./letsencrypt/live/"${DOMAIN}"/nightly-gke-tls-integration-app-client-cert.encrypted" \
  --plaintext-file "./letsencrypt/live/${DOMAIN}/fullchain.pem"