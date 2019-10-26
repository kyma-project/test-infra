#!/usr/bin/env bash

function kubernetes::ensure_kubectl {
    # TODO: (@michal-hudy) Install proper kubectl version for current system
    echo "Install kubectl in version ${1}"
    curl -LO "https://storage.googleapis.com/kubernetes-release/release/${1}/bin/linux/amd64/kubectl" --fail \
        && chmod +x kubectl
}