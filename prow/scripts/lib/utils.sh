#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}"/log.sh

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

# utils::send_to_vm sends file(s) to Google Compute Platform over scp
#
# Arguments
# $1 - compute zone
# $1 - local path
# $2 - remote name
# $3 - remote path
function utils::send_to_vm() {
  if [ -z "$1" ]; then
    echo "Zone is empty. Exiting..."
    exit 1
  fi
  if [ -z "$2" ]; then
    echo "Remote name is empty. Exiting..."
    exit 1
  fi
  if [ -z "$3" ]; then
    echo "Local path is empty. Exiting..."
    exit 1
  fi
  if [ -z "$4" ]; then
    echo "Remote path is empty. Exiting..."
    exit 1
  fi
  ZONE=$1
  REMOTE_NAME=$2
  LOCAL_PATH=$3
  REMOTE_PATH=$4

  for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && log::info 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute scp --quiet --recurse --zone="${ZONE}" "${LOCAL_PATH}" "${REMOTE_NAME}":"${REMOTE_PATH}" && break;
    [[ ${i} -ge 5 ]] && log::error "Failed after $i attempts." && exit 1
  done;
}

# utils::compress_send_to_vm compresses and sends file(s) to Google Compute Platform over scp
#
# Arguments
# $1 - compute zone
# $1 - local path
# $2 - remote name
# $3 - remote path
function utils::compress_send_to_vm() {
  if [ -z "$1" ]; then
    echo "Zone is empty. Exiting..."
    exit 1
  fi
  if [ -z "$2" ]; then
    echo "Remote name is empty. Exiting..."
    exit 1
  fi
  if [ -z "$3" ]; then
    echo "Local path is empty. Exiting..."
    exit 1
  fi
  if [ -z "$4" ]; then
    echo "Remote path is empty. Exiting..."
    exit 1
  fi
  ZONE=$1
  REMOTE_NAME=$2
  LOCAL_PATH=$3
  REMOTE_PATH=$4

  local TMP_DIR
  TMP_DIR=$(mktemp -d)

  tar -czf "${TMP_DIR}/pack.tar.gz" -C "${LOCAL_PATH}" "."
  utils::send_to_vm "${ZONE}" "${REMOTE_NAME}" "${TMP_DIR}/pack.tar.gz" "${HOME}/"
  gcloud compute ssh --quiet --zone="${ZONE}" --command="mkdir ${REMOTE_PATH} && tar -xf ~/pack.tar.gz -C ${REMOTE_PATH}" --ssh-flag="-o ServerAliveInterval=30" "${REMOTE_NAME}" 
  
  rm -rf "${TMP_DIR}"
}