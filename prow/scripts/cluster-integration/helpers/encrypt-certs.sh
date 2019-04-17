#!/usr/bin/env bash
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