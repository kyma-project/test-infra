#!/usr/bin/env bash

# helm::ensure_client downloads proper Helm client version
#
# Arguments:
#   $1 - Helm client version
#   $2 - Host OS
#   $3 - Destination directory
function helm::ensure_client {
    echo "Install Helm client in version ${1}"

    wget "https://get.helm.sh/helm-${1}-${2}-amd64.tar.gz" -O - | tar -xzO "${2}-amd64/helm" > "${3}/helm" \
        && chmod +x "${3}/helm"
        rm -rf "helm-${1}-${2}-amd64.tar.gz" linux-amd64

}
