#!/usr/bin/env bash

# TEST_INFRA_SOURCES_DIR is exported in parent script
# shellcheck source=prow/scripts/lib/azure.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/azure.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

validateAzureGatewayEnvironment() {
    shout "Validating Azure Blob Gateway environment"; date

    requiredVars=(
        AZURE_RS_GROUP
        AZURE_REGION
        AZURE_SUBSCRIPTION_ID
        AZURE_CREDENTIALS_FILE
        AZURE_STORAGE_ACCOUNT_NAME
    )

    utils::checkRequiredVars ${requiredVars[@]}

    echo "Environment validated"; date
}

beforeTest() {
    validateAzureGatewayEnvironment
    az::login "$AZURE_CREDENTIALS_FILE"
    az:set_subscription "$AZURE_SUBSCRIPTION_ID"
    createResourceGroup
    createStorageAccount
}

createResourceGroup() {
    shout "Create Azure Resource Group ${AZURE_RS_GROUP}"; date

    if [[ $(az group exists --name "${AZURE_RS_GROUP}" -o json) == true ]]; then
        echo "Azure Resource Group ${AZURE_RS_GROUP} exists"; date
        return
    fi

    az group create \
        --name "${AZURE_RS_GROUP}" \
        --location "${AZURE_REGION}" \
        --tags "created-by=prow"

    # Wait until resource group will be visible in azure.
    counter=0
    until [[ $(az group exists --name "${AZURE_RS_GROUP}" -o json) == true ]]; do
        sleep 15
        counter=$(( counter + 1 ))
        if (( counter == 5 )); then
            echo -e "---\nAzure resource group ${AZURE_RS_GROUP} still not present after one minute wait.\n---"
            exit 1
        fi
    done

    echo "Resource Group created"; date
}

createStorageAccount() {
    shout "Create ${AZURE_STORAGE_ACCOUNT_NAME} Storage Account"; date

    az storage account create \
        --name "${AZURE_STORAGE_ACCOUNT_NAME}" \
        --resource-group "${AZURE_RS_GROUP}" \
        --tags "created-at=$(date +%s)" "created-by=prow" "ttl=10800"

    echo "Storage Account created"; date
}

afterTest() {
    shout "Delete ${AZURE_STORAGE_ACCOUNT_NAME} Storage Account"; date

    az storage account delete \
        --name "${AZURE_STORAGE_ACCOUNT_NAME}" \
        --resource-group "${AZURE_RS_GROUP}" \
        --yes

    echo "Storage Account deleted"; date
}

installOverrides() {
    shout "Installing Azure Minio Gateway overrides"; date

    local -r AZURE_ACCOUNT_KEY=$(az storage account keys list --account-name "${AZURE_STORAGE_ACCOUNT_NAME}" --resource-group "${AZURE_RS_GROUP}" --query "[?keyName=='key1'].value" --output tsv)

    local -r ASSET_STORE_RESOURCE_NAME="azure-minio-overrides"
    kubectl create -n kyma-installer secret generic "${ASSET_STORE_RESOURCE_NAME}" --from-literal "minio.secretKey=${AZURE_ACCOUNT_KEY}" --from-literal "minio.accessKey=${AZURE_STORAGE_ACCOUNT_NAME}"
    kubectl label -n kyma-installer secret "${ASSET_STORE_RESOURCE_NAME}" "installer=overrides" "component=assetstore" "kyma-project.io/installation="

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "${ASSET_STORE_RESOURCE_NAME}" \
        --data "minio.persistence.enabled=false" \
        --data "minio.azuregateway.enabled=true" \
        --data "minio.DeploymentUpdate.type=RollingUpdate" \
        --data "minio.DeploymentUpdate.maxSurge=0" \
        --data "minio.DeploymentUpdate.maxUnavailable=50%" \
        --label "component=assetstore"
    
    shout "Overrides installed"; date
}
