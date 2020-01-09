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