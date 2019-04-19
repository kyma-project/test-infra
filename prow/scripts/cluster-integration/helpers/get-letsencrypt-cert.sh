#!/usr/bin/env bash
 printf "Checking if certificate is already in GCP Bucket."
      mkdir -p ./letsencrypt/live/"${DOMAIN}"
 if [[ $(gsutil ls gs://kyma-prow-secrets/${KYMA_NIGHTLY_CERT}) ]];
 then
    printf "\nCertificate/privatekey exists in vault. Downloading..."
  #copy the files
    gsutil cp gs://kyma-prow-secrets/${KYMA_NIGHTLY_CERT} "./letsencrypt/live/${DOMAIN}" 
    gsutil cp gs://kyma-prow-secrets/${KYMA_NIGHTLY_KEY} "./letsencrypt/live/${DOMAIN}" 
printf "decrypting certs"
  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/decrypt-certs.sh"
      TLS_CERT=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/fullchain.pem | tr -d '\n')
    export TLS_CERT
    TLS_KEY=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/privkey.pem   | tr -d '\n')
    export TLS_KEY

    else
    printf "Generating Certificates"
    #Generate the certs
      "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-and-export-letsencrypt-TLS-cert.sh"
  fi