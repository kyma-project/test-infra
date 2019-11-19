#!/usr/bin/env bash

# kind::create_cluster creates a kind cluster and points to it in the default kubeconfig
#
# Globals:
#   HOME - Path to current user home directory
# Arguments:
#   $1 - Cluster name
#   $2 - Kubernetes version
#   $3 - Path to Kind config file
function kind::create_cluster {
    local -r image="kindest/node:${2}"
    kind create cluster --name "${1}" --image "${image}" --config "${3}" --wait 3m
    
    local -r kubeconfig="$(kind get kubeconfig-path --name="${1}")"
    cp "${kubeconfig}" "${HOME}/.kube/config"
    kubectl cluster-info
}

# kind::delete_cluster deletes a kind cluster with provided name
#
# Arguments:
#   $1 - Cluster name
function kind::delete_cluster {
    kind delete cluster --name "${1}"
}

# kind::load_image loads Docker Image to the cluster
#
# Arguments:
#   $1 - Cluster name
#   $2 - Docker Image name
function kind::load_image {
    kind load docker-image "${2}" --name "${1}"
}

# kind::install_default installs default resources in cluster - StorageClass and provided files
#
# Arguments:
#   $1 - Path to the resource yaml file or directory with resources yamls
function kind::install_default {
    echo "Make kubernetes.io/host-path Storage Class as non default"
    kubectl annotate storageclass standard storageclass.kubernetes.io/is-default-class="false" storageclass.beta.kubernetes.io/is-default-class="false" --overwrite

    echo "Install default resources from ${1}"
    kubectl apply -f "${1}"
}

# kind::worker_ip returns the IP address of worker node
#
# Arguments:
#   $1 - Cluster name
# Returns:
#   Worker IP address
function kind::worker_ip {
    docker inspect -f "{{ .NetworkSettings.IPAddress }}" "${1}-worker"
}

# kind::export_logs exports all logs from the cluster to the artifacts directory. Creates also an archive with logs.
#
# Globals:
#   ARTIFACTS_DIR - Path to the artifacts directory
# Arguments:
#   $1 - Cluster name
function kind::export_logs {
    echo "Exporting cluster logs to ${ARTIFACTS_DIR}/cluster-logs"
    mkdir -p "${ARTIFACTS_DIR}/cluster-logs"
    kind export logs "${ARTIFACTS_DIR}/cluster-logs" --name "${1}"

    echo "Creating archive ${ARTIFACTS_DIR}/cluster-logs.tar.gz with cluster logs"
    cd "${ARTIFACTS_DIR}" && tar -zcf "${ARTIFACTS_DIR}/cluster-logs.tar.gz" "cluster-logs/"
}
