#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}"/log.sh

# utils::check_required_vars checks if all provided variables are initialized
#
# Arguments
# $1 - list of variables
function utils::check_required_vars() {
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

# utils::generate_letsencrypt_cert generates let's encrypt certificate for the given domain
#
# Expected exported variables
# GOOGLE_APPLICATION_CREDENTIALS
#
# Arguments
# $1 - domain name
function utils::generate_letsencrypt_cert() {
  if [ -z "$1" ]; then
    echo "Domain name is empty. Exiting..."
    exit 1
  fi
  local DOMAIN
  DOMAIN=$1

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
      --manual-public-ip-logging-ok \
      --preferred-challenges dns \
      --manual-auth-hook /prow-tools/certbotauthenticator \
      --manual-cleanup-hook "/prow-tools/certbotauthenticator -D" \
      -d "*.${DOMAIN}"

  TLS_CERT=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/fullchain.pem | tr -d '\n')
  export TLS_CERT
  TLS_KEY=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/privkey.pem   | tr -d '\n')
  export TLS_KEY
}

# utils::receive_from_vm receives file(s) from Google Compute Platform over scp
#
# Arguments
# $1 - compute zone
# $2 - remote name
# $3 - remote path
# $4 - local path
function utils::receive_from_vm() {
  if [ -z "$1" ]; then
    echo "Zone is empty. Exiting..."
    exit 1
  fi
  if [ -z "$2" ]; then
    echo "Remote name is empty. Exiting..."
    exit 1
  fi
  if [ -z "$3" ]; then
    echo "Remote path is empty. Exiting..."
    exit 1
  fi
  if [ -z "$4" ]; then
    echo "Local path is empty. Exiting..."
    exit 1
  fi
  local ZONE=$1
  local REMOTE_NAME=$2
  local REMOTE_PATH=$3
  local LOCAL_PATH=$4

  for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && log::info 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute scp --quiet --recurse --zone="${ZONE}" "${REMOTE_NAME}":"${REMOTE_PATH}" "${LOCAL_PATH}" && break;
    [[ ${i} -ge 5 ]] && log::error "Failed after $i attempts." && exit 1
  done;
}

# utils::send_to_vm sends file(s) to Google Compute Platform over scp
#
# Arguments
# $1 - compute zone
# $2 - remote name
# $3 - local path
# $4 - remote path
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
  local ZONE=$1
  local REMOTE_NAME=$2
  local LOCAL_PATH=$3
  local REMOTE_PATH=$4

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
# $2 - remote name
# $3 - local path
# $4 - remote path
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
  local ZONE=$1
  local REMOTE_NAME=$2
  local LOCAL_PATH=$3
  local REMOTE_PATH=$4

  TMP_DIRECTORY=$(mktemp -d)

  tar -czf "${TMP_DIRECTORY}/pack.tar.gz" -C "${LOCAL_PATH}" "."
  #shellcheck disable=SC2088
  utils::send_to_vm "${ZONE}" "${REMOTE_NAME}" "${TMP_DIRECTORY}/pack.tar.gz" "~/"
  gcloud compute ssh --quiet --zone="${ZONE}" --command="mkdir ${REMOTE_PATH} && tar -xf ~/pack.tar.gz -C ${REMOTE_PATH}" --ssh-flag="-o ServerAliveInterval=30" "${REMOTE_NAME}"
  
  rm -rf "${TMP_DIRECTORY}"
}

# utils::deprovision_gardener_cluster deprovisions a Gardener cluster
#
# Arguments
# $1 - Gardener project name
# $2 - Gardener cluster name
# $3 - path to kubeconfig
function utils::deprovision_gardener_cluster() {
  if [ -z "$1" ]; then
    echo "Project name is empty. Exiting..."
    exit 1
  fi
  if [ -z "$2" ]; then
    echo "Cluster name is empty. Exiting..."
    exit 1
  fi
  if [ -z "$3" ]; then
    echo "Kubeconfig path is empty. Exiting..."
    exit 1
  fi
  GARDENER_PROJECT_NAME=$1
  GARDENER_CLUSTER_NAME=$2
  GARDENER_CREDENTIALS=$3

  local NAMESPACE="garden-${GARDENER_PROJECT_NAME}"

  kubectl --kubeconfig "${GARDENER_CREDENTIALS}" -n "${NAMESPACE}" annotate shoot "${GARDENER_CLUSTER_NAME}" confirmation.gardener.cloud/deletion=true --overwrite
  kubectl --kubeconfig "${GARDENER_CREDENTIALS}" -n "${NAMESPACE}" delete shoot "${GARDENER_CLUSTER_NAME}" --wait=false
}


# utils::save_psp_list generates pod-security-policy list and saves it to json file
#
# Arguments
# $1 - Name of the output json file
function utils::save_psp_list() {
  if [ -z "$1" ]; then
    echo "File name is empty. Exiting..."
    exit 1
  fi
  local output_path=$1

  # this is false-positive as we need to use single-quotes for jq
  # shellcheck disable=SC2016
  PSP_LIST=$(kubectl get pods --all-namespaces -o json | jq '{ pods: [ .items[] | .metadata.ownerReferences[0].name as $owner | .metadata.annotations."kubernetes.io\/psp" as $psp | { name: .metadata.name, namespace: .metadata.namespace, owner: $owner, psp: $psp} ] | group_by(.name) | map({ name: .[0].name, namespace: .[0].namespace, owner: .[0].owner, psp: .[0].psp }) | sort_by(.psp, .name)}' )
  echo "${PSP_LIST}" > "${output_path}"
}
