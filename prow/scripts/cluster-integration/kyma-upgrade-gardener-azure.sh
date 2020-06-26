#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on a real Gardener cluster.
#
#
#Expected vars:
#
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - GARDENER_REGION - Gardener compute region
# - GARDENER_ZONES - Gardener compute zones inside the region
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME Name of the azure secret configured in the gardener project to access the cloud provider
# - MACHINE_TYPE (optional): AKS machine type
#
#Permissions: In order to run this script you need to use an AKS service account with the contributor role

set -e

discoverUnsetVar=false

VARIABLES=(
    KYMA_PROJECT_DIR
    GARDENER_REGION
    GARDENER_ZONES
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
    RS_GROUP
    REGION
    EVENTHUB_NAMESPACE_NAME
    AZURE_SUBSCRIPTION_ID
    AZURE_SUBSCRIPTION_APP_ID
    AZURE_SUBSCRIPTION_SECRET
    AZURE_SUBSCRIPTION_TENANT
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
    EVENTHUB_NAMESPACE_NAME \
    REGION \
    AZURE_SUBSCRIPTION_ID \
    AZURE_SUBSCRIPTION_APP_ID \
    AZURE_SUBSCRIPTION_SECRET \
    AZURE_SUBSCRIPTION_TENANT
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export EVENTHUB_SECRET_OVERRIDE_FILE="eventhubs-secret-overrides.yaml"
export UPGRADE_TEST_HELM_TIMEOUT_SEC=10000
export UPGRADE_TEST_NAMESPACE="e2e-upgrade-test"
export UPGRADE_TEST_RELEASE_NAME="${UPGRADE_TEST_NAMESPACE}"
export UPGRADE_TEST_PATH="${KYMA_SOURCES_DIR}/tests/end-to-end/upgrade/chart/upgrade"
export UPGRADE_TEST_RESOURCE_LABEL="kyma-project.io/upgrade-e2e-test"
export UPGRADE_TEST_LABEL_VALUE_PREPARE="prepareData"
export UPGRADE_TEST_LABEL_VALUE_EXECUTE="executeTests"
export TEST_CONTAINER_NAME="tests"
export KYMA_UPDATE_TIMEOUT="30m"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/testing-helpers.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/kyma-cli.sh"

#!Put cleanup code in this function! Function is executed at exit from the script and on interuption.
cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?
    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    if [[ -n "${SUITE_NAME}" ]]; then
        testSummary
    fi 

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        shout "AN ERROR OCCURED! Take a look at preceding log entries."
        echo
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

    rm -rf "${TMP_DIR}"

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    shout "Job is finished ${MSG}"
    date
    set -e

    exit "${EXIT_STATUS}"
}

createTestResources() {
    shout "Create e2e upgrade test resources"
    date

    injectTestingAddons

    if [  -f "$(helm home)/ca.pem" ]; then
        local HELM_ARGS="--tls"
    fi

    helm install "${UPGRADE_TEST_PATH}" \
        --name "${UPGRADE_TEST_RELEASE_NAME}" \
        --namespace "${UPGRADE_TEST_NAMESPACE}" \
        --timeout "${UPGRADE_TEST_HELM_TIMEOUT_SEC}" \
        --wait ${HELM_ARGS}

    prepareResult=$?
    if [ "${prepareResult}" != 0 ]; then
        echo "Helm install operation failed: ${prepareResult}"
        exit "${prepareResult}"
    fi

    set +o errexit
    checkTestPodTerminated
    prepareTestResult=$?
    set -o errexit

    echo "Logs for prepare data operation to test e2e upgrade: "
    # shellcheck disable=SC2046
    kubectl logs -n "${UPGRADE_TEST_NAMESPACE}" $(kubectl get pod -n "${UPGRADE_TEST_NAMESPACE}" -l "${UPGRADE_TEST_RESOURCE_LABEL}=${UPGRADE_TEST_LABEL_VALUE_PREPARE}" -o json | jq -r '.items | .[] | .metadata.name') -c "${TEST_CONTAINER_NAME}"
    if [ "${prepareTestResult}" != 0 ]; then
        echo "Exit status for prepare upgrade e2e tests: ${prepareTestResult}"
        exit "${prepareTestResult}"
    fi
}

function upgradeKyma() {
    shout "Delete the kyma-installation CR and kyma-installer deployment"
    # Remove the finalizer form kyma-installation the merge type is used because strategic is not supported on CRD.
    # More info about merge strategy can be found here: https://tools.ietf.org/html/rfc7386
    kubectl patch Installation kyma-installation -n default --patch '{"metadata":{"finalizers":null}}' --type=merge
    kubectl delete Installation -n default kyma-installation

    # Remove the current installer to prevent it performing any action.
    kubectl delete deployment -n kyma-installer kyma-installer

    TARGET_VERSION=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)

    curl -L --silent --fail --show-error "https://raw.githubusercontent.com/kyma-project/kyma/${TARGET_VERSION}/installation/resources/tiller.yaml" \
        --output /tmp/kyma-gke-upgradeability/upgraded-tiller.yaml

    curl -L --silent --fail --show-error "https://storage.googleapis.com/kyma-development-artifacts/master-${TARGET_VERSION:0:8}/kyma-installer-cluster.yaml" \
        --output /tmp/kyma-gke-upgradeability/upgraded-release-installer.yaml

    shout "Install Tiller from version ${TARGET_VERSION}"
    date
    kubectl apply -f /tmp/kyma-gke-upgradeability/upgraded-tiller.yaml
    
    shout "Wait until tiller is correctly rolled out"
    kubectl -n kube-system rollout status deployment/tiller-deploy
    
    shout "Use release artifacts from version ${TARGET_VERSION}"
    date
    kubectl apply -f /tmp/kyma-gke-upgradeability/upgraded-release-installer.yaml

    shout "Update triggered with timeout ${KYMA_UPDATE_TIMEOUT}"
    date
    "${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout ${KYMA_UPDATE_TIMEOUT}
}

function testKyma() {
    shout "Test Kyma"
    date
    "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh
}

testSummary() {
    echo "Test Summary"
    kyma test status "${SUITE_NAME}" -owide

    statusSucceeded=$(kubectl get cts "${SUITE_NAME}" -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
    if [[ "${statusSucceeded}" != *"True"* ]]; then
        echo "- Fetching logs due to test suite failure"

        echo "- Fetching logs from testing pods in Failed status..."
        kyma test logs "${SUITE_NAME}" --test-status Failed

        echo "- Fetching logs from testing pods in Unknown status..."
        kyma test logs "${SUITE_NAME}" --test-status Unknown

        echo "- Fetching logs from testing pods in Running status due to running afer test suite timeout..."
        kyma test logs "${SUITE_NAME}" --test-status Running

        echo "ClusterTestSuite details"
        kubectl get cts "${SUITE_NAME}" -oyaml

        exit 1
    fi

    echo "ClusterTestSuite details"
    kubectl get cts "${SUITE_NAME}" -oyaml
}

trap cleanup EXIT INT

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c4)
readonly COMMON_NAME_PREFIX="grdnr"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")

### Cluster name must be less than 10 characters!
export CLUSTER_NAME="${COMMON_NAME}"

# Local variables
# DNS_SUBDOMAIN="${COMMON_NAME}"

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

export INSTALL_DIR=${TMP_DIR}
install::kyma_cli

shout "Updated script"

shout "Provision cluster: \"${CLUSTER_NAME}\""

if [ -z "$MACHINE_TYPE" ]; then
      export MACHINE_TYPE="Standard_D8_v3"
fi

CLEANUP_CLUSTER="true"
(
set -x
kyma provision gardener az \
        --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" --name "${CLUSTER_NAME}" \
        --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
        --region "${GARDENER_REGION}" -z "${GARDENER_ZONES}" -t "${MACHINE_TYPE}" \
        --scaler-max 4 --scaler-min 3 \
        --kube-version=${GARDENER_CLUSTER_VERSION}
)

shout "Installing Kyma"
date

shout "Downloading Kyma installer CR"
curl -L --silent --fail --show-error "https://raw.githubusercontent.com/kyma-project/kyma/master/installation/resources/installer-cr-azure-eventhubs.yaml.tpl" \
    --output installer-cr-gardener-azure.yaml.tpl

echo "Downlading production profile"
curl -L --silent --fail --show-error "https://raw.githubusercontent.com/kyma-project/kyma/master/installation/resources/installer-config-production.yaml.tpl" \
    --output installer-config-production.yaml.tpl

shout "Downloading Azure EventHubs config"
curl -L --silent --fail --show-error "https://raw.githubusercontent.com/kyma-project/kyma/master/installation/resources/installer-config-azure-eventhubs.yaml.tpl" \
    --output installer-config-azure-eventhubs.yaml.tpl

shout "Generate Azure Event Hubs overrides"
date
# shellcheck disable=SC1090
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-azure-event-hubs-secret.sh
cat "${EVENTHUB_SECRET_OVERRIDE_FILE}" >> installer-config-azure-eventhubs.yaml.tpl
readonly RELEASE_VERSION=1.13.0 #!!! remove fix version after test $(cat "${TEST_INFRA_SOURCES_DIR}/prow/RELEASE_VERSION")
(
set -x
kyma install \
    --ci \
    --source "${RELEASE_VERSION}" \
    -o installer-config-production.yaml.tpl \
    -o installer-config-azure-eventhubs.yaml.tpl \
    --timeout 90m
)

shout "Checking the versions"
date
kyma version

"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-helm-certs.sh"

createTestResources

upgradeKyma

testKyma

shout "Job finished with success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"