#!/usr/bin/env bash

source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
function generateAndExportLetsEncryptCert() {
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

    TLS_CERT=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/fullchain.pem | tr -d '\n')
    export TLS_CERT
    TLS_KEY=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/privkey.pem   | tr -d '\n')
    export TLS_KEY

}

generateAndExportLetsEncryptCert