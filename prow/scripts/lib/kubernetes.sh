#!/usr/bin/env bash

# kubernetes::ensure_kubectl downloads proper kubectl version
#
# Arguments:
#   $1 - Kubernetes version
#   $2 - Host OS
#   $3 - Destination directory
function kubernetes::ensure_kubectl {
    echo "Install kubectl in version ${1}"
    curl -LO "https://storage.googleapis.com/kubernetes-release/release/${1}/bin/${2}/amd64/kubectl" --fail \
        && chmod +x kubectl \
        && mv kubectl "${3}/kubectl"

    kubectl version --client
}

# kubernetes::is_pod_ready validate if specified POD is up and ready
#
# Arguments:
#   $1 - Path to the Kyma sources
#   $2 - Pod namespace
#   $3 - Pod's label name
#   $4 - Pod's label value
function kubernetes::is_pod_ready {
    "${1}/installation/scripts/is-ready.sh" "${2}" "${3}" "${4}"
}