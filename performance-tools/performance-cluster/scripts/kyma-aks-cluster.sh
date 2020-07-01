#!/usr/bin/env bash

# Description: Kyma Upgradeability plan on GKE. The purpose of this script is to install last Kyma release on real GKE cluster, upgrade it with current changes and trigger testing.
#
#
# Expected vars:
#
#  - INPUT_CLUSTER_NAME - name for the new cluster
#  - DOCKER_PUSH_REPOSITORY - Docker repository hostname. Ex. "docker.io/anyrepository"
#  - DOCKER_PUSH_DIRECTORY - Docker "top-level" directory (with leading "/")
#   Ex. "/perf"
#
#  - CLOUDSDK_CORE_PROJECT - GCP project for all GCP resources used during execution (Service Account, IP Address, DNS Zone, image registry etc.)
#  - CLOUDSDK_COMPUTE_REGION - GCP compute region. Ex. "europe-west3"
#  - CLOUDSDK_COMPUTE_ZONE Ex. "europe-west3-a"
#  - GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path.
#    Ex. "/etc/credentials/sa-gke-kyma-integration/service-account.json"
#
#  - DOCKER_IN_DOCKER_ENABLED true
#  - MACHINE_TYPE (optional): GKE machine type
#  - CLUSTER_VERSION (optional): GKE cluster version
#
# Permissions: In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
#  - Compute Admin
#  - Kubernetes Engine Admin
#  - Kubernetes Engine Cluster Admin
#  - DNS Administrator
#  - Service Account User
#  - Storage Admin
#  - Compute Network Admin

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.
discoverUnsetVar=false


for var in INPUT_CLUSTER_NAME DOCKER_PUSH_REPOSITORY DOCKER_PUSH_DIRECTORY RS_GROUP REGION DOCKER_IN_DOCKER_ENABLED ACTION; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

if [ -f "../../prow/scripts/library.sh" ]; then
    export TEST_INFRA_SOURCES_DIR="../.."

elif [ -f "../test-infra/prow/scripts/library.sh" ]; then
    export TEST_INFRA_SOURCES_DIR="../test-infra"

else
	echo "File 'library.sh' can't be found."
    exit 1;
fi

export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

if [[ ! -f "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/aks-library.sh" ]]; then
    echo "File 'aks-library.sh' can't be found."
    exit 1; 
fi

export RS_GROUP="${RS_GROUP}"
export BUILD_TYPE="master"
readonly CURRENT_TIMESTAMP=$(date +%Y%m%d)
readonly STANDARIZED_NAME=$(echo "${INPUT_CLUSTER_NAME}" | tr "[:upper:]" "[:lower:]")
readonly DNS_SUBDOMAIN="${STANDARIZED_NAME}"

export CLUSTER_NAME="${STANDARIZED_NAME}"

export STANDARIZED_NAME
export REPO_OWNER
export REPO_NAME
export CURRENT_TIMESTAMP

# aks-library.sh includes library.sh
# shellcheck disable=SC1090
source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/aks-library.sh"

# Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

shout "Starting workflow"
date
init

shout "Authenticating"
#azureAuthentication

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

function generateAndExportClusterName() {
    readonly REPO_OWNER=$(echo "${REPO_OWNER}" | tr '[:upper:]' '[:lower:]')
    readonly REPO_NAME=$(echo "${REPO_NAME}" | tr '[:upper:]' '[:lower:]')
    readonly RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c10)

    # Otherwise (master), operate on triggering commit id
    readonly COMMON_NAME_PREFIX="aks-perf-test"
    COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")

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

if [[ "${ACTION}" == "delete" ]]; then

    shout "Cleanup"
    date
    cleanup

elif [[ "${ACTION}" == "create" ]]; then
    shout "Create new cluster"
    date

    if [[ "${CLUSTER_GRADE}" == "" ]]; then
      shoutFail "ERROR: ${CLUSTER_GRADE} is not set"
      exit 0
    fi

    if [[ "${CLUSTER_GRADE}" == "production" ]]; then
        export REPO_OWNER="kyma-project"
        export REPO_NAME="kyma"
        shout "Production"
        mkdir -p /${SRC_DIR}/${REPO_OWNER}/${REPO_NAME}
        git clone https://github.com/${REPO_OWNER}/${REPO_NAME}.git ${SRC_DIR}/${REPO_OWNER}/${REPO_NAME}
        export KYMA_SOURCES_DIR="${SRC_DIR}/${REPO_OWNER}/${REPO_NAME}"
    else
        for var in REPO_OWNER REPO_NAME; do
            if [ -z "${!var}" ] ; then
                echo "ERROR: $var is not set"
                discoverUnsetVar=true
            fi
        done
        export KYMA_SOURCES_DIR="${GOPATH}/src/github.com/kyma-project/kyma"
    fi

    #DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --project="${CLOUDSDK_CORE_PROJECT}" --format="value(dnsName)")"
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
else
   shoutFail "None of the actions met"
fi

shout "Success"

# Mark execution as successfully
ERROR_LOGGING_GUARD="false"
