#!/usr/bin/env bash

###
# The following script deprovisions a Gardener cluster.
#
# INPUT:
# - GARDENER_PROJECT_NAME
# - GARDENER_CLUSTER_NAME
# - GARDENER_CREDENTIALS
#
# REQUIREMENTS:
# - kubectl
###

readonly NAMESPACE="garden-${GARDENER_PROJECT_NAME}"

kubectl --kubeconfig "${GARDENER_CREDENTIALS}" -n "${NAMESPACE}" annotate shoot "${GARDENER_CLUSTER_NAME}" confirmation.gardener.cloud/deletion=true --overwrite
kubectl --kubeconfig "${GARDENER_CREDENTIALS}" -n "${NAMESPACE}" delete shoot "${GARDENER_CLUSTER_NAME}" --wait=false
