#!/usr/bin/env bash

#Azure:
#Expected vars (additional to common vars):
# - RS_GROUP - azure resource group
# - REGION - azure region
# - AZURE_SUBSCRIPTION_ID
# - AZURE_SUBSCRIPTION_APP_ID
# - AZURE_SUBSCRIPTION_SECRET
# - AZURE_SUBSCRIPTION_TENANT
# - CLOUDSDK_CORE_PROJECT - required for cleanup of resources
#Permissions: In order to run this script you need to use an AKS service account with the contributor role

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/fluent-bit-stackdriver-logging.sh"

#!Put cleanup code in this function! Function is executed at exit from the script and on interuption.
gardener::cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?
    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        log::error "AN ERROR OCCURED! Take a look at preceding log entries."
    fi

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    log::info "Cleanup"
    set +e

    # describe nodes to file in artifacts directory
    utils::describe_nodes


    if [ "${DEBUG_COMMANDO_OOM}" = "true" ]; then
      # copy output from debug container to artifacts directory
      utils::oom_get_output
    fi

    if [ -n "${CLEANUP_CLUSTER}" ]; then
        if  [ -z "${CLEANUP_ONLY_SUCCEEDED}" ] || [[ -n "${CLEANUP_ONLY_SUCCEEDED}" && ${EXIT_STATUS} -eq 0 ]]; then
            log::info "Deprovision cluster: \"${CLUSTER_NAME}\""
            utils::deprovision_gardener_cluster "${GARDENER_KYMA_PROW_PROJECT_NAME}" "${CLUSTER_NAME}" "${GARDENER_KYMA_PROW_KUBECONFIG}"

            log::info "Deleting Azure EventHubs Namespace: \"${EVENTHUB_NAMESPACE_NAME}\""
            # Delete the Azure Event Hubs namespace which was created
            az eventhubs namespace delete -n "${EVENTHUB_NAMESPACE_NAME}" -g "${RS_GROUP}"
        fi
    fi

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    log::info "Job is finished ${MSG}"
    set -e

    exit "${EXIT_STATUS}"
}

gardener::init() {
    requiredVars=(
        JOB_TYPE
        KYMA_PROJECT_DIR
        GARDENER_REGION
        GARDENER_ZONES
        GARDENER_KYMA_PROW_KUBECONFIG
        GARDENER_KYMA_PROW_PROJECT_NAME
        GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
        RS_GROUP
        REGION
        AZURE_SUBSCRIPTION_ID
        AZURE_CREDENTIALS_FILE
        CLOUDSDK_CORE_PROJECT
        KYMA_SOURCE
    )

    utils::check_required_vars "${requiredVars[@]}"

    export INSTALLATION_OVERRIDE_STACKDRIVER="installer-config-logging-stackdiver.yaml"

    # we need to start the docker daemon
    docker::start

    EVENTHUB_NAMESPACE_NAME=""
    # Local variables
    if [[ -n "${PULL_NUMBER}" ]]; then  ### Creating name of the eventhub namespaces for pre-submit jobs
        EVENTHUB_NAMESPACE_NAME="pr-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}"
    else
        EVENTHUB_NAMESPACE_NAME="kyma-gardener-azure-${RANDOM_NAME_SUFFIX}"
    fi
    export EVENTHUB_NAMESPACE_NAME
}

gardener::set_machine_type() {
    if [ -z "${MACHINE_TYPE}" ]; then
        if [[ "$EXECUTION_PROFILE" == "evaluation" ]]; then
            export MACHINE_TYPE="Standard_D4_v3"
        else
            export MACHINE_TYPE="Standard_D8_v3"
        fi
    fi
}

gardener::generate_overrides() {
    log::info "Generate Azure Event Hubs overrides"

    EVENTHUB_SECRET_OVERRIDE_FILE=$(mktemp)
    export EVENTHUB_SECRET_OVERRIDE_FILE

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-azure-event-hubs-secret.sh
}

gardener::provision_cluster() {
    log::info "Provision cluster: \"${CLUSTER_NAME}\""

    CLEANUP_CLUSTER="true"
    set -x
    if [[ "$EXECUTION_PROFILE" == "evaluation" ]]; then
        kyma provision gardener az \
            --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" --name "${CLUSTER_NAME}" \
            --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
            --region "${GARDENER_REGION}" -z "${GARDENER_ZONES}" -t "${MACHINE_TYPE}" \
            --scaler-max 1 --scaler-min 1 \
            --disk-type StandardSSD_LRS \
            --kube-version="${GARDENER_CLUSTER_VERSION}" \
            --verbose
    else
        kyma provision gardener az \
            --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" --name "${CLUSTER_NAME}" \
            --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
            --region "${GARDENER_REGION}" -z "${GARDENER_ZONES}" -t "${MACHINE_TYPE}" \
            --disk-type StandardSSD_LRS \
            --kube-version="${GARDENER_CLUSTER_VERSION}" \
            --verbose
    fi
    set +x
}

gardener::install_kyma() {
    log::info "Installing Kyma"

    prepare_stackdriver_logging "${INSTALLATION_OVERRIDE_STACKDRIVER}"
    if [[ "$?" -ne 0 ]]; then
        return 1
    fi

    INSTALLATION_RESOURCES_DIR=${KYMA_SOURCES_DIR}/installation/resources

    set -x
    if [[ "$EXECUTION_PROFILE" == "evaluation" ]]; then
        kyma install \
            --ci \
            --source "${KYMA_SOURCE}" \
            -c "${INSTALLATION_RESOURCES_DIR}"/installer-cr-azure-eventhubs.yaml.tpl \
            -o "${INSTALLATION_RESOURCES_DIR}"/installer-config-azure-eventhubs.yaml.tpl \
            -o "${EVENTHUB_SECRET_OVERRIDE_FILE}" \
            -o "${INSTALLATION_OVERRIDE_STACKDRIVER}" \
            --timeout 60m \
            --profile evaluation \
            --verbose
    elif [[ "$EXECUTION_PROFILE" == "production" ]]; then
        kyma install \
            --ci \
            --source "${KYMA_SOURCE}" \
            -c "${INSTALLATION_RESOURCES_DIR}"/installer-cr-azure-eventhubs.yaml.tpl \
            -o "${INSTALLATION_RESOURCES_DIR}"/installer-config-azure-eventhubs.yaml.tpl \
            -o "${EVENTHUB_SECRET_OVERRIDE_FILE}" \
            -o "${INSTALLATION_OVERRIDE_STACKDRIVER}" \
            --timeout 60m \
            --profile production \
            --verbose
    else
        kyma install \
            --ci \
            --source "${KYMA_SOURCE}" \
            -c "${INSTALLATION_RESOURCES_DIR}"/installer-cr-azure-eventhubs.yaml.tpl \
            -o "${INSTALLATION_RESOURCES_DIR}"/installer-config-azure-eventhubs.yaml.tpl \
            -o "${EVENTHUB_SECRET_OVERRIDE_FILE}" \
            -o "${INSTALLATION_OVERRIDE_STACKDRIVER}" \
            --timeout 90m \
            --verbose
    fi
    set +x
}

gardener::hibernate_kyma() {
    local SAVED_KUBECONFIG=$KUBECONFIG
    export KUBECONFIG=$GARDENER_KYMA_PROW_KUBECONFIG

    log::info "Checking if cluster can be hibernated"
    HIBERNATION_POSSIBLE=$(kubectl get shoots "${CLUSTER_NAME}" -o jsonpath='{.status.constraints[?(@.type=="HibernationPossible")].status}')

    if [[ "$HIBERNATION_POSSIBLE" != "True" ]]; then
      log::error "Hibernation for this cluster is not possible! Please take a look at the constraints :"
      kubectl get shoots "${CLUSTER_NAME}}" -o jsonpath='{.status.constraints}'
      exit 1
    fi

    log::info "Cluster can be hibernated"

    log::info "Hibernating kyma cluster"
    kubectl patch shoots "${CLUSTER_NAME}" -p '{"spec": {"hibernation" : {"enabled" : true }}}'

    log::info "Checking state of kyma hibernation...hold on"

    local STATUS
    SECONDS=0
    local END_TIME=$((SECONDS+1000))
    while [ ${SECONDS} -lt ${END_TIME} ];do
        STATUS=$(kubectl get shoot "${CLUSTER_NAME}" -o jsonpath='{.status.hibernated}')
        if [ "$STATUS" == "true" ]; then
            log::info "Kyma is hibernated."
            break
        fi
        log::info "waiting 60s for operation to complete, kyma hibernated : ${STATUS}"
        sleep 60
    done
    if [ "$STATUS" != "true" ]; then
        log::error "Timeout. Kyma cluster is not hibernated after $SECONDS seconds"
        exit 1
    fi
    export KUBECONFIG=$SAVED_KUBECONFIG
}

pods_running(){
    list=$(kubectl get pods -n "$1" -o=jsonpath='{range .items[*]}{.status.phase}{"\n"}')
    if [[ -z $list ]]; then
      log::error "Failed to get pod list"
      return 1
    fi

    for status in $list
    do
        if [[ "$status" != "Running" && "$status" != "Succeeded" ]]; then
          return 1
        fi
    done

    return 0
}

check_pods_in_namespaces(){
    local namespaces=("$@")
    for ns in "${namespaces[@]}"; do
        log::info "checking pods in namespace : $ns"
        if ! pods_running "$ns"; then
            log::info "pods in $ns are still not running..."
            return 1
        fi
    done
    return 0
}

wait_for_pods_in_namespaces(){
    local namespaces=("$@")
    local done=1
    SECONDS=0
    local END_TIME=$((SECONDS+900))
    while [ ${SECONDS} -lt ${END_TIME} ];do
        if check_pods_in_namespaces "${namespaces[@]}"; then
            done=0
            break
        fi
        log::info "waiting for 30s"
        sleep 30
    done

    if [ ! $done ]; then
        log::error "Timeout exceeded, pods are still not running in required namespaces"
    else
        log::info "all pods in required namespaces are running"
    fi
}

gardener::wake_up_kyma() {
    local SAVED_KUBECONFIG=$KUBECONFIG
    export KUBECONFIG=$GARDENER_KYMA_PROW_KUBECONFIG

    log::info "Waking up kyma cluster"
    kubectl patch shoots "${CLUSTER_NAME}" -p '{"spec": {"hibernation" : {"enabled" : false }}}'

    log::info "Checking state of kyma waking up...hold on"

    local STATUS
    SECONDS=0
    local END_TIME=$((SECONDS+1200))
    while [ ${SECONDS} -lt ${END_TIME} ];do
        STATUS=$(kubectl get shoot "${CLUSTER_NAME}" -o jsonpath='{.status.hibernated}')
        if [ "$STATUS" == "false" ]; then
            log::info "Kyma is awake."
            break
        fi
        log::info "Waiting 60s for operation to complete, kyma cluster is waking up."
        sleep 60
    done
    if [ "$STATUS" != "false" ]; then
        log::error "Timeout. Kyma cluster is not awake after $SECONDS seconds"
        exit 1
    fi

    log::info "Waiting for pods to be running"
    export KUBECONFIG=$SAVED_KUBECONFIG

    local namespaces=("istio-system" "kyma-system" "kyma-integration")
    wait_for_pods_in_namespaces "${namespaces[@]}"
    date
}

gardener::test_fast_integration_kyma() {
    log::info "Running Kyma Fast Integration tests"

    pushd /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration
    make ci-no-install
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
        --timeout 120m \
        --watch \
        --non-interactive
    )

    # collect logs from failed tests before deprovisioning
    kyma::run_test_log_collector "kyma-integration-gardener-azure"
    if ! kyma::test_summary; then
      log::error "Tests have failed"
      set -e
      return 1
    fi
    set -e
    log::success "Tests completed"
}
