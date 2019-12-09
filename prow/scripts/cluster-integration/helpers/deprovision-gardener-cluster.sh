# !/bin/bash

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

RES=$(kubectl --kubeconfig ${GARDENER_CREDENTIALS} -n ${NAMESPACE} annotate shoot "${GARDENER_CLUSTER_NAME}" confirmation.garden.sapcloud.io/deletion=true --overwrite)
echo "Annotate shoot for deletion: ${RES}"
RES=$(kubectl --kubeconfig ${GARDENER_CREDENTIALS} -n ${NAMESPACE} delete   shoot "${GARDENER_CLUSTER_NAME}")
echo "Delete shoot CRD: ${RES}"