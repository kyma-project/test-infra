#!/usr/bin/env bash

#Description: Generates and exports LetsEncrypt TLS certificates
#
#Expected vars:
# - DOMAIN: Combination of gcloud managed-zones and cluster name "${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
# - GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

requiredVars=(
   DOMAIN
   GOOGLE_APPLICATION_CREDENTIALS
)

utils::check_required_vars "${requiredVars[@]}"

log::info "Generate lets encrypt certificate"

mkdir -p ./letsencrypt
cp "${GOOGLE_APPLICATION_CREDENTIALS}" letsencrypt
docker run  --name certbot \
    --rm  \
    -v "$(pwd)/letsencrypt:/etc/letsencrypt"    \
    -v "$(pwd)/certbot-log:/var/log/letsencrypt"    \
    -v "/prow-tools:/prow-tools" \
    -e "GOOGLE_APPLICATION_CREDENTIALS=/etc/letsencrypt/service-account.json" \
    certbot/certbot \
    certonly \
    -m "kyma.bot@sap.com" \
    --agree-tos \
    --no-eff-email \
    --server https://acme-v02.api.letsencrypt.org/directory \
    --manual \
    --preferred-challenges dns \
    --manual-auth-hook /prow-tools/certbotauthenticator \
    --manual-cleanup-hook "/prow-tools/certbotauthenticator -D" \
    -d "*.${DOMAIN}"

TLS_CERT=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/fullchain.pem | tr -d '\n')
export TLS_CERT
TLS_KEY=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/privkey.pem   | tr -d '\n')
export TLS_KEY
