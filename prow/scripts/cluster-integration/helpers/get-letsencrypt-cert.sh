#!/usr/bin/env bash
export KYMA_NIGHTLY_KEY=nightly-gke-tls-integration-app-client-key.encrypted
export KYMA_NIGHTLY_CERT=nightly-gke-tls-integration-app-client-cert.encrypted
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="/Users/i862106/go/src/github.com/test-infra/prow/scripts/cluster-integration/helpers"
printf "Copying certificate if it is already in GCP Bucket.\n"
mkdir -p ./letsencrypt/live/"${DOMAIN}"
if [[ $(gsutil ls "gs://kyma-prow-secrets/${KYMA_NIGHTLY_CERT}") ]]; then
    printf "\nCertificate exists in vault. Downloading Key"
    #copy the files
    
    gsutil cp "gs://kyma-prow-secrets/${KYMA_NIGHTLY_CERT}" "./letsencrypt/live/${DOMAIN}" 
    gsutil cp "gs://kyma-prow-secrets/${KYMA_NIGHTLY_KEY}" "./letsencrypt/live/${DOMAIN}"
    
    printf "decrypting certs\n"
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/decrypt-certs.sh"
 if [[ !$(openssl x509 -checkend 86400 -noout -in "./letsencrypt/live/${DOMAIN}/fullchain.pem") ]]; then
  printf "Cert is Valid"
  fi
else
    printf "Generating Certificates"
    #Generate the certs
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-and-export-letsencrypt-TLS-cert.sh"

fi
   TLS_CERT=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/fullchain.pem | tr -d '\n')
    export TLS_CERT
    TLS_KEY=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/privkey.pem   | tr -d '\n')
    export TLS_KEY