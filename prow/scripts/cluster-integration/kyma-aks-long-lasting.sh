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

# INIT ENVIRONMENT VARIABLES
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
export KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

readonly STANDARIZED_NAME=$(echo "${INPUT_CLUSTER_NAME}" | tr "[:upper:]" "[:lower:]")
readonly DNS_SUBDOMAIN="${STANDARIZED_NAME}"
readonly REPO_OWNER="kyma-project"
readonly REPO_NAME="kyma"
readonly CURRENT_TIMESTAMP=$(date +%Y%m%d)

export CLUSTER_NAME="${STANDARIZED_NAME}"
export CLUSTER_SIZE="Standard_DS2_v2"
# set cluster version as MAJOR.MINOR without PATCH part (e.g. 1.10, 1.11)
export CLUSTER_K8S_VERSION="1.11"
export CLUSTER_ADDONS="monitoring,http_application_routing"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

function cleanup() {
    shout "Cleanup"
    date

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e
    EXIT_STATUS=$?

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

        echo "Remove DNS Record for Remote Environments IP"
        REMOTEENVS_DNS_FULL_NAME="gateway.${DOMAIN}."
        REMOTEENVS_IP_NAME="remoteenvs-${STANDARIZED_NAME}"

        REMOTEENVS_IP_ADDRESS=$(az network public-ip show -g "${CLUSTER_RS_GROUP}" -n "${REMOTEENVS_IP_NAME}" --query ipAddress -o tsv)
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

        IP_ADDRESS=${REMOTEENVS_IP_ADDRESS} DNS_FULL_NAME=${REMOTEENVS_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh
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

        echo "Remove Cluster, IP Address for Ingressgateway, IP Address for Remote Environments"
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

function createGroup() {
    shout "Create Azure group"
    date

    az group create --name "${RS_GROUP}" --location "${REGION}"
}

function installCluster() {
    shout "Install Kubernetes on Azure"
    date

    echo "Find latest cluster version"
    CLUSTER_VERSION=$(az aks get-versions -l "${REGION}" | jq '.orchestrators|.[]|select(.orchestratorVersion | contains("'"${CLUSTER_K8S_VERSION}"'"))' | jq -s '.' | jq -r 'sort_by(.orchestratorVersion | split(".") | map(tonumber)) | .[-1].orchestratorVersion')
    echo "Latest available version is: ${CLUSTER_VERSION}"

    az aks create \
      --resource-group "${RS_GROUP}" \
      --name "${CLUSTER_NAME}" \
      --node-count 3 \
      --node-vm-size "${CLUSTER_SIZE}" \
      --kubernetes-version "${CLUSTER_VERSION}" \
      --enable-addons "${CLUSTER_ADDONS}" \
      --service-principal "${AZURE_SUBSCRIPTION_APP_ID}" \
      --client-secret "${AZURE_SUBSCRIPTION_SECRET}" \
      --generate-ssh-keys
}

function azureAuthenticating() {
    shout "Authenticating to azure"
    date

    az login --service-principal -u "${AZURE_SUBSCRIPTION_APP_ID}" -p "${AZURE_SUBSCRIPTION_SECRET}" --tenant "${AZURE_SUBSCRIPTION_TENANT}"
    az account set --subscription "${AZURE_SUBSCRIPTION_ID}"
}

function createPublicIPandDNS() {
    CLUSTER_RS_GROUP=$(az aks show -g "${RS_GROUP}" -n "${CLUSTER_NAME}" --query nodeResourceGroup -o tsv)

    # IP address and DNS for Ingressgateway
    shout "Reserve IP Address for Ingressgateway"
	date

    GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"
    az network public-ip create -g "${CLUSTER_RS_GROUP}" -n "${GATEWAY_IP_ADDRESS_NAME}" -l "${REGION}" --allocation-method static

    GATEWAY_IP_ADDRESS=$(az network public-ip show -g "${CLUSTER_RS_GROUP}" -n "${GATEWAY_IP_ADDRESS_NAME}" --query ipAddress -o tsv)
    echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

    shout "Create DNS Record for Ingressgateway IP"
	date

    GATEWAY_DNS_FULL_NAME="*.${DOMAIN}."
    IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-dns-record.sh

    # IP address and DNS for Remote Environments
    shout "Reserve IP Address for Remote Environments"
	date

    REMOTEENVS_IP_NAME="remoteenvs-${STANDARIZED_NAME}"
    az network public-ip create -g "${CLUSTER_RS_GROUP}" -n "${REMOTEENVS_IP_NAME}" -l "${REGION}" --allocation-method static

    REMOTEENVS_IP_ADDRESS=$(az network public-ip show -g "${CLUSTER_RS_GROUP}" -n "${REMOTEENVS_IP_NAME}" --query ipAddress -o tsv)
    echo "Created IP Address for Remote Environments: ${REMOTEENVS_IP_ADDRESS}"

    shout "Create DNS Record for Remote Environments IP"
    date

    REMOTEENVS_DNS_FULL_NAME="gateway.${DOMAIN}."
    IP_ADDRESS=${REMOTEENVS_IP_ADDRESS} DNS_FULL_NAME=${REMOTEENVS_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-dns-record.sh
}

function addGithubDexConnector() {
    shout "Add Github Dex Connector"
    date
    pushd "${KYMA_PROJECT_DIR}/test-infra/development/tools"
    dep ensure -v -vendor-only
    popd
    export DEX_CALLBACK_URL="https://dex.${CLUSTER_NAME}.build.kyma-project.io/callback"
    go run "${KYMA_PROJECT_DIR}/test-infra/development/tools/cmd/enablegithubauth/main.go"
}

function generateAndExportLetsEncryptCert() {
	shout "Generate lets encrypt certificate"
	date

    mkdir letsencrypt
    cp "${GOOGLE_APPLICATION_CREDENTIALS}" letsencrypt
    docker run  --name certbot \
        --rm  \
        -v "$(pwd)/letsencrypt:/etc/letsencrypt"    \
        certbot/dns-google \
        certonly \
        -m "kyma.bot@sap.com" \
        --agree-tos \
        --no-eff-email \
        --dns-google \
        --dns-google-credentials /etc/letsencrypt/service-account.json \
        --server https://acme-v02.api.letsencrypt.org/directory \
        --dns-google-propagation-seconds=600 \
        -d "*.${DOMAIN}"

    TLS_CERT=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/fullchain.pem | tr -d '\n')
    export TLS_CERT
    TLS_KEY=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/privkey.pem   | tr -d '\n')
    export TLS_KEY
}

function setupKubeconfig() {
    shout "Setup kubeconfig and create ClusterRoleBinding"
    date

    az aks get-credentials --resource-group "${RS_GROUP}" --name "${CLUSTER_NAME}"
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(az account show | jq -r .user.name)"
}

function installTiller() {
    shout "Install tiller"
    date

    "${KYMA_SCRIPTS_DIR}"/install-tiller.sh
}

function waitUntilInstallerApiAvailable() {
    shout "Waiting for Installer API"

	attempts=5
    for ((i=1; i<=attempts; i++)); do
        numberOfLines=$(kubectl api-versions | grep -c "installer.kyma-project.io")
        if [[ "$numberOfLines" == "1" ]]; then
            echo "API found"
            break
        elif [[ "${i}" == "${attempts}" ]]; then
            echo "ERROR: API not found, exit"
            exit 1
        fi

        echo "Sleep for 3 seconds"
        sleep 3
    done
}

function installKyma() {
    shout "Install kyma"
    date

    echo "Prepare installation yaml files"
    KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/${STANDARIZED_NAME}/${REPO_OWNER}/${REPO_NAME}:${CURRENT_TIMESTAMP}"
    KYMA_INSTALLER_IMAGE="${KYMA_INSTALLER_IMAGE}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-image.sh

    KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"
	INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
	INSTALLER_CONFIG="${KYMA_RESOURCES_DIR}/installer-config-cluster.yaml.tpl"
	INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"

    echo "Apply Azure crb for healthz"
    kubectl apply -f "${KYMA_RESOURCES_DIR}"/azure-crb-for-healthz.yaml

    shout "Apply Kyma config"
    "${KYMA_SCRIPTS_DIR}"/concat-yamls.sh "${INSTALLER_YAML}" "${INSTALLER_CONFIG}" \
        | sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" \
        | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" \
        | sed -e "s/__DOMAIN__/${DOMAIN}/g" \
        | sed -e "s/__REMOTE_ENV_IP__/${REMOTEENVS_IP_ADDRESS}/g" \
        | sed -e "s#__TLS_CERT__#${TLS_CERT}#g" \
        | sed -e "s#__TLS_KEY__#${TLS_KEY}#g" \
        | sed -e "s/__EXTERNAL_PUBLIC_IP__/${GATEWAY_IP_ADDRESS}/g" \
        | sed -e "s/__SKIP_SSL_VERIFY__/true/g" \
        | sed -e "s/__VERSION__/0.0.1/g" \
        | sed -e "s/__SLACK_CHANNEL_VALUE__/${KYMA_ALERTS_CHANNEL}/g" \
        | sed -e "s#__SLACK_API_URL_VALUE__#${KYMA_ALERTS_SLACK_API_URL}#g" \
        | sed -e "s/__.*__//g" \
        | kubectl apply -f-

    waitUntilInstallerApiAvailable

	shout "Trigger installation"
	date

    sed -e "s/__VERSION__/0.0.1/g" "${INSTALLER_CR}"  | sed -e "s/__.*__//g" | kubectl apply -f-
    kubectl label installation/kyma-installation action=install
    "${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout 80m

    if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
        shout "Create DNS Record for Apiserver proxy IP"
        date
        APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        APISERVER_DNS_FULL_NAME="apiserver.${DOMAIN}."
        IP_ADDRESS=${APISERVER_IP_ADDRESS} DNS_FULL_NAME=${APISERVER_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
    fi
}

function installStabilityChecker() {
    shout "Install stability-checker"
    date

	STATS_FAILING_TEST_REGEXP=${STATS_FAILING_TEST_REGEXP:-"'([0-9A-Za-z_-]+)' (?:has Failed status?|failed due to too long Running status?|failed due to too long Pending status?|failed with Unknown status?)"}
	STATS_SUCCESSFUL_TEST_REGEXP=${STATS_SUCCESSFUL_TEST_REGEXP:-"Test of '([0-9A-Za-z_-]+)' was successful"}
	STATS_ENABLED="true"

	SC_DIR=${TEST_INFRA_SOURCES_DIR}/stability-checker

	kubectl create -f "${SC_DIR}/local/provisioning.yaml"
	bash "${SC_DIR}/local/helpers/isready.sh" kyma-system app  stability-test-provisioner
	kubectl exec stability-test-provisioner -n kyma-system --  mkdir -p /home/input
	kubectl cp "${KYMA_SCRIPTS_DIR}/testing.sh" stability-test-provisioner:/home/input/ -n kyma-system
	kubectl cp "${KYMA_SCRIPTS_DIR}/utils.sh" stability-test-provisioner:/home/input/ -n kyma-system
	kubectl cp "${KYMA_SCRIPTS_DIR}/testing-common.sh" stability-test-provisioner:/home/input/ -n kyma-system

	kubectl exec stability-test-provisioner -n kyma-system --  mkdir -p /root/.helm
    kubectl cp "$(helm home)/ca.pem"   stability-test-provisioner:/root/.helm/ -n kyma-system
    kubectl cp "$(helm home)/cert.pem" stability-test-provisioner:/root/.helm/ -n kyma-system
    kubectl cp "$(helm home)/key.pem"  stability-test-provisioner:/root/.helm/ -n kyma-system

	kubectl delete pod -n kyma-system stability-test-provisioner

    # create a secret with service account used for storing logs
    kubectl create secret generic sa-stability-fluentd-storage-writer --from-file=service-account.json=/etc/credentials/sa-stability-fluentd-storage-writer/service-account.json -n kyma-system

	helm install --set clusterName="${CLUSTER_NAME}" \
	        --set logsPersistence.enabled=true \
	        --set slackClientWebhookUrl="${SLACK_CLIENT_WEBHOOK_URL}" \
	        --set slackClientChannelId="${STABILITY_SLACK_CLIENT_CHANNEL_ID}" \
	        --set slackClientToken="${SLACK_CLIENT_TOKEN}" \
	        --set stats.enabled="${STATS_ENABLED}" \
	        --set stats.failingTestRegexp="${STATS_FAILING_TEST_REGEXP}" \
	        --set stats.successfulTestRegexp="${STATS_SUCCESSFUL_TEST_REGEXP}" \
	        --set testResultWindowTime="${TEST_RESULT_WINDOW_TIME}" \
	        "${SC_DIR}/deploy/chart/stability-checker" \
	        --namespace=kyma-system \
	        --name=stability-checker \
	        --wait \
	        --timeout=600
}


init
azureAuthenticating

DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --project="${CLOUDSDK_CORE_PROJECT}" --format="value(dnsName)")"
export DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"

cleanup
addGithubDexConnector

createGroup
installCluster

createPublicIPandDNS
generateAndExportLetsEncryptCert

setupKubeconfig
installTiller
installKyma
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-helm-certs.sh"

installStabilityChecker
