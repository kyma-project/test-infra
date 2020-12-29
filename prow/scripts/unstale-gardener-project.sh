#!/usr/bin/env bash

# The purpose of this script is to create a resource in the Gardener project, then remove it.
# When the project is not active for some time it becomes scheduled for removal.
# This script unsets the project to removal.

set -eu

KYMA_PROJECT_DIR="/home/prow/go/src/github.com/kyma-project"

#shellcheck source=prow/scripts/lib/log.sh
source "$KYMA_PROJECT_DIR/test-infra/prow/scripts/lib/log.sh"
#shellcheck source=prow/scripts/lib/kyma.sh
source "$KYMA_PROJECT_DIR/test-infra/prow/scripts/lib/kyma.sh"

log::info "Install Kyma CLI"
kyma::install_cli

log::info "Provision Gardener cluster in GCP"
RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c4)
CLUSTER_NAME="nghbrs$RANDOM_NAME_SUFFIX"
kyma provision gardener gcp \
        --secret "${GARDENER_PROVIDER_SECRET_NAME}" --name "${CLUSTER_NAME}" \
        --project "${GARDENER_PROJECT_NAME}" --credentials "${GARDENER_KUBECONFIG}" \
        --region "${GARDENER_REGION}" -z "${GARDENER_ZONES}" -t "${MACHINE_TYPE}" \
        --scaler-max 4 --scaler-min 2

log::info "Cluster provisioned. Now deleting it..."

readonly NAMESPACE="garden-${GARDENER_PROJECT_NAME}"
kubectl --kubeconfig "${GARDENER_KUBECONFIG}" -n "${NAMESPACE}" annotate shoot "${CLUSTER_NAME}" confirmation.gardener.cloud/deletion=true --overwrite
kubectl --kubeconfig "${GARDENER_KUBECONFIG}" -n "${NAMESPACE}" delete shoot "${CLUSTER_NAME}" --wait=false

log::success "Done! See you next time!"
