#!/usr/bin/env bash

# TEST_INFRA_SOURCES_DIR is exported in parent script
# shellcheck source=prow/scripts/lib/azure.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/azure.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"

validateAzureGatewayEnvironment() {
    log::info "Validating Azure Blob Gateway environment"

    requiredVars=(
        AZURE_RS_GROUP
        AZURE_REGION
        AZURE_SUBSCRIPTION_ID
        AZURE_CREDENTIALS_FILE
        AZURE_STORAGE_ACCOUNT_NAME
    )

    utils::check_required_vars "${requiredVars[@]}"

    log::info "Environment validated"
}

beforeTest() {
    validateAzureGatewayEnvironment
    az::authenticate -f "$AZURE_CREDENTIALS_FILE"
    az::set_subscription -s "$AZURE_SUBSCRIPTION_ID"
    az::create_resource_group -g "$AZURE_RS_GROUP" -r "$AZURE_REGION" -t "created-by=prow"
    az::create_storage_account \
        -n "$AZURE_STORAGE_ACCOUNT_NAME" \
        -g "$AZURE_RS_GROUP" \
        -t "created-at=$(date +%s)" \
        -t "created-by=prow" \
        -t "ttl=10800"
}

installOverrides() {
    log::info "Installing Azure Minio Gateway overrides"

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
    
    log::info "Overrides installed"
}
