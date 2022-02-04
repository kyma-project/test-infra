#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"

export TEST_NAMESPACE="serverless-integration"
export TEST_JOB_NAME="serverless-tests"


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

function run_serverless_integration_tests() {
    log::info "Running Serverless Integration tests"
    pushd /home/prow/go/src/github.com/kyma-project/kyma/resources/serverless/

    log::info "Creating test namespace"
    kubectl create ns "${TEST_NAMESPACE}"
    kubectl label ns "${TEST_NAMESPACE}" created-by=serverless-controller-manager-test

    helm install serverless-test "charts/k3s-tests" -n "${TEST_NAMESPACE}" \
        -f values.yaml --set jobName="${TEST_JOB_NAME}" \
        --set testSuite="serverless-integration"

    git checkout -
    popd
}

function run_serverless_metrics_collector() {
    log::info "Collecting serverless controller metrics"
    kubectl -n "${TEST_NAMESPACE}" apply -f "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/serverless-metrics-collector.yaml"
    kubectl -n "${TEST_NAMESPACE}" wait job/metrics-collector --for=condition=Complete=True --timeout=300s
    echo
    kubectl logs -n "${TEST_NAMESPACE}" -l job-name=metrics-collector
    echo
}

function clean_serverless_integration_tests() {
    log::info "Removing test namespace"
    kubectl delete ns -l created-by=serverless-controller-manager-test
}

connect_to_cluster
# in case of failed runs
clean_serverless_integration_tests

run_serverless_integration_tests
job_status=""
# helm does not wait for jobs to complete even with --wait
# TODO but helm@v3.5 has a flag that enables that, get rid of this function once we use helm@v3.5
while true; do
    echo "Test job not completed yet..."
    [[ $(kubectl -n "${TEST_NAMESPACE}" get jobs "${TEST_JOB_NAME}" -o jsonpath='{.status.conditions[?(@.type=="Failed")].status}') == "True" ]] && job_status=1 && echo "Test job failed" && break
    [[ $(kubectl -n "${TEST_NAMESPACE}" get jobs "${TEST_JOB_NAME}" -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}') == "True" ]] && job_status=0 && echo "Test job completed successfully" && break
    sleep 5
done

collect_results "${TEST_JOB_NAME}" "${TEST_NAMESPACE}"
run_serverless_metrics_collector


clean_serverless_integration_tests


echo "Exit code ${job_status}"

exit $job_status
