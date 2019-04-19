#!/usr/bin/env bash

#Description: encrypts cert files from gcloud
# The purpose of the script is to encrypt the private key and cert for HTTPS in nightly builds.
#
#Expected vars:
# - KYMA_NIGHTLY_CERT: kyma nighly cert
# - DOMAIN: encryption key name used to encrypt the files
# - KYMA_NIGHTLY_KEY: kyma nightly key
# - TEST_INFRA_SOURCES_DIR: directory of scripts

set -o errexit

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

function generateLetsEncryptCert() {
	shout "Generate lets encrypt certificate"
	date

    mkdir letsencrypt
    cp /etc/credentials/sa-gke-kyma-integration/service-account.json letsencrypt
    docker run  --name certbot \
        --rm  \
        -v "$(pwd)/letsencrypt:/etc/letsencrypt"    \
        certbot/dns-google \
        certonly \
        -m "kyma.bot@sap.com" \
        --agree-tos \
        --no-eff-email \
        --dns-google \
        --dns-google-credentials /etc/letsencrypt/service-account.json \
        --server https://acme-v02.api.letsencrypt.org/directory \
        --dns-google-propagation-seconds=600 \
        -d "*.${DOMAIN}"

}
shout "Copying certificate if it is already in GCP Bucket."
mkdir -p ./letsencrypt/live/"${DOMAIN}"
if [[ $(gsutil ls "gs://kyma-prow-secrets/${KYMA_NIGHTLY_CERT}") ]]; then
    shout "\nCertificate exists in vault. Downloading Key"
    #copy the files
    
    gsutil cp "gs://kyma-prow-secrets/${KYMA_NIGHTLY_CERT}" "./letsencrypt/live/${DOMAIN}" 
    gsutil cp "gs://kyma-prow-secrets/${KYMA_NIGHTLY_KEY}" "./letsencrypt/live/${DOMAIN}"
    
    shout "Decrypting certs"
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/decrypt-certs.sh"
fi
if [[ $(openssl x509 -checkend 86400 -noout -in "./letsencrypt/live/${DOMAIN}/fullchain.pem") ]]; then
  shout "Cert is Valid"

else
    shout "Generating Certificates"
    #Generate the certs
    #generateLetsEncryptCert

fi
TLS_CERT=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/fullchain.pem | tr -d '\n')
export TLS_CERT
TLS_KEY=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/privkey.pem   | tr -d '\n')
export TLS_KEY

