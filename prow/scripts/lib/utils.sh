#!/usr/bin/env bash

utils::check_required_vars() {
    local discoverUnsetVar=false
    for var in "$@"; do
      if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
      fi
    done
    if [ "${discoverUnsetVar}" = true ] ; then
      exit 1
    fi
}
# utils::generate_self_signed_cert generates self-signed certificate for the given domain
#
# Optional exported variables
# CERT_VALID_DAYS - days when the certificate is valid
# Arguments
# $1 - domain name
function utils::generate_self_signed_cert() {
  if [ -z "$1" ]; then
    echo "Domain name is empty. Exiting..."
    exit 1
  fi
  local DOMAIN
  DOMAIN=$1
  local CERT_PATH
  local KEY_PATH
  tmpDir=$(mktemp -d)
  CERT_PATH="${tmpDir}/cert.pem"
  KEY_PATH="${tmpDir}/key.pem"
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
}
