#!/usr/bin/env bash

#Permissions: In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
# - Compute Admin
# - Service Account User
# - Service Account Admin
# - Service Account Token Creator
# - Make sure the service account is enabled for the Google Identity and Access Management API.

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gardener/gardener.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gardener.sh"

#!Put cleanup code in this function! Function is executed at exit from the script and on interuption.
gardener::cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        log::error "AN ERROR OCCURED! Take a look at preceding log entries."
    fi

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    # describe nodes to file in artifacts directory
    utils::describe_nodes

    if [ "${DEBUG_COMMANDO_OOM}" = "true" ]; then
      # copy output from debug container to artifacts directory
      utils::oom_get_output
    fi

    if  [[ "${CLEANUP_CLUSTER}" == "true" ]] ; then
        log::info "Deprovision cluster: \"${CLUSTER_NAME}\""
        gardener::deprovision_cluster \
            -p "${GARDENER_KYMA_PROW_PROJECT_NAME}" \
            -c "${CLUSTER_NAME}" \
            -f "${GARDENER_KYMA_PROW_KUBECONFIG}"
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
    if [ "${#CLUSTER_NAME}" -gt 9 ]; then
        log::error "Provided cluster name is too long"
        return 1
    fi

    CLEANUP_CLUSTER="true"
      # enable trap to catch kyma provision failures
      trap gardener::reprovision_cluster ERR
      # decreasing attempts to 2 because we will try to create new cluster from scratch on exit code other than 0
      kyma provision gardener aws \
        --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" \
        --name "${CLUSTER_NAME}" \
        --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" \
        --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
        --region "${GARDENER_REGION}" \
        -z "${GARDENER_ZONES}" \
        -t "${MACHINE_TYPE}" \
        --scaler-max 4 \
        --scaler-min 2 \
        --kube-version="${GARDENER_CLUSTER_VERSION}" \
        --attempts 1 \
        --verbose
    # trap cleanup we want other errors fail pipeline immediately
    trap - ERR
    if [ "$DEBUG_COMMANDO_OOM" = "true" ]; then
    # run oom debug pod
        utils::debug_oom
    fi
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

gardener::pre_upgrade_test_fast_integration_kyma() {
    log::info "Running pre-upgrade Kyma Fast Integration tests - AWS"

    kymaDirectory="$(utils::get_kyma_fast_integration_dir "$@")"
    log::info "Switching directory to '$kymaDirectory'"
    pushd "$kymaDirectory"
    make ci-pre-upgrade
    popd

    log::success "Tests completed"
}

gardener::post_upgrade_test_fast_integration_kyma() {
    log::info "Running post-upgrade Kyma Fast Integration tests - AWS"

    kymaDirectory="$(utils::get_kyma_fast_integration_dir "$@")"
    log::info "Switching directory to '$kymaDirectory'"
    pushd "$kymaDirectory"
    make ci-post-upgrade
    popd

    log::success "Tests completed"
}

gardener::test_kyma() {
    log::info "Running Kyma tests"

    readonly SUITE_NAME="testsuite-all-$(date '+%Y-%m-%d-%H-%M')"
    readonly CONCURRENCY=5
    set +e
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

    # collect logs from failed tests before deprovisioning
    kyma::run_test_log_collector "kyma-integration-gardener-aws"

    kyma::test_summary \
        -s "$SUITE_NAME"
    set -e
    return "${kyma_test_summary_return_exit_code:?}"
}
