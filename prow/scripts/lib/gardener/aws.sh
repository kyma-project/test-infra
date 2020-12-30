#!/usr/bin/env bash

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

#!Put cleanup code in this function! Function is executed at exit from the script and on interuption.
gardener::cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        log::error "AN ERROR OCCURED! Take a look at preceding log entries."
    fi

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    # collect logs from failed tests before deprovisioning
    kyma::run_test_log_collector "kyma-integration-gardener-aws"

    if [[ -n "${SUITE_NAME}" ]]; then
        kyma::test_summary
        SUITE_EXIT_STATUS=$?
        if [[ ${EXIT_STATUS} -eq 0 ]]; then
            EXIT_STATUS=$SUITE_EXIT_STATUS
        fi
    fi

    if [ -n "${CLEANUP_CLUSTER}" ]; then
        log::info "Deprovision cluster: \"${CLUSTER_NAME}\""
        # Export envvars for the script
        export GARDENER_CLUSTER_NAME=${CLUSTER_NAME}
        export GARDENER_PROJECT_NAME=${GARDENER_KYMA_PROW_PROJECT_NAME}
        export GARDENER_CREDENTIALS=${GARDENER_KYMA_PROW_KUBECONFIG}
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/deprovision-gardener-cluster.sh
    fi

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    log::info "Job is finished ${MSG}"
    set -e

    exit "${EXIT_STATUS}"
}

gardener::init() {
    requiredVars=(
        KYMA_PROJECT_DIR
        GARDENER_REGION
        GARDENER_ZONES
        GARDENER_KYMA_PROW_KUBECONFIG
        GARDENER_KYMA_PROW_PROJECT_NAME
        GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
        KYMA_SOURCE
    )

    utils::check_required_vars "${requiredVars[@]}"
}

gardener::set_machine_type() {
    if [ -z "$MACHINE_TYPE" ]; then
        export MACHINE_TYPE="m5.xlarge"
    fi
}

gardener::generate_overrides() {
    # currently only Azure generates anything in this function
    return
}

gardener::provision_cluster() {
    log::info "Provision cluster: \"${CLUSTER_NAME}\""

    CLEANUP_CLUSTER="true"
    (
    set -x
    kyma provision gardener aws \
            --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" --name "${CLUSTER_NAME}" \
            --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
            --region "${GARDENER_REGION}" -z "${GARDENER_ZONES}" -t "${MACHINE_TYPE}" \
            --scaler-max 4 --scaler-min 2 \
            --kube-version="${GARDENER_CLUSTER_VERSION}"
    )
}

gardener::install_kyma() {
    log::info "Installing Kyma"
    (
    set -x
    kyma install \
        --ci \
        --source "${KYMA_SOURCE}" \
        --timeout 90m
    )
}

gardener::hibernate_kyma() {
    return
}

gardener::wake_up_kyma() {
    return
}

gardener::test_fast_integration_kyma() {
    return
}

gardener::test_kyma() {
    log::info "Running Kyma tests"

    readonly SUITE_NAME="testsuite-all-$(date '+%Y-%m-%d-%H-%M')"
    readonly CONCURRENCY=5
    (
    set -x
    kyma test run \
        --name "${SUITE_NAME}" \
        --concurrency "${CONCURRENCY}" \
        --max-retries 1 \
        --timeout 90m \
        --watch \
        --non-interactive
    )

    log::success "Tests completed"
}
