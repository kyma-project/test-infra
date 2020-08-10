#!/usr/bin/env bash

#Description: Kyma upgrade plan on Gardener with Azure. This scripts implements a pipeline that consists of many steps. The purpose is to install, upgrade, and test Kyma using the CLI on a real Gardener Azure cluster.
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
# - BOT_GITHUB_TOKEN: Bot github token used for API queries
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
    BOT_GITHUB_TOKEN
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
export HELM_TIMEOUT_SEC=10000s # timeout in sec for helm install/test operation
export TEST_TIMEOUT_SEC=600   # timeout in sec for test pods until they reach the terminating state
export UPGRADE_TEST_NAMESPACE="e2e-upgrade-test"
export UPGRADE_TEST_RELEASE_NAME="${UPGRADE_TEST_NAMESPACE}"
export UPGRADE_TEST_PATH="${KYMA_SOURCES_DIR}/tests/end-to-end/upgrade/chart/upgrade"
export UPGRADE_TEST_RESOURCE_LABEL="kyma-project.io/upgrade-e2e-test"
export TEST_RESOURCE_LABEL_VALUE_PREPARE="prepareData"
export EXTERNAL_SOLUTION_TEST_PATH="${KYMA_SOURCES_DIR}/tests/end-to-end/external-solution-integration/chart/external-solution"
export EXTERNAL_SOLUTION_TEST_NAMESPACE="integration-test"
export EXTERNAL_SOLUTION_TEST_RELEASE_NAME="${EXTERNAL_SOLUTION_TEST_NAMESPACE}"
export EXTERNAL_SOLUTION_TEST_RESOURCE_LABEL="kyma-project.io/external-solution-e2e-test"
export TEST_CONTAINER_NAME="tests"
export KYMA_UPDATE_TIMEOUT="40m"
export INSTALLATION_OVERRIDE_STACKDRIVER="installer-config-logging-stackdiver.yaml"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/testing-helpers.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/kyma-cli.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/fluent-bit-stackdriver-logging.sh"

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

function provisionCluster() {
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
}

function getLastReleaseVersion() {
    version=$(curl --silent --fail --show-error "https://api.github.com/repos/kyma-project/kyma/releases?access_token=${BOT_GITHUB_TOKEN}" \
     | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-1].tag_name')

    echo "${version}"
}

function installKyma() {
    LAST_RELEASE_VERSION=$(getLastReleaseVersion)
    mkdir -p /tmp/kyma-gardener-upgradeability
    if [ -z "$LAST_RELEASE_VERSION" ]; then
        shout "Couldn't grab latest version from GitHub API, stopping."
        exit 1
    fi

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

    prepare_stackdriver_logging "${INSTALLATION_OVERRIDE_STACKDRIVER}"
    if [[ "$?" -ne 0 ]]; then
        return 1
    fi

    (
    set -x
    kyma install \
        --ci \
        --source "${LAST_RELEASE_VERSION}" \
        -o installer-config-production.yaml.tpl \
        -o installer-config-azure-eventhubs.yaml.tpl \
        -o "${INSTALLATION_OVERRIDE_STACKDRIVER}" \
        --timeout 90m
    )
}

function checkTestPodTerminated() {
    local namespace=$1
    local retry=0
    local runningPods=0
    local succeededPods=0
    local failedPods=0

    while [[ "${retry}" -lt "${TEST_TIMEOUT_SEC}" ]]; do
        # check status phase for each testing pods
        for podName in $(kubectl get pods -n "${namespace}" -o json | jq -sr '.[]|.items[].metadata.name')
        do
            runningPods=$((runningPods + 1))
            phase=$(kubectl get pod "${podName}" -n "${namespace}" -o json | jq '.status.phase')
            echo "Test pod '${podName}' has phase: ${phase}"

            if [[ "${phase}" == *"Succeeded"* ]]
            then
                succeededPods=$((succeededPods + 1))
            fi

            if [[ "${phase}" == *"Failed"* ]]; then
                failedPods=$((failedPods + 1))
            fi
        done

        # exit permanently if one of cluster has failed status
        if [[ "${failedPods}" -gt 0 ]]
        then
            echo "${failedPods} pod(s) has failed status"
            return 1
        fi

        # exit from function if each pod has succeeded status
        if [[ "${runningPods}" == "${succeededPods}" ]]
        then
            echo "All pods in ${namespace} namespace have succeeded phase"
            return 0
        fi

        # reset all counters and rerun checking
        delta=$((runningPods-succeededPods))
        echo "${delta} pod(s) in ${namespace} namespace have not terminated phase. Retry checking."
        runningPods=0
        succeededPods=0
        retry=$((retry + 1))
        sleep 5
    done

    echo "The maximum number of attempts: ${retry} has been reached"
    return 1
}

function installTestChartOrFail() {
  local path=$1
  local name=$2
  local namespace=$3

  shout "Create ${name} resources"
  date

  helm install "${name}" \
      --namespace "${namespace}" \
      --create-namespace \
      "${path}" \
      --timeout "${HELM_TIMEOUT_SEC}" \
      --set domain="${DOMAIN}" \
      --wait

  prepareResult=$?
  if [[ "${prepareResult}" != 0 ]]; then
      echo "Helm install ${name} operation failed: ${prepareResult}"
      exit "${prepareResult}"
  fi
}

function waitForTestPodToFinish() {
  local name=$1
  local namespace=$2
  local label=$3

  set +o errexit
  checkTestPodTerminated "${namespace}"
  prepareTestResult=$?
  set -o errexit

  echo "Logs for prepare data operation to ${name}: "
  # shellcheck disable=SC2046
  kubectl logs -n "${namespace}" $(kubectl get pod -n "${name}" -l "${label}=${TEST_RESOURCE_LABEL_VALUE_PREPARE}" -o json | jq -r '.items | .[] | .metadata.name') -c "${TEST_CONTAINER_NAME}"
  if [[ "${prepareTestResult}" != 0 ]]; then
      echo "Exit status for prepare ${name}: ${prepareTestResult}"
      exit "${prepareTestResult}"
  fi
}

createTestResources() {
    injectTestingAddons

    # install upgrade test
    installTestChartOrFail "${UPGRADE_TEST_PATH}" "${UPGRADE_TEST_RELEASE_NAME}" "${UPGRADE_TEST_NAMESPACE}"

    # install external-solution test
    installTestChartOrFail "${EXTERNAL_SOLUTION_TEST_PATH}" "${EXTERNAL_SOLUTION_TEST_RELEASE_NAME}" "${EXTERNAL_SOLUTION_TEST_NAMESPACE}"

    # wait for upgrade test to finish
    waitForTestPodToFinish "${UPGRADE_TEST_RELEASE_NAME}" "${UPGRADE_TEST_NAMESPACE}" "${UPGRADE_TEST_RESOURCE_LABEL}"

    # wait for external-solution test to finish
    waitForTestPodToFinish "${EXTERNAL_SOLUTION_TEST_RELEASE_NAME}" "${EXTERNAL_SOLUTION_TEST_NAMESPACE}" "${EXTERNAL_SOLUTION_TEST_RESOURCE_LABEL}"
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
        --output /tmp/kyma-gardener-upgradeability/upgraded-tiller.yaml

    curl -L --silent --fail --show-error "https://storage.googleapis.com/kyma-development-artifacts/master-${TARGET_VERSION:0:8}/kyma-installer-cluster.yaml" \
        --output /tmp/kyma-gardener-upgradeability/upgraded-release-installer.yaml

    shout "Install Tiller from version ${TARGET_VERSION}"
    date
    kubectl apply -f /tmp/kyma-gardener-upgradeability/upgraded-tiller.yaml

    shout "Wait until tiller is correctly rolled out"
    kubectl -n kube-system rollout status deployment/tiller-deploy

    shout "Use release artifacts from version ${TARGET_VERSION}"
    date
    kubectl apply -f /tmp/kyma-gardener-upgradeability/upgraded-release-installer.yaml

    shout "Update triggered with timeout ${KYMA_UPDATE_TIMEOUT}"
    date
    "${KYMA_SOURCES_DIR}"/installation/scripts/is-installed.sh --timeout ${KYMA_UPDATE_TIMEOUT}
}

remove_addons_if_necessary() {
  tdWithAddon=$(kubectl get td --all-namespaces -l testing.kyma-project.io/require-testing-addon=true -o custom-columns=NAME:.metadata.name --no-headers=true)

  if [ -z "$tdWithAddon" ]
  then
      echo "- Removing ClusterAddonsConfiguration which provides the testing addons"
      removeTestingAddons
      if [[ $? -eq 1 ]]; then
        exit 1
      fi
  else
      echo "- Skipping removing ClusterAddonsConfiguration"
  fi
}

function testKyma() {
    shout "Test Kyma"
    date
    "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh
}

trap cleanup EXIT INT

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c4)
readonly COMMON_NAME_PREFIX="grdnr"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")

### Cluster name must be less than 10 characters!
export CLUSTER_NAME="${COMMON_NAME}"

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

export INSTALL_DIR=${TMP_DIR}
install::kyma_cli

provisionCluster

installKyma
if [[ "$?" -ne 0 ]]; then
    return 1
fi

createTestResources

upgradeKyma

remove_addons_if_necessary

testKyma

shout "Job finished with success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
