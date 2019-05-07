#!/usr/bin/env bash

#Description: Generates and exports LetsEncrypt TLS certificates
#
#Expected vars:
# - DOMAIN: Combination of gcloud managed-zones and cluster name "${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
# - GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

discoverUnsetVar=false

for var in DOMAIN GOOGLE_APPLICATION_CREDENTIALS; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

shout "Generate lets encrypt certificate"
date

mkdir letsencrypt
cp ${GOOGLE_APPLICATION_CREDENTIALS} letsencrypt
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