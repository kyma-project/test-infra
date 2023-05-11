#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"

export KYMA_SOURCES_DIR="./kyma"
export COLLECTOR_NAMESPACE="serverless-integration"
export TEST_JOB_NAME="serverless-tests"


# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"
# shellcheck source=prow/scripts/lib/serverless-shared-k3s.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/serverless-shared-k3s.sh"
# shellcheck source=prow/scripts/cluster-integration/kyma-serverless-run-test.sh
source "${SCRIPT_DIR}/kyma-serverless-run-test.sh"



requiredVars=(
    CLUSTER_PROVIDER
	INPUT_CLUSTER_NAME
	KYMA_PROJECT_DIR
)

utils::check_required_vars "${requiredVars[@]}"

function connect_to_azure_cluster() {
    # shellcheck source=prow/scripts/lib/azure.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/azure.sh"

    requiredVars=(
        AZURE_CREDENTIALS_FILE
        RS_GROUP
    )
    utils::check_required_vars "${requiredVars[@]}"

    az::authenticate \
        -f "$AZURE_CREDENTIALS_FILE"
    az aks get-credentials --resource-group "${RS_GROUP}" --name "${INPUT_CLUSTER_NAME}"
}

function connect_to_gcp_cluster() {
    # shellcheck source=prow/scripts/lib/gcp.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcp.sh"

    requiredVars=(
        CLOUDSDK_CORE_PROJECT
        CLOUDSDK_COMPUTE_REGION
        CLOUDSDK_COMPUTE_ZONE
        GOOGLE_APPLICATION_CREDENTIALS
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
        log::error "GARDENER_PROVIDER ${CLUSTER_PROVIDER} is not yet supported"
        exit 1
    fi
}

function create_collector_namespace() {
    log::info "Creating collector namespace"
    kubectl create ns "${COLLECTOR_NAMESPACE}"
    kubectl label ns "${COLLECTOR_NAMESPACE}" created-by=serverless-controller-manager-test
    git checkout -
    popd
}

function run_serverless_metrics_collector() {
    log::info "Collecting serverless controller metrics"
    kubectl -n "${COLLECTOR_NAMESPACE}" apply -f "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/serverless-metrics-collector.yaml"
    kubectl -n "${COLLECTOR_NAMESPACE}" wait job/metrics-collector --for=condition=Complete=True --timeout=300s
    echo
    kubectl logs -n "${COLLECTOR_NAMESPACE}" -l job-name=metrics-collector
    echo
}

function clean_serverless_integration_tests() {
    log::info "Removing test namespace"
    kubectl delete ns -l created-by=serverless-controller-manager-test
}

connect_to_cluster
# in case of failed runs
clean_serverless_integration_tests

create_collector_namespace

set +o errexit
run_tests "${INTEGRATION_SUITE}" "${KYMA_SOURCES_DIR}"
TEST_STATUS=$?
set -o errexit

collect_results
run_serverless_metrics_collector

clean_serverless_integration_tests
echo "Exit code ${TEST_STATUS}"

exit ${TEST_STATUS}
