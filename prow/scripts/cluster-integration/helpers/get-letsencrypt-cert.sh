#!/usr/bin/env bash

#Description: gets cert files from gcloud
# The purpose of the script is to get the private key and cert for HTTPS in nightly builds, if they are valid and availble
#
#Expected vars:
# - DOMAIN: encryption key name used to encrypt the files
# - KYMA_NIGHTLY_KEY: kyma nightly key
# - TEST_INFRA_SOURCES_DIR: directory of scripts
# - GOOGLE_APPLICATION_CREDENTIALS: credentials to read/write to gcloud storage
set -o errexit

#shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

function generateLetsEncryptCert() {
    DOMAIN="${DOMAIN}" GOOGLE_APPLICATION_CREDENTIALS="${GOOGLE_APPLICATION_CREDENTIALS}" "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/helpers/generate-and-export-letsencrypt-TLS-cert.sh

    shout "Encrypting certs"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/encrypt.sh" \
        "./letsencrypt/live/${DOMAIN}/privkey.pem"  \
        "./letsencrypt/live/${DOMAIN}/${DOMAIN}.key.encrypted"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/encrypt.sh"  \
        "./letsencrypt/live/${DOMAIN}/fullchain.pem"  \
        "./letsencrypt/live/${DOMAIN}/${DOMAIN}.cert.encrypted"

    gsutil cp "./letsencrypt/live/${DOMAIN}/${DOMAIN}.cert.encrypted" "gs://${CERTIFICATES_BUCKET}/certificates/"
    gsutil cp "./letsencrypt/live/${DOMAIN}/${DOMAIN}.key.encrypted" "gs://${CERTIFICATES_BUCKET}/certificates/"

}

shout "Copying certificate if it is already in GCP Bucket."

set +e # temp disable fail on exit to retrieve error codes of stat
gsutil -q stat "gs://${CERTIFICATES_BUCKET}/certificates/${DOMAIN}.cert.encrypted"
VALID_CERT_FILE=$?
gsutil -q stat "gs://${CERTIFICATES_BUCKET}/certificates/${DOMAIN}.key.encrypted"
VALID_KEY_FILE=$?
set -o errexit # reset to errexit

if [[ $VALID_CERT_FILE -eq 0 && $VALID_KEY_FILE -eq 0 ]]; then
    shout "Certificate exists in vault. Downloading Key"

    #copy the files
    mkdir -p "./letsencrypt/live/${DOMAIN}"
    gsutil cp "gs://${CERTIFICATES_BUCKET}/certificates/${DOMAIN}.cert.encrypted" "./letsencrypt/live/${DOMAIN}"
    gsutil cp "gs://${CERTIFICATES_BUCKET}/certificates/${DOMAIN}.key.encrypted" "./letsencrypt/live/${DOMAIN}"


    shout "Decrypting certs"
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/decrypt.sh" \
    "./letsencrypt/live/${DOMAIN}/privkey.pem"  \
    "./letsencrypt/live/${DOMAIN}/${DOMAIN}.key.encrypted"
    
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/decrypt.sh"  \
    "./letsencrypt/live/${DOMAIN}/fullchain.pem"  \
    "./letsencrypt/live/${DOMAIN}/${DOMAIN}.cert.encrypted"
    set +e
    openssl x509 -checkend 1209600 -noout -in "$(pwd)/letsencrypt/live/${DOMAIN}/fullchain.pem"
    VALID_CERT=$?
    set -o errexit
    if [[ $VALID_CERT -eq 0 ]]; then
        shout "Cert is Valid"

    else
        shout "Generating Certificates because it's invalid"
        #Generate the certs
        rm -rf "./letsencrypt/live"
        generateLetsEncryptCert

    fi
else
    shout "Generating Certificates because none exist"
    #Generate the certs
    generateLetsEncryptCert
fi
