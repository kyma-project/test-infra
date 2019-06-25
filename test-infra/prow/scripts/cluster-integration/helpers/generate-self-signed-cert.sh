#!/usr/bin/env bash

# This script is meant to output key and cert to stdout.
# Key and certificate are being stored in files and later on removed because openssl doesn't have option to put them directly into variables

set -o errexit

if [ -z "$DOMAIN" ]; then
      echo "\$DOMAIN is empty"
      exit 1
fi

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CERT_PATH="${SCRIPTS_DIR}/cert.pem"
KEY_PATH="${SCRIPTS_DIR}/key.pem"
CERT_VALID_DAYS=${CERT_VALID_DAYS:-5}

openssl req -x509 -nodes -days "${CERT_VALID_DAYS}" -newkey rsa:4069 \
                 -subj "/CN=${DOMAIN}" \
                 -reqexts SAN -extensions SAN \
                 -config <(cat /etc/ssl/openssl.cnf \
        <(printf "\\n[SAN]\\nsubjectAltName=DNS:*.%s" "${DOMAIN}")) \
                 -keyout "${KEY_PATH}" \
                 -out "${CERT_PATH}"

TLS_CERT=$(base64 "${CERT_PATH}" | tr -d '\n')
TLS_KEY=$(base64 "${KEY_PATH}" | tr -d '\n')

echo "${TLS_CERT}"
echo "${TLS_KEY}"

rm "${KEY_PATH}"
rm "${CERT_PATH}"
