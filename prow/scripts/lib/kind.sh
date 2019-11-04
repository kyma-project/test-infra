#!/usr/bin/env bash

function kind::create_cluster {
    local -r image="kindest/node:${2}"
    kind create cluster --name "${1}" --image "${image}" --config "${3}" --wait 3m
    
    local -r kubeconfig="$(kind get kubeconfig-path --name="${1}")"
    cp "${kubeconfig}" "${HOME}/.kube/config"
    kubectl cluster-info
}

function kind::delete_cluster {
    kind delete cluster --name "${1}"
}

function kind::load_image {
    kind load docker-image "${2}" --name "${1}"
}

function kind::install_default {
    echo "Make kubernetes.io/host-path Storage Class as non default"
    kubectl annotate storageclass standard storageclass.kubernetes.io/is-default-class="false" storageclass.beta.kubernetes.io/is-default-class="false" --overwrite

    echo "Install default resources from ${1}"
    kubectl apply -f "${1}"
}

function kind::worker_ip {
    docker inspect -f "{{ .NetworkSettings.IPAddress }}" "${1}-worker"
}

function kind::export_logs {
    echo "Exporting cluster logs to ${ARTIFACTS_DIR}/cluster-logs"
    mkdir -p "${ARTIFACTS_DIR}/cluster-logs"
    kind export logs "${ARTIFACTS_DIR}/cluster-logs" --name "${1}"

    echo "Creating archive ${ARTIFACTS_DIR}/cluster-logs.tar.gz with cluster logs"
    cd "${ARTIFACTS_DIR}" && tar -zcf "${ARTIFACTS_DIR}/cluster-logs.tar.gz" "cluster-logs/"
}