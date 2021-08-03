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
	GOOGLE_APPLICATION_CREDENTIALS
)

utils::check_required_vars "${requiredVars[@]}"

function connect_to_azure_cluster() {
    # shellcheck source=prow/scripts/lib/azure.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/azure.sh"

    az::authenticate \
        -f "$AZURE_CREDENTIALS_FILE"

    # az::set_subscription -s "$AZURE_SUBSCRIPTION_ID"
    az aks get-credentials --resource-group "${RS_GROUP}" --name "${INPUT_CLUSTER_NAME}"
}

function connect_to_gcp_cluster() {
    # shellcheck source=prow/scripts/lib/gcp.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcp.sh"

    requiredVars=(
        CLOUDSDK_CORE_PROJECT
        CLOUDSDK_COMPUTE_REGION
        CLOUDSDK_COMPUTE_ZONE
    )
    utils::check_required_vars "${requiredVars[@]}"

    log::info "Authenticate"
    gcp::authenticate \
        -c "${GOOGLE_APPLICATION_CREDENTIALS}"

    log::info "get kubeconfig"

    gcp::get_cluster_kubeconfig \
    -c "$INPUT_CLUSTER_NAME" \
    -p "$CLOUDSDK_CORE_PROJECT" \
    -z "$CLOUDSDK_COMPUTE_ZONE" \
    -R "$CLOUDSDK_COMPUTE_REGION" \
    -r "$PROVISION_REGIONAL_CLUSTER"
}

function connect_to_cluster() {
    if [[ $CLUSTER_PROVIDER == "azure" ]]; then
        connect_to_azure_cluster
    elif [[ $CLUSTER_PROVIDER == "gcp" ]]; then
        connect_to_gcp_cluster
    else
        ## TODO what should I put here? Is this a backend?
        log::error "GARDENER_PROVIDER ${CLUSTER_PROVIDER} is not yet supported"
        exit 1
    fi
}

function test_fast_integration_kyma() {
    log::info "Running Kyma Fast Integration tests"

    pushd /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration
    make ci-no-install
    popd
}

connect_to_cluster

test_fast_integration_kyma

log::success "Tests completed"
