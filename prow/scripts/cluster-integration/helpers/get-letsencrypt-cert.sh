 printf "\nChecking if certificate is already in GCP Bucket."
      mkdir -p ./letsencrypt/live/"${DOMAIN}"
 if [[ $(gsutil ls gs://kyma-prow-secrets/nightly-gke-tls-integration-app-client-cert.encrypted) ]];
 then
    printf "\nCertificate/privatekey exists in vault. Downloading..."

    cp /etc/credentials/sa-gke-kyma-integration/service-account.json letsencrypt
  #copy the files
    gsutil cp gs://kyma-prow-secrets/nightly-gke-tls-integration-app-client-cert.encrypted "./letsencrypt/live/${DOMAIN}" 
    gsutil cp gs://kyma-prow-secrets/nightly-gke-tls-integration-app-client-key.encrypted "./letsencrypt/live/${DOMAIN}" 
#decrypt key and cert
printf "decrypting certs"
  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/decrypt-certs.sh"
    else
    printf "Generating Certificates"
    #Generate the certs
      "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-and-export-letsencrypt-TLS-cert.sh"
  fi