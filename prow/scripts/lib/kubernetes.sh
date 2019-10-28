#!/usr/bin/env bash

function kubernetes::ensure_kubectl {
    # TODO: (@michal-hudy) Install proper kubectl version for current system
    if command -v kubectl >/dev/null 2>&1; then
        echo "Removing built-in kubectl version"
        rm -f "$(command -v kubectl)"
    fi

    echo "Install kubectl in version ${1}"
    curl -LO "https://storage.googleapis.com/kubernetes-release/release/${1}/bin/linux/amd64/kubectl" --fail \
        && chmod +x kubectl \
        && mv kubectl /usr/local/bin/kubectl

    kubectl version --client
}