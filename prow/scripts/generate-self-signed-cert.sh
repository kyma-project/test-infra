#!/usr/bin/env bash

# This script is meant to output key and cert to stdout.
# Key and certificate are being stored in files and later on removed because openssl doesn't have option to put them directly into variables

set -o errexit

if [ -z "$DOMAIN" ]; then
      echo "\$DOMAIN is empty"
      exit 1
fi

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KEY_PATH="${SCRIPTS_DIR}/key.pem"
CERT_PATH="${SCRIPTS_DIR}/cert.pem"

openssl req -x509 -nodes -days 5 -newkey rsa:4069 \
                 -subj "/CN=${DOMAIN}" \
                 -reqexts SAN -extensions SAN \
                 -config <(cat /etc/ssl/openssl.cnf \
        <(printf "\\n[SAN]\\nsubjectAltName=DNS:*.%s" "${DOMAIN}")) \
                 -keyout "${KEY_PATH}" \
                 -out "${CERT_PATH}"

TLS_CERT=$(< /cert.pem base64 | tr -d '\n')
TLS_KEY=$(< /key.pem base64 | tr -d '\n')

echo "TLS_CERT=${TLS_CERT}"
echo "TLS_KEY=${TLS_KEY}"

rm "${KEY_PATH}"
rm "${CERT_PATH}"
