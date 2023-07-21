#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_DIR="${KYMA_PROJECT_DIR}/kyma"

export TEST_NAMESPACE="serverless-benchmarks"
export ALL_FUNCTIONS=(nodejs18-xs nodejs18-s nodejs18-m nodejs18-l nodejs18-xl python39-s python39-m python39-l python39-xl)


# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"
# shellcheck source=prow/scripts/lib/serverless-shared-k3s.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/serverless-shared-k3s.sh"

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

function run_serverless_test_function() {
    log::info "Deploying test functions"
    kubectl -n "${TEST_NAMESPACE}" apply -f \
        "${KYMA_DIR}/tests/serverless-bench/fixtures/functions/"

    log::info "Waiting for test functions to be Running"
    for FUNCTION in "${ALL_FUNCTIONS[@]}"; do
        kubectl -n "${TEST_NAMESPACE}" wait "function/${FUNCTION}" \
            --for=condition=Running=True --timeout 2m
    done
    log::info "Done"
}

function collect_benchmark_results() {
    log::info "Running benchmarks and collecting results"
    kubectl -n "${TEST_NAMESPACE}" apply -f \
        "${KYMA_DIR}/tests/serverless-bench/fixtures/serverless-benchmark-job.yaml"
    kubectl -n "${TEST_NAMESPACE}" wait job/serverless-benchmark \
        --for=condition=Complete=True --timeout=20m
    kubectl -n "${TEST_NAMESPACE}" logs -l jobName=serverless-benchmark --tail=-1
}

function clean_serverless_integration_tests() {
    log::info "Removing test namespace"
    kubectl delete ns -l created-by=serverless-benchmarks
}

connect_to_cluster
# in case of failed runs
clean_serverless_integration_tests

log::info "Creating test namespace"
kubectl create ns "${TEST_NAMESPACE}"
kubectl label ns "${TEST_NAMESPACE}" created-by=serverless-benchmarks
run_serverless_test_function

collect_benchmark_results

job_status=""
[[ $(kubectl -n "${TEST_NAMESPACE}" get jobs serverless-benchmark -o jsonpath='{.status.conditions[?(@.type=="Failed")].status}') == "True" ]] && job_status=1
[[ $(kubectl -n "${TEST_NAMESPACE}" get jobs serverless-benchmark -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}') == "True" ]] && job_status=0


log::info "Cleaning up test resources"
# clean_serverless_integration_tests

echo "Exit code ${job_status}"
exit $job_status
