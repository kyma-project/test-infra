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
# - EVENTHUB_NAMESPACE_NAME
# - AZURE_SUBSCRIPTION_ID
# - AZURE_SUBSCRIPTION_APP_ID
# - AZURE_SUBSCRIPTION_SECRET
# - AZURE_SUBSCRIPTION_TENANT
#
#Permissions: In order to run this script you need to use an AKS service account with the contributor role

set -e

discoverUnsetVar=false

VARIABLES=(
    JOB_TYPE
    KYMA_PROJECT_DIR
    DOCKER_PUSH_REPOSITORY
    DOCKER_PUSH_DIRECTORY
    GARDENER_REGION
    GARDENER_ZONES
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
    RS_GROUP
    REGION
    AZURE_SUBSCRIPTION_ID
    AZURE_SUBSCRIPTION_APP_ID
    AZURE_SUBSCRIPTION_SECRET
    AZURE_SUBSCRIPTION_TENANT
    CLOUDSDK_CORE_PROJECT
)

for var in "${VARIABLES[@]}"; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
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

TMP_DIR=$(mktemp -d)

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/testing-helpers.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/kyma-cli.sh"
set -o

# we need to start the docker daemon. This is done by calling init from the library.sh
init

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

        shout "Deleting Azure Resource Group: \"${RS_GROUP}\""
        # Delete the Azure Resource Group
        az group delete -n "${RS_GROUP}" -y
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

trap cleanup EXIT INT

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

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

export INSTALL_DIR=${TMP_DIR}
install::kyma_cli

shout "Generate Azure Event Hubs overrides"
date

EVENTHUB_SECRET_OVERRIDE_FILE=$(mktemp)
export EVENTHUB_SECRET_OVERRIDE_FILE

"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-azure-event-hubs-secret.sh
cat "${EVENTHUB_SECRET_OVERRIDE_FILE}"

shout "Provision cluster: \"${CLUSTER_NAME}\""

if [ -z "$MACHINE_TYPE" ]; then
      export MACHINE_TYPE="Standard_D8_v3"
fi


CLEANUP_CLUSTER="true"
set -x
kyma provision gardener az \
    --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" --name "${CLUSTER_NAME}" \
    --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
    --region "${GARDENER_REGION}" -z "${GARDENER_ZONES}" -t "${MACHINE_TYPE}" \
    --scaler-max 4 --scaler-min 3 \
    --kube-version=${GARDENER_CLUSTER_VERSION} \
    --verbose
set +x


if [[ "$JOB_TYPE" == "presubmit" ]]; then
    shout "Build Kyma-Installer Docker image"
    date
    CLEANUP_DOCKER_IMAGE="true"
    export KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gardener-azure-integration/${REPO_OWNER}/${REPO_NAME}:PR-${PULL_NUMBER}"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-image.sh"
else
    KYMA_INSTALLER_IMAGE=latest-published
fi

shout "Installing Kyma"
date

INSTALLATION_RESOURCES_DIR=${KYMA_SOURCES_DIR}/installation/resources
set -x
kyma install \
    --ci \
    --source $KYMA_INSTALLER_IMAGE \
    -c "${INSTALLATION_RESOURCES_DIR}"/installer-cr-azure-eventhubs.yaml.tpl \
    -o "${INSTALLATION_RESOURCES_DIR}"/installer-config-production.yaml.tpl \
    -o "${EVENTHUB_SECRET_OVERRIDE_FILE}" \
    --timeout 90m \
    --verbose
set +x

shout "Checking the versions"
date
kyma version

shout "Running Kyma tests"
date

readonly SUITE_NAME="testsuite-all-$(date '+%Y-%m-%d-%H-%M')"
readonly CONCURRENCY=5

set -x
kyma test run \
    --name "${SUITE_NAME}" \
    --concurrency "${CONCURRENCY}" \
    --max-retries 1 \
    --timeout 90m \
    --watch \
    --non-interactive
set +x

shout "Tests completed"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
