#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}"/log.sh

# utils::check_required_vars checks if all provided variables are initialized
#
# Arguments
# $1 - list of variables
function utils::check_required_vars() {
  log::info "Checks if all provided variables are initialized"
    local discoverUnsetVar=false
    for var in "$@"; do
      if [ -z "${!var}" ] ; then
        log::warn "ERROR: $var is not set"
        discoverUnsetVar=true
      fi
    done
    if [ "${discoverUnsetVar}" = true ] ; then
      exit 1
    fi
}

# utils::generate_self_signed_cert generates self-signed certificate for the given domain
#
# Arguments:
#
# required:
# d - domain name
# s - subdomain
#
# optional:
# v - number days certificate will be valid, default 5 days
#
# Return variables
# utils_generate_self_signed_cert_return_tls_cert - generated tls certificate
# utils_generate_self_signed_cert_return_tls_key - generated tls key
function utils::generate_self_signed_cert() {

    local OPTIND
    local dnsSubDomain
    local dnsDomain
    local certValidDays="5"

    while getopts ":s:d:v:" opt; do
        case $opt in
            s)
                dnsSubDomain="$OPTARG";;
            d)
                dnsDomain="${OPTARG%.}";;
            v)
                certValidDays="${OPTARG:-$certValidDays}";;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$dnsDomain" "Domain name not provided."
    utils::check_empty_arg "$dnsSubDomain" "Subdomain not provided."

    log::info "Generate self-signed certificate"
    local dnsFQDN="$dnsSubDomain.$dnsDomain"
    tmpDir=$(mktemp -d)
    local certPath="$tmpDir/cert.pem"
    local keyPath="$tmpDir/key.pem"

    openssl req -x509 -nodes -days "$certValidDays" -newkey rsa:4096 \
        -subj "/CN=$dnsFQDN" \
        -reqexts SAN -extensions SAN \
        -config <(cat /etc/ssl/openssl.cnf \
        <(printf "\\n[SAN]\\nsubjectAltName=DNS:*.%s" "$dnsFQDN")) \
        -keyout "$keyPath" \
        -out "$certPath"

    # return value
    # shellcheck disable=SC2034
    utils_generate_self_signed_cert_return_tls_cert=$(base64 "$certPath" | tr -d '\n')
    # return value
    # shellcheck disable=SC2034
    utils_generate_self_signed_cert_return_tls_key=$(base64 "$keyPath" | tr -d '\n')

    rm "$keyPath"
    rm "$certPath"
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
    return 1
  fi
  if [ -z "$2" ]; then
    echo "Remote name is empty. Exiting..."
    return 1
  fi
  if [ -z "$3" ]; then
    echo "Remote path is empty. Exiting..."
    return 1
  fi
  if [ -z "$4" ]; then
    echo "Local path is empty. Exiting..."
    return 1
  fi
  local ZONE=$1
  local REMOTE_NAME=$2
  local REMOTE_PATH=$3
  local LOCAL_PATH=$4

  for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && log::info 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute scp --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" --verbosity="${GCLOUD_SCP_LOG_LEVEL:-error}" --strict-host-key-checking=no --quiet --recurse --zone="${ZONE}" "${REMOTE_NAME}":"${REMOTE_PATH}" "${LOCAL_PATH}" && break;
    [[ ${i} -ge 5 ]] && log::error "Failed after $i attempts." && return 1
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
  log::info "Sends file(s) to Google Compute Platform over scp"
  log::info "Checking compute zone, remote name, local path and remote path arguments"
  if [ -z "$1" ]; then
    echo "Zone is empty. Exiting..."
    return 1
  fi
  if [ -z "$2" ]; then
    echo "Remote name is empty. Exiting..."
    return 1
  fi
  if [ -z "$3" ]; then
    echo "Local path is empty. Exiting..."
    return 1
  fi
  if [ -z "$4" ]; then
    echo "Remote path is empty. Exiting..."
    return 1
  fi
  local ZONE=$1
  local REMOTE_NAME=$2
  local LOCAL_PATH=$3
  local REMOTE_PATH=$4

  for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && log::info 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute scp --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" --verbosity="${GCLOUD_SCP_LOG_LEVEL:-error}" --strict-host-key-checking=no --quiet --recurse --zone="${ZONE}" "${LOCAL_PATH}" "${REMOTE_NAME}":"${REMOTE_PATH}" && break;
    [[ ${i} -ge 5 ]] && log::error "Failed after $i attempts." && return 1
  done;
}

# utils::ssh_to_vm_with_script communicate to Google Compute Platform over ssh
#
# Arguments:
#
# required:
# z - compute zone
# n - remote name
# c - ssh command
#
# optional:
# p - local script path
function utils::ssh_to_vm_with_script() {
  log::info "Communicate to Google Compute Platform over ssh"
  local OPTIND
  local ZONE
  local REMOTE_NAME
  local COMMAND
  local LOCAL_SCRIPT_PATH

  log::info "Checking Compute Zone, Remote name and ssh command required arguments"
  while getopts ":z:n:c:p:" opt; do
      case $opt in
          z)
              ZONE="$OPTARG";;
          n)
              REMOTE_NAME="$OPTARG";;
          c)
              COMMAND="$OPTARG";;
          p)
              LOCAL_SCRIPT_PATH="$OPTARG";;
          \?)
              echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
          :)
              echo "Option -$OPTARG argument not provided" >&2 ;;
      esac
  done

  utils::check_empty_arg "$ZONE" "compute zone not provided."
  utils::check_empty_arg "$REMOTE_NAME" "remote name not provided."
  utils::check_empty_arg "$COMMAND" "ssh command not provided."

  if [ -z "${LOCAL_SCRIPT_PATH}" ]; then
      gcloud compute ssh --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" --verbosity="${GCLOUD_SSH_LOG_LEVEL:-error}" --quiet --zone="${ZONE}" --command="${COMMAND}" --ssh-flag="-o ServerAliveInterval=10 -o TCPKeepAlive=no -o ServerAliveCountMax=60 -v" "${REMOTE_NAME}"
  else
      gcloud compute ssh --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" --verbosity="${GCLOUD_SSH_LOG_LEVEL:-error}" --quiet --zone="${ZONE}" --command="${COMMAND}" --ssh-flag="-o ServerAliveInterval=10 -o TCPKeepAlive=no -o ServerAliveCountMax=60 -v" "${REMOTE_NAME}" < "${LOCAL_SCRIPT_PATH}"
  fi
}

# utils::compress_send_to_vm compresses and sends file(s) to Google Compute Platform over scp
#
# Arguments
# $1 - compute zone
# $2 - remote name
# $3 - local path
# $4 - remote path
function utils::compress_send_to_vm() {
  log::info "Compresses and sends file(s) to Google Compute Platform over scp"
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
  gcloud compute ssh --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" --verbosity="${GCLOUD_SSH_LOG_LEVEL:-error}" --strict-host-key-checking=no --quiet --zone="${ZONE}" --command="mkdir -p ${REMOTE_PATH} && tar -xf ~/pack.tar.gz -C ${REMOTE_PATH}" --ssh-flag="-o ServerAliveInterval=10 -o TCPKeepAlive=no -o ServerAliveCountMax=60" "${REMOTE_NAME}"

  rm -rf "${TMP_DIRECTORY}"
}


# utils::save_psp_list generates pod-security-policy list and saves it to json file
#
# Arguments
# $1 - Name of the output json file
function utils::save_psp_list() {
  log::info "generates pod-security-policy list and saves it to json file"
  log::info "json file name: $1"
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

# utils::describe_nodes call k8s statistics commands and check if oom event was recorded.
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

# utils::oom_get_output download output from debug command pod if exist.
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
  # run oom debug pod
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

# utils::post_hook runs at the end of a script or on any error.
#
# Arguments:
# required:
# p - GCP project name
# E - exit status to report at the end of function execution
# j - job name
#
# optional:
# c - if set to true cleanup cluster, default false
# g - if set to true cleanup gateway DNS, default false
# G - gateway hostname to clean, default *
# a - if set to true cleanup apiserver DNS, default false
# A - apiserver hostname to clean, default apiserver
# I - if set to true cleanup gateway IP, default false
# l - if set to true enable error logging guard
# z - GCP compute zone, default europe-west4-b
# R - GCP compute region, default europe-west4
# r - if true clean regional cluster, default false
# d - if true deprovision cluster in async mode, default true
# n - cluster name to deprovision
# s - dns subdomain
# e - gateway IP address
# f - apiserver IP address
# N - gateway IP address name
# Z - GCP dns zone name
# k - if set to true, this is a Kyma 2.0 cluster; default false
#
function utils::post_hook() {
    # enabling path globbing, disabled in a trap before utils::post_hook call
    set +f

    local OPTIND
    local projectName
    local exitStatus
    local cleanupCluster="false"
    local cleanupGatewayDns="false"
    local gatewayHostname='*'
    local cleanupApiserverDns="false"
    local apiserverHostname='apiserver'
    local cleanupGatewayIP="false"
    local errorLoggingGuard="false"
    local computeZone="europe-west4-b"
    local computeRegion="europe-west4"
    local cleanRegionalCluster="false"
    local asyncDeprovision="true"
    local jobname
    local kyma2="false"

    while getopts ":n:c:l:p:a:G:g:z:I:r:d:R:A:e:f:s:Z:N:E:j:k:" opt; do
        case $opt in
            p)
                projectName="$OPTARG" ;;
            E)
                exitStatus="$OPTARG" ;;
            j)
                jobname="$OPTARG" ;;
            c)
                cleanupCluster="${OPTARG:-$cleanupCluster}" ;;
            g)
                cleanupGatewayDns="${OPTARG:-$cleanupGatewayDns}" ;;
            G)
                gatewayHostname="${OPTARG:-$gatewayHostname}" ;;
            a)
                cleanupApiserverDns="${OPTARG:-$cleanupApiserverDns}" ;;
            A)
                apiserverHostname="${OPTARG:-$apiserverHostname}" ;;
            I)
                cleanupGatewayIP="${OPTARG:-$cleanupGatewayIP}" ;;
            l)
                errorLoggingGuard="${OPTARG:-$errorLoggingGuard}" ;;
            z)
                computeZone=${OPTARG:-$computeZone} ;;
            R)
                computeRegion=${OPTARG:-$computeRegion} ;;
            r)
                cleanRegionalCluster=${OPTARG:-$cleanRegionalCluster} ;;
            d)
                asyncDeprovision=${OPTARG:-$asyncDeprovision} ;;
            k)
                kyma2=${OPTARG:-$kyma2} ;;
            n)
                if [ -n "$OPTARG" ]; then
                    local clusterName="$OPTARG"
                fi ;;
            s)
                if [ -n "$OPTARG" ]; then
                    local dnsSubDomain="$OPTARG"
                fi ;;
            e)
                if [ -n "$OPTARG" ]; then
                    local gatewayIP="$OPTARG"
                fi ;;
            f)
                if [ -n "$OPTARG" ]; then
                    local apiserverIP="$OPTARG"
                fi ;;
            N)
                if [ -n "$OPTARG" ]; then
                    local gatewayIpAddressName="$OPTARG"
                fi ;;
            Z)
                if [ -n "$OPTARG" ]; then
                    local gcpDnsZoneName="$OPTARG"
                fi ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2; ;;
        esac
    done

    if [ "$kyma2" = "false" ]; then
        kubectl get installation kyma-installation -o go-template --template='{{- range .status.errorLog }}{{printf "%s:\n %s\n" .component .log}}{{- end}}'
        kubectl logs -n kyma-installer -l name=kyma-installer
    fi

    utils::check_empty_arg "$projectName" "Project name not provided." "graceful"
    utils::check_empty_arg "$exitStatus" "Exit status not provided." "graceful"
    utils::check_empty_arg "$jobname" "Job name not provided." "graceful"

    if [ "$errorLoggingGuard" = "true" ]; then
        log::info "AN ERROR OCCURED! Take a look at preceding log entries."
    fi

    log::info "Collect logs"

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    log::info "Cleanup"

    utils::oom_get_output
    if [ "$cleanupCluster" = "true" ]; then
        gcp::deprovision_k8s_cluster \
            -n "$clusterName" \
            -p "$projectName" \
            -z "$computeZone" \
            -R "$computeRegion" \
            -r "$cleanRegionalCluster" \
            -d "$asyncDeprovision"
    fi
    if [ "$cleanupGatewayDns" = "true" ]; then
        gcp::delete_dns_record \
            -a "$gatewayIP" \
            -p "$projectName" \
            -h "$gatewayHostname" \
            -s "$dnsSubDomain" \
            -z "$gcpDnsZoneName"
    fi
    if [ "$cleanupGatewayIP" = "true" ]; then
        gcp::delete_ip_address \
            -p "$projectName" \
            -n "$gatewayIpAddressName" \
            -R "$computeRegion"
    fi
    if [ "$cleanupApiserverDns" = "true" ]; then
        gcp::delete_dns_record \
            -a "$apiserverIP" \
            -p "$projectName" \
            -h "$apiserverHostname" \
            -s "$dnsSubDomain" \
            -z "$gcpDnsZoneName"
    fi

    local msg=""
    if [[ $exitStatus -ne 0 ]]; then msg="(exit status: $exitStatus)"; fi
    log::info "Job is finished $msg"
    set -e

    exit "$exitStatus"
}


# utils::run_jobguard will start jobguard if build type is set to pr
# Arguments
# b - Build type set for prowjob
# P - path to test-infra repository sources root directory
function utils::run_jobguard() {

    local OPTIND
    local testInfraSourcesDir="/home/prow/go/src/github.com/kyma-project"

    while getopts ":b:P:" opt; do
        case $opt in
            b)
                local buildType="$OPTARG" ;;
            P)
                testInfraSourcesDir=${OPTARG:-$testInfraSourcesDir} ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2; ;;
        esac
    done

    utils::check_empty_arg "$buildType"
    buildType=$( echo "$buildType" | tr "[:upper:]" "[:lower:]")
    if [[ "$buildType" == "pr" ]]; then
        log::info "Execute Job Guard"
        # shellcheck source=cmd/jobguard/run.sh
        "$testInfraSourcesDir"/cmd/jobguard/run.sh
    fi
}

# utils::generate_CommonName generate common name
# It generates random string and prefix it, with provided arguments
#
# Arguments:
#
# optional:
# n - string to use as a common name prefix
# p - pull request number or commit id to use as a common name prefix
#
# Return:
# utils_generate_commonName_return_commonName - generated common name string
utils::generate_commonName() {

    local OPTIND

    while getopts ":n:p:" opt; do
        case $opt in
            n)
                local namePrefix="$OPTARG" ;;
            p)
                local id="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2; ;;
        esac
    done

    namePrefix=$(echo "$namePrefix" | tr '_' '-')
    namePrefix=${namePrefix#-}

    local randomNameSuffix
    randomNameSuffix=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
    # return value
    # shellcheck disable=SC2034
    utils_generate_commonName_return_commonName=$(echo "$namePrefix$id$randomNameSuffix" | tr "[:upper:]" "[:lower:]" )
}

# check_empty_arg will check if first argument is empty.
# If it's empty it will log second argument as error message and exit with code 1.
# If second argument is empty, generic default log message will be used.
#
# Arguments:
# $1 - argument to check if it's empty
# $2 - log message to use if $1 is empty
function utils::check_empty_arg {
    if [ -z "$2" ]; then
        logMessage="Mandatory argument is empty."
    else
        logMessage="$2"
    fi
    if [ -z "$1" ]; then
        if [ -n "$3" ]; then
            log::error "$logMessage"
        else
            log::error "$logMessage Exiting"
            exit 1
        fi
    fi
}

# utils::generate_vars_for_build generate string values for specific build types
#
# Arguments:
#
# optional:
# b - build type
# p - pull request number, required for build type pr
# s - pull request base SHA, required for build type commit
# n - prowjob name required for other build types
#
# Return variables:
# utils_set_vars_for_build_return_commonName - generated common name
# utils_set_vars_for_build_return_kymaSource - generated kyma source
function utils::generate_vars_for_build {
  log::info "Generate string values for specific build types"

    local OPTIND

    while getopts ":b:p:s:n:" opt; do
        case $opt in
            b)
                local buildType="$OPTARG" ;;
            p)
                local prNumber="$OPTARG" ;;
            s)
                local prBaseSha="$OPTARG" ;;
            n)
                local prowjobName="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
   done

    if [ "$buildType" = "pr" ]; then
        utils::check_empty_arg "$prNumber" "Pull request number not provided."
    fi

    # In case of PR, operate on PR number
    if [[ "$buildType" == "pr" ]]; then
        utils::generate_commonName \
            -n "pr" \
            -p "$prNumber"
        utils_generate_vars_for_build_return_commonName=${utils_generate_commonName_return_commonName:?}
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_kymaSource="PR-$prNumber"
    elif [[ "$buildType" == "release" ]]; then
        log::info "Reading release version from VERSION file"
        readonly releaseVersion=$(cat "VERSION")
        log::info "Read release version: $releaseVersion"
        utils::generate_commonName \
            -n "rel"
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_commonName=${utils_generate_commonName_return_commonName:?}
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_kymaSource="$releaseVersion"
    # Otherwise (master), operate on triggering commit id
    elif [ -n "$prBaseSha" ]; then
        readonly commitID="${prBaseSha::8}"
        utils::generate_commonName \
            -n "commit" \
            -p "$commitID"
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_commonName=${utils_generate_commonName_return_commonName:?}
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_kymaSource="$commitID"
    elif [ -n "$prowjobName" ]; then
        prowjobName=${prowjobName: -20:20}
        utils::generate_commonName \
            -n "$prowjobName"
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_commonName=${utils_generate_commonName_return_commonName:?}
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_kymaSource="main"
    else
        log::error "Build type not known. Set -b parameter to value 'pr' or 'release', or set -s or -n parameter."
    fi
}

# utils::mask_debug_output disables bash option x. This way it prevents from printing variables and command argument values.
# Function should be called just before statement which need masking in case set -x is used to troubleshoot bash script.
# A block with masked debug output must be closed by calling function utils::unmask_debug_output
function utils::mask_debug_output {
    if ( echo $- | grep x ); then
        log::info "Disabling bash option x. Enter secret masking block."
        utils_mask_debug_output_return_masked="true"
        set +x
    fi
}

# utils::unmask_debug_output enables bash option x if it was disabled by utils::mask_debug_output.
# Function should be called just after statement which need masking in case set -x is used to troubleshoot bash script.
function utils::unmask_debug_output {
    if [[ ${utils_mask_debug_output_return_masked:-"false"} == "true" ]]; then
        log::info "Enabling bash option x. Exit secret masking block"
        unset utils_mask_debug_output_return_masked
        set -x
    fi
}

# utils::install_yq installs yq CLI
function utils::install_yq {
    local settings
    local yq_version
    settings="$(set +o); set -$-"
    mkdir -p "/tmp/bin"
    export PATH="/tmp/bin:${PATH}"

    pushd "/tmp/bin" || exit

    log::info "--> Install yq CLI locally to /tmp/bin"

    curl -sSLo yq.tar.gz "https://github.com/mikefarah/yq/releases/download/v4.25.1/yq_linux_amd64.tar.gz"

    tar xvzf yq.tar.gz
    mv ./yq_linux_amd64 yq

    chmod +x yq
    yq_version=$(yq --version)
    log::info "--> yq CLI version: ${yq_version}"
    log::info "OK"
    popd || exit
    eval "${settings}"
}

# utils::install_helm installs helm CLI
function utils::install_helm {
    local settings
    local helm_version
    settings="$(set +o); set -$-"
    mkdir -p "/tmp/bin"
    export PATH="/tmp/bin:${PATH}"

    pushd "/tmp/bin" || exit

    log::info "--> Install helm CLI locally to /tmp/bin"

    curl -sSLo helm.tar.gz "https://get.helm.sh/helm-v3.8.0-linux-amd64.tar.gz"

    tar xvzf helm.tar.gz
    mv ./linux-amd64/helm .

    chmod +x helm
    helm_version=$(helm version)
    log::info "--> helm CLI version: ${helm_version}"
    log::info "OK"
    popd || exit
    eval "${settings}"
}
