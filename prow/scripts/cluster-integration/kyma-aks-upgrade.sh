#!/usr/bin/env bash

set -o errexit   # Exit immediately if a command exits with a non-zero status.
set -o pipefail  # Fail a pipe if any sub-command fails.

VARIABLES=(
    RS_GROUP
    REGION
    AZURE_SUBSCRIPTION_ID
    AZURE_SUBSCRIPTION_APP_ID
    AZURE_SUBSCRIPTION_SECRET
    AZURE_SUBSCRIPTION_TENANT
    KYMA_PROJECT_DIR
    INPUT_CLUSTER_NAME
    GOOGLE_APPLICATION_CREDENTIALS
    CLOUDSDK_DNS_ZONE_NAME
    CLOUDSDK_CORE_PROJECT
    KYMA_ALERTS_CHANNEL
    KYMA_ALERTS_SLACK_API_URL
    SLACK_CLIENT_WEBHOOK_URL
    STABILITY_SLACK_CLIENT_CHANNEL_ID
    SLACK_CLIENT_TOKEN
    TEST_RESULT_WINDOW_TIME
    DOCKER_PUSH_REPOSITORY
    DOCKER_PUSH_DIRECTORY
)

discoverUnsetVar=false

for var in "${VARIABLES[@]}"; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
export KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"
export KYMA_INSTALL_TIMEOUT="30m"
export KYMA_UPDATE_TIMEOUT="25m"
export UPGRADE_TEST_PATH="${KYMA_SOURCES_DIR}/tests/end-to-end/upgrade/chart/upgrade"
# timeout in sec for helm operation install/test
export UPGRADE_TEST_HELM_TIMEOUT_SEC=10000
# timeout in sec for e2e upgrade test pods until they reach the terminating state
export UPGRADE_TEST_TIMEOUT_SEC=600
export UPGRADE_TEST_NAMESPACE="e2e-upgrade-test"
export UPGRADE_TEST_RELEASE_NAME="${UPGRADE_TEST_NAMESPACE}"
export UPGRADE_TEST_RESOURCE_LABEL="kyma-project.io/upgrade-e2e-test"
export UPGRADE_TEST_LABEL_VALUE_PREPARE="prepareData"
export UPGRADE_TEST_LABEL_VALUE_EXECUTE="executeTests"
export TEST_CONTAINER_NAME="runner"

readonly STANDARIZED_NAME=$(echo "${INPUT_CLUSTER_NAME}" | tr "[:upper:]" "[:lower:]")
readonly DNS_SUBDOMAIN="${STANDARIZED_NAME}"
readonly REPO_OWNER="kyma-project"
readonly REPO_NAME="kyma"
readonly CURRENT_TIMESTAMP=$(date +%Y%m%d)

export CLUSTER_NAME="${STANDARIZED_NAME}"
export CLUSTER_SIZE="Standard_D4_v3"
# set cluster version as MAJOR.MINOR without PATCH part (e.g. 1.10, 1.11)
export CLUSTER_K8S_VERSION="1.11"
export CLUSTER_ADDONS="monitoring,http_application_routing"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
# shellcheck disable=SC1090
source "${KYMA_SCRIPTS_DIR}/testing-common.sh"

# shellcheck disable=SC1090
source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/aks-library.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    shout "Execute Job Guard"
    "${TEST_INFRA_SOURCES_DIR}/development/tools/cmd/jobguard/run.sh"
fi

function cleanup() {
    shout "Cleanup"
    date

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e
    EXIT_STATUS=$?

    if [[ "${ERROR_LOGGING_GUARD}" = "true" ]]; then
        shout "AN ERROR OCCURED! Take a look at preceding log entries."
        echo
    fi

    CHECK_GROUP=$(az group list --query '[?name==`'"${RS_GROUP}"'`].name' -otsv)
    if [ "${CHECK_GROUP}" = "${RS_GROUP}" ]; then
        CLUSTER_RS_GROUP=$(az aks show -g "${RS_GROUP}" -n "${CLUSTER_NAME}" --query nodeResourceGroup -o tsv)
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

        echo "Remove DNS Record for Ingressgateway"
        GATEWAY_DNS_FULL_NAME="*.${DOMAIN}."
        GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"

        GATEWAY_IP_ADDRESS=$(az network public-ip show -g "${CLUSTER_RS_GROUP}" -n "${GATEWAY_IP_ADDRESS_NAME}" --query ipAddress -o tsv)
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

        IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

        echo "Remove DNS Record for Apiserver Proxy IP"
        APISERVER_DNS_FULL_NAME="apiserver.${DOMAIN}."
        APISERVER_IP_ADDRESS=$(gcloud dns record-sets list --zone "${CLOUDSDK_DNS_ZONE_NAME}" --name "${APISERVER_DNS_FULL_NAME}" --format="value(rrdatas[0])")
        if [[ -n ${APISERVER_IP_ADDRESS} ]]; then
            IP_ADDRESS=${APISERVER_IP_ADDRESS} DNS_FULL_NAME=${APISERVER_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/delete-dns-record.sh"
            TMP_STATUS=$?
            if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
        fi

        echo "Remove Cluster, IP Address for Ingressgateway"
        az aks delete -g "${RS_GROUP}" -n "${CLUSTER_NAME}" -y
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

        echo "Remove group"
        az group delete -n "${RS_GROUP}" -y
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    else
        echo "Azure group does not exist, skip cleanup process"
    fi

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    echo "Cleanup function is finished ${MSG}"

    # Turn on exit-on-error
    set -e
}

trap cleanup EXIT INT

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    shout "Execute Job Guard"
    "${TEST_INFRA_SOURCES_DIR}/development/tools/cmd/jobguard/run.sh"
fi

function generateAndExportClusterName() {
    readonly REPO_OWNER=$(echo "${REPO_OWNER}" | tr '[:upper:]' '[:lower:]')
    readonly REPO_NAME=$(echo "${REPO_NAME}" | tr '[:upper:]' '[:lower:]')
    readonly RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c10)

    if [[ "$BUILD_TYPE" == "pr" ]]; then
        readonly COMMON_NAME_PREFIX="gke-upgrade-pr"
        # In case of PR, operate on PR number
        COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
    elif [[ "$BUILD_TYPE" == "release" ]]; then
        readonly COMMON_NAME_PREFIX="gke-upgrade-rel"
        readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
        readonly RELEASE_VERSION=$(cat "${SCRIPT_DIR}/../../RELEASE_VERSION")
        shout "Reading release version from RELEASE_VERSION file, got: ${RELEASE_VERSION}"
        COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
    else
        # Otherwise (master), operate on triggering commit id
        readonly COMMON_NAME_PREFIX="gke-upgrade-commit"
        COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)
        COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
    fi

    ### Cluster name must be less than 40 characters!
    export CLUSTER_NAME="${COMMON_NAME}"

    export AZURE_NETWORK_NAME="${COMMON_NAME_PREFIX}-net"
    export AZURE_SUBNET_NAME="${COMMON_NAME_PREFIX}-subnet"
}

function generateAndExportCerts() {
    shout "Generate self-signed certificate"
    date
    CERT_KEY=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-self-signed-cert.sh")

    TLS_CERT=$(echo "${CERT_KEY}" | head -1)
    export TLS_CERT
    TLS_KEY=$(echo "${CERT_KEY}" | tail -1)
    export TLS_KEY
}

function getLastReleaseVersion() {
    version=$(curl --silent --fail --show-error "https://api.github.com/repos/kyma-project/kyma/releases?access_token=${BOT_GITHUB_TOKEN}" \
     | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-1].tag_name')

    echo "${version}"
}

function installKyma() {
    mkdir -p /tmp/kyma-gke-upgradeability
    LAST_RELEASE_VERSION=$(getLastReleaseVersion)

    shout "Install Tiller from version ${LAST_RELEASE_VERSION}"
    date
    kubectl apply -f "https://raw.githubusercontent.com/kyma-project/kyma/${LAST_RELEASE_VERSION}/installation/resources/tiller.yaml"
    "${KYMA_SCRIPTS_DIR}"/is-ready.sh kube-system name tiller

    shout "Apply Kyma config from version ${LAST_RELEASE_VERSION}"
    date
    kubectl create namespace "kyma-installer"

     "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "knative-serving-overrides" \
        --data "knative-serving.domainName=${DOMAIN}" \
        --label "component=knative-serving" #Backward compatibility for releases <= 1.1.X

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "installation-config-overrides" \
        --data "global.domainName=${DOMAIN}" \
        --data "global.loadBalancerIP=${GATEWAY_IP_ADDRESS}" \
        --data "cluster-users.users.adminGroup=" #Backward compatibility for releases <= 1.1.X

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "core-test-ui-acceptance-overrides" \
        --data "test.acceptance.ui.logging.enabled=true" \
        --label "component=core"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "intallation-logging-overrides" \
        --data "global.logging.promtail.config.name=${PROMTAIL_CONFIG_NAME}" \
        --label "component=logging" #Backward compatibility for releases <= 1.1.X

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "cluster-certificate-overrides" \
        --data "global.tlsCrt=${TLS_CERT}" \
        --data "global.tlsKey=${TLS_KEY}"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "istio-overrides" \
        --data "gateways.istio-ingressgateway.loadBalancerIP=${GATEWAY_IP_ADDRESS}" \
        --label "component=istio"

    waitUntilInstallerApiAvailable

    shout "Use released artifacts from version ${LAST_RELEASE_VERSION}"
    date

    curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${LAST_RELEASE_VERSION}/kyma-installer-cluster.yaml" --output /tmp/kyma-gke-upgradeability/last-release-installer.yaml
    kubectl apply -f /tmp/kyma-gke-upgradeability/last-release-installer.yaml

    kubectl label installation/kyma-installation action=install --overwrite #Backward compatibility for releases <= 1.1.X

    shout "Installation triggered with timeout ${KYMA_INSTALL_TIMEOUT}"
    date
    "${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout ${KYMA_INSTALL_TIMEOUT}

    # re-check if this is needed here
    # if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
    #     shout "Create DNS Record for Apiserver proxy IP"
    #     date
    #     APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    #     APISERVER_DNS_FULL_NAME="apiserver.${DOMAIN}."
    #     IP_ADDRESS=${APISERVER_IP_ADDRESS} DNS_FULL_NAME=${APISERVER_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
    # fi
}

function checkTestPodTerminated() {
    local retry=0
    local runningPods=0
    local succeededPods=0
    local failedPods=0

    while [ "${retry}" -lt "${UPGRADE_TEST_TIMEOUT_SEC}" ]; do
        # check status phase for each testing pods
        for podName in $(kubectl get pods -n "${UPGRADE_TEST_NAMESPACE}" -o json | jq -sr '.[]|.items[].metadata.name')
        do
            runningPods=$((runningPods + 1))
            phase=$(kubectl get pod "${podName}" -n "${UPGRADE_TEST_NAMESPACE}" -o json | jq '.status.phase')
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
        if [ "${failedPods}" -gt 0 ]
        then
            echo "${failedPods} pod(s) has failed status"
            return 1
        fi

        # exit from function if each pod has succeeded status
        if [ "${runningPods}" == "${succeededPods}" ]
        then
            echo "All pods in ${UPGRADE_TEST_NAMESPACE} namespace have succeeded phase"
            return 0
        fi

        # reset all counters and rerun checking
        delta=$((runningPods-succeededPods))
        echo "${delta} pod(s) in ${UPGRADE_TEST_NAMESPACE} namespace have not terminated phase. Retry checking."
        runningPods=0
        succeededPods=0
        retry=$((retry + 1))
        sleep 5
    done

    echo "The maximum number of attempts: ${retry} has been reached"
    return 1
}

function createTestResources() {
    shout "Create e2e upgrade test resources"
    date

    injectTestingBundles

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
    kubectl logs -n "${UPGRADE_TEST_NAMESPACE}" -l "${UPGRADE_TEST_RESOURCE_LABEL}=${UPGRADE_TEST_LABEL_VALUE_PREPARE}" -c "${TEST_CONTAINER_NAME}"
    if [ "${prepareTestResult}" != 0 ]; then
        echo "Exit status for prepare upgrade e2e tests: ${prepareTestResult}"
        exit "${prepareTestResult}"
    fi
}

function upgradeKyma() {
    shout "Delete the kyma-installation CR"
    # Remove the finalizer form kyma-installation the merge type is used because strategic is not supported on CRD.
    # More info about merge strategy can be found here: https://tools.ietf.org/html/rfc7386
    kubectl patch Installation kyma-installation -n default --patch '{"metadata":{"finalizers":null}}' --type=merge
    kubectl delete Installation -n default kyma-installation

    if [[ "$BUILD_TYPE" == "release" ]]; then
        echo "Use released artifacts"
        gsutil cp "${KYMA_ARTIFACTS_BUCKET}/${RELEASE_VERSION}/kyma-installer-cluster.yaml" /tmp/kyma-gke-upgradeability/new-release-kyma-installer.yaml
        gsutil cp "${KYMA_ARTIFACTS_BUCKET}/${RELEASE_VERSION}/tiller.yaml" /tmp/kyma-gke-upgradeability/new-tiller.yaml

        shout "Update tiller"
        kubectl apply -f /tmp/kyma-gke-upgradeability/new-tiller.yaml

        shout "Update kyma installer"
        kubectl apply -f /tmp/kyma-gke-upgradeability/new-release-kyma-installer.yaml
    else
        shout "Build Kyma Installer Docker image"
        date
        COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)
        KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-upgradeability/${REPO_OWNER}/${REPO_NAME}:COMMIT-${COMMIT_ID}"
        export KYMA_INSTALLER_IMAGE
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-image.sh"

        KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"
        INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
        INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"

        shout "Update tiller"
        kubectl apply -f "${KYMA_RESOURCES_DIR}/tiller.yaml"

        shout "Manual concatenating and applying installer.yaml and installer-cr-cluster.yaml YAMLs"
        "${KYMA_SCRIPTS_DIR}"/concat-yamls.sh "${INSTALLER_YAML}" "${INSTALLER_CR}" \
        | sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" \
        | sed -e "s/__VERSION__/0.0.1/g" \
        | sed -e "s/__.*__//g" \
        | kubectl apply -f-
    fi

    shout "Update triggered with timeout ${KYMA_UPDATE_TIMEOUT}"
    date
    "${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout ${KYMA_UPDATE_TIMEOUT}

    if [ -n "$(kubectl get  service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
        shout "Create DNS Record for Apiserver proxy IP"
        date
        APISERVER_IP_ADDRESS=$(kubectl get  service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
        IP_ADDRESS=${APISERVER_IP_ADDRESS} DNS_FULL_NAME=${APISERVER_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
    fi
}

function testKyma() {
    shout "Test Kyma end-to-end upgrade scenarios"
    date

    if [  -f "$(helm home)/ca.pem" ]; then
        local HELM_ARGS="--tls"
    fi

    set +o errexit
    helm test "${UPGRADE_TEST_RELEASE_NAME}" --timeout "${UPGRADE_TEST_HELM_TIMEOUT_SEC}" ${HELM_ARGS}
    testEndToEndResult=$?

    echo "Test e2e upgrade logs: "
    kubectl logs -n "${UPGRADE_TEST_NAMESPACE}" -l "${UPGRADE_TEST_RESOURCE_LABEL}=${UPGRADE_TEST_LABEL_VALUE_EXECUTE}" -c "${TEST_CONTAINER_NAME}"

    if [ "${testEndToEndResult}" != 0 ]; then
        echo "Helm test operation failed: ${testEndToEndResult}"
        exit "${testEndToEndResult}"
    fi
    set -o errexit

    removeTestingBundles

    shout "Test Kyma"
    date
    "${KYMA_SCRIPTS_DIR}"/testing.sh
}

# Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

init
azureAuthenticating

DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --project="${CLOUDSDK_CORE_PROJECT}" --format="value(dnsName)")"
export DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"

generateAndExportClusterName

addGithubDexConnector # github needed?

createGroup
installCluster

createPublicIPandDNS

generateAndExportCerts

setupKubeConfig
installKyma
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-helm-certs.sh"

createTestResources

upgradeKyma

testKyma

shout "Job finished with success"

# Mark execution as successfully
ERROR_LOGGING_GUARD="false"
