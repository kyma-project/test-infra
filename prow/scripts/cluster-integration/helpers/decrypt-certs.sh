#!/usr/bin/env bash
printf "decrypting nightly-gke-tls-integration-app-client-key.encrypted"
  local KYMA_KEYRING="kyma-prow"
  local KYMA_ENCRYPTION_KEY="projects/kyma-project/locations/global/keyRings/kyma-prow/cryptoKeys/kyma-prow-encryption"
    gcloud kms decrypt --location global \
	--keyring "${KYMA_KEYRING}" \
	--key "${KYMA_ENCRYPTION_KEY}" \
	--ciphertext-file letsencrypt/live/"${DOMAIN}"/nightly-gke-tls-integration-app-client-key.encrypted \
  --plaintext-file letsencrypt/live/"${DOMAIN}"/privkey.pem 