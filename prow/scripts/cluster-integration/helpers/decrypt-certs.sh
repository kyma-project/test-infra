#!/usr/bin/env bash
printf "decrypting nightly-gke-tls-integration-app-client-key.encrypted"
    gcloud kms decrypt --location global \
	--keyring "${KYMA_KEYRING}" \
	--key "${KYMA_ENCRYPTION_KEY}" \
	--ciphertext-file letsencrypt/live/"${DOMAIN}"/nightly-gke-tls-integration-app-client-key.encrypted \
  --plaintext-file letsencrypt/live/"${DOMAIN}"/privkey.pem 

	printf "decrypting nightly-gke-tls-integration-app-client-cert.encrypted"
   gcloud kms decrypt --location global \
	--keyring "${KYMA_KEYRING}" \
	--key "${KYMA_ENCRYPTION_KEY}" \
	--ciphertext-file letsencrypt/live/"${DOMAIN}"/nightly-gke-tls-integration-app-client-cert.encrypted \
  --plaintext-file "./letsencrypt/live/${DOMAIN}/fullchain.pem"