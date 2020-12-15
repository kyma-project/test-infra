#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on a real Gardener cluster.
#
#Expected vars:
#
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - DOCKER_PUSH_REPOSITORY - Docker repository hostname
# - DOCKER_PUSH_DIRECTORY - Docker "top-level" directory (with leading "/")
# - GARDENER_REGION - Gardener compute region
# - GARDENER_ZONES - Gardener compute zones inside the region
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME Name of the azure secret configured in the gardener project to access the cloud provider
# - MACHINE_TYPE (optional): AKS machine type
# - RS_GROUP - azure resource group
# - REGION - azure region
# - AZURE_SUBSCRIPTION_ID
# - AZURE_SUBSCRIPTION_APP_ID
# - AZURE_SUBSCRIPTION_SECRET
# - AZURE_SUBSCRIPTION_TENANT
# - CLOUDSDK_CORE_PROJECT - required for cleanup of resources
#
#Permissions: In order to run this script you need to use an AKS service account with the contributor role

set -e

discoverUnsetVar=false
ENABLE_TEST_LOG_COLLECTOR=false

VARIABLES=(
    JOB_TYPE
    KYMA_PROJECT_DIR
    DOCKER_PUSH_REPOSITORY
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
)

if [[ "$JOB_TYPE" == "presubmit" ]]; then
    VARIABLES+=( DOCKER_PUSH_DIRECTORY )
fi

for var in "${VARIABLES[@]}"; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

if [[ "${BUILD_TYPE}" == "master" ]]; then
    if [ -z "${LOG_COLLECTOR_SLACK_TOKEN}" ] ; then
        echo "ERROR: LOG_COLLECTOR_SLACK_TOKEN is not set"
        exit 1
    fi
fi

readonly GARDENER_CLUSTER_VERSION="1.16"

#Exported variables
export RS_GROUP \
    REGION \
    AZURE_SUBSCRIPTION_ID \
    AZURE_SUBSCRIPTION_APP_ID \
    AZURE_SUBSCRIPTION_SECRET \
    AZURE_SUBSCRIPTION_TENANT
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export INSTALLATION_OVERRIDE_STACKDRIVER="installer-config-logging-stackdiver.yaml"

TMP_DIR=$(mktemp -d)

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/testing-helpers.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/kyma-cli.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/fluent-bit-stackdriver-logging.sh"
set -o

# we need to start the docker daemon. This is done by calling init from the library.sh
init

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
readonly COMMON_NAME_PREFIX="grd"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")

### Cluster name must be less than 10 characters!
export CLUSTER_NAME="${COMMON_NAME}"
EVENTHUB_NAMESPACE_NAME=""
# Local variables
if [[ -n "${PULL_NUMBER}" ]]; then  ### Creating name of the eventhub namespaces for pre-submit jobs
    EVENTHUB_NAMESPACE_NAME="pr-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}"
else
    EVENTHUB_NAMESPACE_NAME="kyma-gardener-azure-${RANDOM_NAME_SUFFIX}"
fi
export EVENTHUB_NAMESPACE_NAME

echo "Execution profile: ${EXECUTION_PROFILE}"

if [ -z "${MACHINE_TYPE}" ]; then
    if [[ "$EXECUTION_PROFILE" == "evaluation" ]]; then
        export MACHINE_TYPE="Standard_D4_v3"
    else
        export MACHINE_TYPE="Standard_D8_v3"
    fi
fi

#!Put cleanup code in this function! Function is executed at exit from the script and on interuption.
cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?
    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        shout "AN ERROR OCCURED! Take a look at preceding log entries."
        echo
    fi

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    shout "Cleanup"
    set +e

    shout "Describe nodes"
    kubectl describe nodes
    kubectl top nodes
    kubectl top pods --all-namespaces

    # collect logs from failed tests before deprovisioning
    runTestLogCollector

    if [[ -n "${SUITE_NAME}" ]]; then
        testSummary
        SUITE_EXIT_STATUS=$?
        if [[ ${EXIT_STATUS} -eq 0 ]]; then
            EXIT_STATUS=$SUITE_EXIT_STATUS
        fi
    fi


    if [ -n "${CLEANUP_CLUSTER}" ]; then
        shout "Deprovision cluster: \"${CLUSTER_NAME}\""
        date
        # Export envvars for the script
        export GARDENER_CLUSTER_NAME=${CLUSTER_NAME}
        export GARDENER_PROJECT_NAME=${GARDENER_KYMA_PROW_PROJECT_NAME}
        export GARDENER_CREDENTIALS=${GARDENER_KYMA_PROW_KUBECONFIG}
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/deprovision-gardener-cluster.sh

        shout "Deleting Azure EventHubs Namespace: \"${EVENTHUB_NAMESPACE_NAME}\""
        # Delete the Azure Event Hubs namespace which was created
        az eventhubs namespace delete -n "${EVENTHUB_NAMESPACE_NAME}" -g "${RS_GROUP}"
    fi

    if [ -n "${CLEANUP_DOCKER_IMAGE}" ]; then
        shout "Delete temporary Kyma-Installer Docker image"
        date
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/delete-image.sh"
    fi

    rm -rf "${TMP_DIR}"

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    shout "Job is finished ${MSG}"
    date
    set -e

    exit "${EXIT_STATUS}"
}

testSummary() {
    local tests_exit=0
    echo "Test Summary"
    kyma test status "${SUITE_NAME}" -owide

    statusSucceeded=$(kubectl get cts "${SUITE_NAME}" -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
    if [[ "${statusSucceeded}" != *"True"* ]]; then
        tests_exit=1
        echo "- Fetching logs due to test suite failure"

        echo "- Fetching logs from testing pods in Failed status..."
        kyma test logs "${SUITE_NAME}" --test-status Failed

        echo "- Fetching logs from testing pods in Unknown status..."
        kyma test logs "${SUITE_NAME}" --test-status Unknown

        echo "- Fetching logs from testing pods in Running status due to running afer test suite timeout..."
        kyma test logs "${SUITE_NAME}" --test-status Running
    fi

    echo "ClusterTestSuite details"
    kubectl get cts "${SUITE_NAME}" -oyaml
    return $tests_exit
}

install_cli(){
    export INSTALL_DIR=${TMP_DIR}
    install::kyma_cli
}

generate_azure_overrides() {
    shout "Generate Azure Event Hubs overrides"
    date

    EVENTHUB_SECRET_OVERRIDE_FILE=$(mktemp)
    export EVENTHUB_SECRET_OVERRIDE_FILE

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-azure-event-hubs-secret.sh
}

provision_cluster() {
    shout "Provision cluster: \"${CLUSTER_NAME}\""

    CLEANUP_CLUSTER="true"
    set -x
    if [[ "$EXECUTION_PROFILE" == "evaluation" ]]; then
        kyma provision gardener az \
            --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" --name "${CLUSTER_NAME}" \
            --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
            --region "${GARDENER_REGION}" -z "${GARDENER_ZONES}" -t "${MACHINE_TYPE}" \
            --scaler-max 1 --scaler-min 1 \
            --disk-type StandardSSD_LRS \
            --kube-version=${GARDENER_CLUSTER_VERSION} \
            --verbose
    else
        kyma provision gardener az \
            --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" --name "${CLUSTER_NAME}" \
            --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
            --region "${GARDENER_REGION}" -z "${GARDENER_ZONES}" -t "${MACHINE_TYPE}" \
            --disk-type StandardSSD_LRS \
            --kube-version=${GARDENER_CLUSTER_VERSION} \
            --verbose
    fi
    set +x
}

build_image() {
    if [[ "$JOB_TYPE" == "presubmit" ]]; then
        shout "Build Kyma-Installer Docker image"
        date
        CLEANUP_DOCKER_IMAGE="true"
        export KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gardener-azure-integration/${REPO_OWNER}/${REPO_NAME}:PR-${PULL_NUMBER}"

        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-image.sh"
    else
        KYMA_INSTALLER_IMAGE=master
    fi
}

install_kyma() {
    shout "Installing Kyma"
    date

    prepare_stackdriver_logging "${INSTALLATION_OVERRIDE_STACKDRIVER}"
    if [[ "$?" -ne 0 ]]; then
        return 1
    fi

    INSTALLATION_RESOURCES_DIR=${KYMA_SOURCES_DIR}/installation/resources

    set -x
    if [[ "$EXECUTION_PROFILE" == "evaluation" ]]; then
        kyma install \
            --ci \
            --source "${KYMA_INSTALLER_IMAGE}" \
            -c "${INSTALLATION_RESOURCES_DIR}"/installer-cr-azure-eventhubs.yaml.tpl \
            -o "${INSTALLATION_RESOURCES_DIR}"/installer-config-azure-eventhubs.yaml.tpl \
            -o "${EVENTHUB_SECRET_OVERRIDE_FILE}" \
            --timeout 35m \
            --profile evaluation \
            --verbose
    else
        kyma install \
            --ci \
            --source "${KYMA_INSTALLER_IMAGE}" \
            -c "${INSTALLATION_RESOURCES_DIR}"/installer-cr-azure-eventhubs.yaml.tpl \
            -o "${INSTALLATION_RESOURCES_DIR}"/installer-config-production.yaml.tpl \
            -o "${INSTALLATION_RESOURCES_DIR}"/installer-config-azure-eventhubs.yaml.tpl \
            -o "${EVENTHUB_SECRET_OVERRIDE_FILE}" \
            -o "${INSTALLATION_OVERRIDE_STACKDRIVER}" \
            --timeout 90m \
            --verbose
    fi
    set +x

    shout "Checking the versions"
    date
    kyma version
}

test_kyma(){
    shout "Running Kyma tests"
    date

    readonly SUITE_NAME="testsuite-all-$(date '+%Y-%m-%d-%H-%M')"
    readonly CONCURRENCY=5

    set -x
    kyma test run \
        --name "${SUITE_NAME}" \
        --concurrency "${CONCURRENCY}" \
        --max-retries 1 \
        --timeout 120m \
        --watch \
        --non-interactive
    set +x

    date
    shout "Tests completed"
}


hibernate_kyma(){
    shout "Checking if cluster can be hibernated"
    HIBERNATION_POSSIBLE=$(kubectl get shoots "${CLUSTER_NAME}" -o jsonpath='{.status.constraints[?(@.type=="HibernationPossible")].status}')

    if [[ "$HIBERNATION_POSSIBLE" != "True" ]]; then
      echo "Hibenration for this cluster is not possible! Please take a look at the constraints :"
      kubectl get shoots "${CLUSTER_NAME}}" -o jsonpath='{.status.constraints}'
      exit 1
    fi

    echo "Cluster can be hibernated"

    local SAVED_KUBECONFIG=$KUBECONFIG
    export KUBECONFIG=$GARDENER_KYMA_PROW_KUBECONFIG

    shout "Hibernating kyma cluster"
    date
    kubectl patch shoots "${CLUSTER_NAME}" -p '{"spec": {"hibernation" : {"enabled" : true }}}'

    shout "Checking state of kyma hibernation...hold on"

    local STATUS
    SECONDS=0
    local END_TIME=$((SECONDS+1000))
    while [ ${SECONDS} -lt ${END_TIME} ];do
        STATUS=$(kubectl get shoot "${CLUSTER_NAME}" -o jsonpath='{.status.hibernated}')
        if [ "$STATUS" == "true" ]; then
            echo "Kyma is hibernated."
            break
        fi
        echo "waiting 60s for operation to complete, kyma hibernated : ${STATUS}"
        sleep 60
    done
    if [ "$STATUS" != "true" ]; then
        echo "Timeout. Kyma cluster is not hibernated after $SECONDS seconds"
        exit 1
    fi
    export KUBECONFIG=$SAVED_KUBECONFIG
}

wake_up_kyma(){
    local SAVED_KUBECONFIG=$KUBECONFIG
    export KUBECONFIG=$GARDENER_KYMA_PROW_KUBECONFIG

    shout "Waking up kyma cluster"
    date
    kubectl patch shoots "${CLUSTER_NAME}" -p '{"spec": {"hibernation" : {"enabled" : false }}}'

    shout "Checking state of kyma waking up...hold on"

    local STATUS
    SECONDS=0
    local END_TIME=$((SECONDS+1000))
    while [ ${SECONDS} -lt ${END_TIME} ];do
        STATUS=$(kubectl get shoot "${CLUSTER_NAME}" -o jsonpath='{.status.hibernated}')
        if [ "$STATUS" == "false" ]; then
            date
            echo "Kyma is awake."
            break
        fi
        echo "Waiting 60s for operation to complete, kyma cluster is waking up."
        sleep 60
    done
    if [ "$STATUS" != "false" ]; then
        echo "Timeout. Kyma cluster is not awake after $SECONDS seconds"
        exit 1
    fi

    shout "Waiting for pods to be running"
    date
    export KUBECONFIG=$SAVED_KUBECONFIG

    local namespaces=("istio-system" "kyma-system" "kyma-integration")
    wait_for_pods_in_namespaces "${namespaces[@]}"
    date
}

pods_running(){
    list=$(kubectl get pods -n "$1" -o=jsonpath='{range .items[*]}{.status.phase}{"\n"}')
    if [[ -z $list ]]; then
      echo "Failed to get pod list"
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
        echo "checking pods in namespace : $ns"
        if ! pods_running "$ns"; then
            echo "pods in $ns are still not running..."
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
        echo "waiting for 30s"
        sleep 30
    done

    if [ ! $done ]; then
        echo "Timeout exceeded, pods are still not running in required namespaces"
    else
        echo "all pods in required namespaces are running"
    fi
}

test_local_kyma(){
    shout "Running Kyma tests (from local-kyma repo)"
    date

    pushd /home/prow/go/src/github.com/kyma-incubator/local-kyma
    shout "Executing app-connector-example"
    ./app-connector-example.sh
    popd

    date
    shout "Tests completed"
}

test_fast_integration_kyma() {
    shout "Running Kyma Fast Integration tests"
    date

    pushd /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration
    npm install
    npm test
    popd

    shout "Tests completed"
    date
}

runTestLogCollector(){
    if [ "${ENABLE_TEST_LOG_COLLECTOR}" = true ] ; then
        if [[ "$BUILD_TYPE" == "master" ]] || [[ -z "$BUILD_TYPE" ]]; then
            shout "Install test-log-collector"
            date
            export PROW_JOB_NAME="kyma-integration-gardener-azure"
            (
                "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/install-test-log-collector.sh" || true # we want it to work on "best effort" basis, which does not interfere with cluster
            )
        fi
    fi
}

trap cleanup EXIT INT

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

install_cli

generate_azure_overrides

provision_cluster

build_image

install_kyma
shout "Describe nodes (after installation)"

set +e
kubectl describe nodes
kubectl top nodes
kubectl top pods --all-namespaces
set -e

if [[ "$?" -ne 0 ]]; then
    return 1
fi

if [[ "$HIBERNATION_ENABLED" == "true" ]]; then
    hibernate_kyma
    sleep 120
    wake_up_kyma
fi

if [[ "$FAST_INTEGRATION_TESTS" == "true" ]]; then
    test_fast_integration_kyma
elif [[ "$EXECUTION_PROFILE" == "evaluation" ]]; then
    test_local_kyma
else
    ENABLE_TEST_LOG_COLLECTOR=true # enable test-log-collector before tests; if prowjob fails before test phase we do not have any reason to enable it earlier
    test_kyma
fi

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
