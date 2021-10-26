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
    gcloud compute scp --strict-host-key-checking=no --quiet --recurse --zone="${ZONE}" "${REMOTE_NAME}":"${REMOTE_PATH}" "${LOCAL_PATH}" && break;
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
    gcloud compute scp --strict-host-key-checking=no --quiet --recurse --zone="${ZONE}" "${LOCAL_PATH}" "${REMOTE_NAME}":"${REMOTE_PATH}" && break;
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
  gcloud compute ssh --strict-host-key-checking=no --quiet --zone="${ZONE}" --command="mkdir ${REMOTE_PATH} && tar -xf ~/pack.tar.gz -C ${REMOTE_PATH}" --ssh-flag="-o ServerAliveInterval=30" "${REMOTE_NAME}"

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

# utils::save_env_file creates a .env file with all provided variables
#
# Arguments
# $1 - list of variables
function utils::save_env_file() {
  touch .env
  for var in "$@"; do
    if [ -z "${!var}" ] ; then
      echo "INFO: $var is not set"
      continue
    fi

    echo "${var}"=\""$(printenv "${var}")"\" >> .env
  done
}

function utils::describe_nodes() {
    {
      log::info "calling describe nodes"
      kubectl describe nodes
      log::info "calling top nodes"
      kubectl top nodes
      log::info "calling top pods"
      kubectl top pods --all-namespaces
    } > "${ARTIFACTS}/describe_nodes.txt"
    grep "System OOM encountered" "${ARTIFACTS}/describe_nodes.txt"
    last=$?
    if [[ $last -eq 0 ]]; then
      log::banner "OOM event found"
    fi
}


function utils::oom_get_output() {
  if [ ! -e "${ARTIFACTS}/describe_nodes.txt" ]; then
    utils::describe_nodes
  fi
  if [ "${DEBUG_COMMANDO_OOM}" = "true" ]; then
  log::info "Download OOM events details"
  pods=$(kubectl get pod -l "name=oom-debug" -o=jsonpath='{.items[*].metadata.name}')
  for pod in $pods; do
    kubectl logs "$pod" -c oom-debug > "${ARTIFACTS}/$pod.txt"
  done
  debugFiles=$(ls -1 "${ARTIFACTS}"/oom-debug-*.txt)
  for debugFile in $debugFiles; do
    grep "OOM event received" "$debugFile" > /dev/null
    last=$?
    if [[ $last -eq 0 ]]; then
      log::info "Print OOM events details"
      cat "$debugFile"
    else
      rm "$debugFile"
    fi
  done
  fi
}

# utils::debug_oom will create oom-debug daemonset
# it will create necessary clusterrolebindings to allow oom-debug pods run as privileged
function utils::debug_oom() {
  kubectl apply -f "${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/debug-container.yaml"
}

# utils::kubeaudit_create_report downlaods kubeaudit if necessary and checks for privileged containers
# Arguments
# $1 - Name of the output log file
function utils::kubeaudit_create_report() {
   if [ -z "$1" ]; then
    echo "Kubeuadit file path is empty. Exiting..."
    exit 1
  fi
  local kubeaudit_file=$1

  log::info "Gather Kubeaudit logs"
  if ! [[ -x "$(command -v ./kubeaudit)" ]]; then
    curl -sL https://github.com/Shopify/kubeaudit/releases/download/v0.11.8/kubeaudit_0.11.8_linux_amd64.tar.gz | tar -xzO kubeaudit > ./kubeaudit
    chmod +x ./kubeaudit
  fi
  # kubeaudit returns non-zero exit code when it finds issues
  # In the context of this job we just want to grab the logs
  # It should not break the execution of this script
  ./kubeaudit privileged privesc -p json  > "$kubeaudit_file" || true
}

# utils::kubeaudit_check_report analyzes kubeaudit.log file and returns list of non-compliant resources in kyma-system namespace
# Arguments
# $1 - Name of the input log file
# S2 - optional, name of the resource namespace. Defaults to "kyma-system"
function utils::kubeaudit_check_report() {
  if [ -z "$1" ]; then
    echo "Kubeuadit file path is empty. Exiting..."
    exit 1
  fi
  local kubeaudit_file=$1

  incompliant_resources=$(jq -c 'select( .ResourceNamespace == "kyma-system" )' < "$kubeaudit_file")
  compliant=$(echo "$incompliant_resources" | jq -r -s 'if length == 0 then "true" else "false" end')

  if [[ "$compliant" == "true" ]]; then
    log::info "All resources are compliant"
  else
    log::error "Not all resources are compliant:"
    echo "$incompliant_resources"
    exit 1
  fi
}

# post_hook runs at the end of a script or on any error
function utils::post_hook() {
  #!!! Must be at the beginning of this function !!!
  EXIT_STATUS=$?

  log::info "Cleanup"

  if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
    log::info "AN ERROR OCCURED! Take a look at preceding log entries."
  fi

  #Turn off exit-on-error so that next step is executed even if previous one fails.
  set +e

  # collect logs from failed tests before deprovisioning
  kyma::run_test_log_collector "post-main-kyma-gke-integration"

  gcloud::cleanup

  MSG=""
  if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
  log::info "Job is finished ${MSG}"
  set -e

  exit "${EXIT_STATUS}"
}


# run_jobguard will start jobguard if build type is set to pr
# Arguments
# $1 - Build type set for prowjob
function utils::run_jobguard() {
  buildType=$( echo "${1}" | tr "[:upper:]" "[:lower:]")
  if [[ "${buildType}" == "pr" ]]; then
    log::info "Execute Job Guard"
    # shellcheck source=development/jobguard/scripts/run.sh
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
  fi
}

# utils::generate_CommonName create and export COMMON_NAME variable
# it generates random part of COMMON_NAME and prefix it with provided arguments
#
# Arguments:
# $1 - string to use as a common name prefix /optional
# $2 - pull request number or commit id to use as a common name prefix /optional
# Exports
# COMMON_NAME
utils::generate_commonName() {
  NAME_PREFIX=$1
  local commitID=$2
  if [ ${#commitID} -gt 0 ]; then
    commitID="-${commitID}-"
  fi
  RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
  COMMON_NAME=$(echo "${NAME_PREFIX}${commitID}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
  export COMMON_NAME
}
