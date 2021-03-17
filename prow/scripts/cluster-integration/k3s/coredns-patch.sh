#!/usr/bin/env bash

set -o errexit
set -o pipefail

# Description: Applies the coreDNS patch to a k3s cluster on gcloud
#

BASE_DIR="${PWD}"
export SCRIPTS_DIR="$BASE_DIR/scripts"
echo "SCRIPTS_DIR=${SCRIPTS_DIR}"

#while [[ $(kubectl get nodes -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "Waiting for cluster nodes to be ready, elapsed time: $(( $SECONDS/60 )) min $(( $SECONDS % 60 )) sec"; sleep 2; done
sleep 30

echo "Applying coreDSN patch"
# shellcheck disable=SC2155
export REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /registry.localhost)
echo "REGISTRY_IP=${REGISTRY_IP}"
# shellcheck disable=SC2002
# shellcheck disable=SC2086
kubectl -n kube-system patch cm coredns --patch "$(cat ${SCRIPTS_DIR}/resources/k3s-coredns-patch.tpl.yaml | envsubst)"
