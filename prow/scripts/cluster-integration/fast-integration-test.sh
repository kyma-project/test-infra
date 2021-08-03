#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"

# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"

# TODO check what is necessary
requiredVars=(
    CLUSTER_PROVIDER
	INPUT_CLUSTER_NAME
	KYMA_PROJECT_DIR
	CLOUDSDK_CORE_PROJECT
	CLOUDSDK_COMPUTE_REGION
	CLOUDSDK_COMPUTE_ZONE
	GOOGLE_APPLICATION_CREDENTIALS
)

utils::check_required_vars "${requiredVars[@]}"

# get cluster kubeconfig
if [[ $CLUSTER_PROVIDER == "azure" ]]; then
    # shellcheck source=prow/scripts/lib/azure.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/azure.sh"

    az::authenticate \
        -f "$AZURE_CREDENTIALS_FILE"
    az aks get-credentials --resource-group "${RS_GROUP}" --name "${INPUT_CLUSTER_NAME}"

elif [[ $CLUSTER_PROVIDER == "gcp" ]]; then
    # shellcheck source=prow/scripts/lib/gcp.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcp.sh"

    log::info "Authenticate"
    gcp::authenticate \
        -c "${GOOGLE_APPLICATION_CREDENTIALS}"

    log::info "get kubeconfig"

    gcp::get_cluster_kubeconfig \
    -c "$COMMON_NAME" \
    -p "$CLOUDSDK_CORE_PROJECT" \
    -z "$CLOUDSDK_COMPUTE_ZONE" \
    -R "$CLOUDSDK_COMPUTE_REGION" \
    -r "$PROVISION_REGIONAL_CLUSTER"
else
    ## TODO what should I put here? Is this a backend?
    log::error "GARDENER_PROVIDER ${CLUSTER_PROVIDER} is not yet supported"
    exit 1
fi

log::info "Running Kyma Fast Integration tests"

pushd /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration
# make ci-no-install
log:info "fast-ingetgartion goes here"
popd

log::success "Tests completed"
