#!/usr/bin/env bash

#Description: encrypts cert files from gcloud
# The purpose of the script is to encrypt the private key and cert for HTTPS in nightly builds.
#
#Expected vars:
# - KYMA_KEYRING: kyma keyring name
# - KYMA_ENCRYPTION_KEY: encryption key name used to encrypt the files
#

set -o errexit
printf "encrypting nightly-gke-tls-integration-app-client-key.encrypted"
  gcloud kms encrypt --location global \
	--keyring "${KYMA_KEYRING}" \
	--key "${KYMA_ENCRYPTION_KEY}" \
	--plaintext-file "./letsencrypt/live/${DOMAIN}/privkey.pem" \
 	--ciphertext-file "./letsencrypt/live/${DOMAIN}/nightly-gke-tls-integration-app-client-key.encrypted"

printf "encrypting nightly-gke-tls-integration-app-client-cert.encrypted"
  	gcloud kms encrypt --location global \
	--keyring "${KYMA_KEYRING}" \
	--key "${KYMA_ENCRYPTION_KEY}" \
	--plaintext-file "./letsencrypt/live/${DOMAIN}/fullchain.pem" \
	--ciphertext-file "./letsencrypt/live/${DOMAIN}/nightly-gke-tls-integration-app-client-cert.encrypted"