#!/usr/bin/env bash

if [ -f "../../prow/scripts/library.sh" ]; then
    export TEST_INFRA_SOURCES_DIR="../.."

elif [ -f "../test-infra/prow/scripts/library.sh" ]; then
    export TEST_INFRA_SOURCES_DIR="../test-infra"

else
	echo "File 'library.sh' can't be found."
    exit 1;
fi

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

function azureAuthentication() {
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
}

function addGithubDexConnector() {
    shout "Add Github Dex Connector"
    date
    pushd "${KYMA_PROJECT_DIR}/test-infra/development/tools" || exit 1
    export DEX_CALLBACK_URL="https://dex.${CLUSTER_NAME}.build.kyma-project.io/callback"
    if [ -x /prow-tools/enablegithubauth ];
    then
      /prow-tools/enablegithubauth
    else
      go run "${KYMA_PROJECT_DIR}/test-infra/development/tools/cmd/enablegithubauth/main.go"
    fi
    popd || exit 1
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

function setupKubeconfig() {
    shout "Setup kubeconfig and create ClusterRoleBinding"
    date

    az aks get-credentials --resource-group "${RS_GROUP}" --name "${CLUSTER_NAME}"
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(az account show | jq -r .user.name)"
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