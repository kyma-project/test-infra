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
# - GARDENER_CLUSTER_VERSION - Version of the Kubernetes cluster
# - MACHINE_TYPE (optional): AKS machine type
# - BOT_GITHUB_TOKEN: Bot github token used for API queries
# - RS_GROUP - azure resource group
#
#Permissions: In order to run this script you need to use an AKS service account with the contributor role

set -e

ENABLE_TEST_LOG_COLLECTOR=false

#Exported variables
export RS_GROUP \
    REGION
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export HELM_TIMEOUT_SEC=10000s # timeout in sec for helm install/test operation
export TEST_TIMEOUT_SEC=600   # timeout in sec for test pods until they reach the terminating state
export UPGRADE_TEST_NAMESPACE="e2e-upgrade-test"
export UPGRADE_TEST_RELEASE_NAME="${UPGRADE_TEST_NAMESPACE}"
export UPGRADE_TEST_PATH="${KYMA_SOURCES_DIR}/tests/end-to-end/upgrade/chart/upgrade"
export UPGRADE_TEST_RESOURCE_LABEL="kyma-project.io/upgrade-e2e-test"
export TEST_RESOURCE_LABEL_VALUE_PREPARE="prepareData"
export TEST_CONTAINER_NAME="tests"
export KYMA_UPDATE_TIMEOUT="90m"
export INSTALLATION_OVERRIDE_STACKDRIVER="installer-config-logging-stackdiver.yaml"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/testing-helpers.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/fluent-bit-stackdriver-logging.sh"
# shellcheck source=prow/scripts/lib/gardener/gardener.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gardener.sh"

requiredVars=(
    KYMA_PROJECT_DIR
    GARDENER_REGION
    GARDENER_ZONES
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
    GARDENER_CLUSTER_VERSION
    REGION
    AZURE_SUBSCRIPTION_ID
    AZURE_CREDENTIALS_FILE
    BOT_GITHUB_TOKEN
    RS_GROUP
)

utils::check_required_vars "${requiredVars[@]}"

KYMA_LABEL_PREFIX="kyma-project.io"
KYMA_TEST_LABEL_PREFIX="${KYMA_LABEL_PREFIX}/test"
BEFORE_UPGRADE_LABEL_QUERY="${KYMA_TEST_LABEL_PREFIX}.before-upgrade=true"
POST_UPGRADE_LABEL_QUERY="${KYMA_TEST_LABEL_PREFIX}.after-upgrade=true"

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)

#!Put cleanup code in this function! Function is executed at exit from the script and on interruption.
cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?
    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    if [ "${DEBUG_COMMANDO_OOM}" = "true" ]; then
      # copy output from debug container to artifacts directory
      utils::oom_get_output
    fi

    # collect logs from failed tests before deprovisioning
    kyma::run_test_log_collector "kyma-upgrade-gardener-azure"

    if [[ -n "${SUITE_NAME}" ]]; then
        kyma::test_summary
    fi 

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        log::error "AN ERROR OCCURED! Take a look at preceding log entries."
    fi

    if [ -n "${CLEANUP_CLUSTER}" ]; then
        log::info "Deprovision cluster: \"${CLUSTER_NAME}\""

        gardener::deprovision_cluster "${GARDENER_KYMA_PROW_PROJECT_NAME}" "${CLUSTER_NAME}" "${GARDENER_KYMA_PROW_KUBECONFIG}"
    fi

    rm -rf "${TMP_DIR}"

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    log::info "Job is finished ${MSG}"
    set -e

    exit "${EXIT_STATUS}"
}

function provisionCluster() {
    log::info "Provision cluster: \"${CLUSTER_NAME}\""

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
            --kube-version="${GARDENER_CLUSTER_VERSION}"
    )

    # Schedule pod with oom finder.
    if [ "${DEBUG_COMMANDO_OOM}" = "true" ]; then
      # run oom debug pod
      utils::debug_oom
    fi
}

function getLastReleaseCandidateVersion() {
  version=$(curl --silent --fail --show-error -H "Authorization: token ${BOT_GITHUB_TOKEN}" "https://api.github.com/repos/kyma-project/kyma/releases" |
    jq -r 'del( .[] | select( (.prerelease == false) or (.draft == true) )) | .[0].tag_name ')
  
  echo "${version}"
}

function installKyma() {
    kyma::get_last_release_version \
        -t "${BOT_GITHUB_TOKEN}" \
        -v "1\."
    LAST_RELEASE_VERSION="${kyma_get_last_release_version_return_version:?}"
    mkdir -p /tmp/kyma-gardener-upgradeability
    if [ -z "$LAST_RELEASE_VERSION" ]; then
        log::error "Couldn't grab latest version from GitHub API, stopping."
        exit 1
    fi

    log::info "Installing Kyma ${LAST_RELEASE_VERSION}"

    prepare_stackdriver_logging "${INSTALLATION_OVERRIDE_STACKDRIVER}"

    (
    set -x
    kyma install \
        --ci \
        --source "${LAST_RELEASE_VERSION}" \
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

  # get server IP for HTTPS protocol and split away wildcard ".*" from it using sed
  local domain
  domain=$(kubectl get gateways.networking.istio.io --namespace kyma-system kyma-gateway -o jsonpath='{.spec.servers[?(@.port.protocol=="HTTPS")].hosts[0]}' | sed s/\*\.//g)

  log::info "Create ${name} resources"

  helm install "${name}" \
      --namespace "${namespace}" \
      --create-namespace \
      "${path}" \
      --timeout "${HELM_TIMEOUT_SEC}" \
      --set domain="${domain}" \
      --wait

  prepareResult=$?
  if [[ "${prepareResult}" != 0 ]]; then
      echo "Helm install ${name} operation failed: ${prepareResult}"
      exit "${prepareResult}"
  fi
}

createTestResources() {
    # install upgrade test
    installTestChartOrFail "${UPGRADE_TEST_PATH}" "${UPGRADE_TEST_RELEASE_NAME}" "${UPGRADE_TEST_NAMESPACE}"
}

function upgradeKyma() {
    TARGET_VERSION=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)

    log::info "Upgrading Kyma ${TARGET_VERSION:0:8}"

    log::info "Triggering update with timeout ${KYMA_UPDATE_TIMEOUT}"
    (
        set -x
        kyma upgrade \
            --ci \
            --source "${TARGET_VERSION:0:8}" \
            --timeout ${KYMA_UPDATE_TIMEOUT}
    )
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

# testKyma executes the kyma-testing.sh. labelqueries can be passed as arguments and will be passed to the kyma cli
function testKyma() {
  testing::inject_addons_if_necessary

  local labelquery=${1}
  local suitename=${2}
  local test_args=()

  if [[ -n ${labelquery} ]]; then
    test_args+=("-l")
    test_args+=("${labelquery}")
  fi

  if [[ -n ${suitename} ]]; then
    test_args+=("-n")
    test_args+=("${suitename}")
  fi

  test_args+=("-t")
  test_args+=("2h")

  # remove cluster-users test as it takes more than 1h to run and is not an essential test
  kubectl delete -n kyma-system testdefinition/cluster-users --ignore-not-found

  log::banner "Test Kyma " "${test_args[@]}"
  "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh "${test_args[@]}"

  testing::remove_addons_if_necessary
}

trap cleanup EXIT INT

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c4)
readonly COMMON_NAME_PREFIX="grdnr"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")

### Cluster name must be less than 10 characters!
export CLUSTER_NAME="${COMMON_NAME}"

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

kyma::install_cli

provisionCluster

installKyma

createTestResources

testKyma "${BEFORE_UPGRADE_LABEL_QUERY}" testsuite-all-before-upgrade

upgradeKyma

# enable test-log-collector before tests; if prowjob fails before test phase we do not have any reason to enable it earlier
if [[ "${BUILD_TYPE}" == "master" && -n "${LOG_COLLECTOR_SLACK_TOKEN}" ]]; then
  export ENABLE_TEST_LOG_COLLECTOR=true
fi

testKyma "${POST_UPGRADE_LABEL_QUERY}" testsuite-all-after-upgrade

log::success "Job finished with success"

#!!! Must be at the end of the script because cleanup evaluates it !!!
ERROR_LOGGING_GUARD="false"
