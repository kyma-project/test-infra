#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

function az::verify_deps {
  if ! [[ -x $(command -v az) ]]; then
    log::error "'az' command not found in \$PATH. Exiting..."
    exit 1
  else
    echo "Azure CLI Version:"
    az version
  fi
  if ! [[ -x $(command -v jq) ]]; then
    log::error "'jq' command not found in \$PATH. Exiting..."
    exit 1
  else
    echo "jq version:"
    jq --version
  fi
}

# az::login logs in to the azure service using provided credentials file in the function argument.
# Function accepts JSON file formatted below:
# {
#   "tenant_id": "tenant_id",
#   "app_id": "subscription_app_id",
#   "secret": "subscription_secret"
# }
function az::login {
  local AZURE_SUBSCRIPTION_TENANT
  local AZURE_SUBSCRIPTION_APP_ID
  local AZURE_SUBSCRIPTION_SECRET

  # Check the provided credentials in the argument.
  # Use arguments to avoid exporting sensitive values.
  if [[ -z "$1" ]]; then
    log::error "Azure credentials file not provided. please provide azure credentials filepath in the argument. Exiting..."
    exit 1
  elif [[ ! -f "$1" ]]; then
    log::error "Azure credentials file not found. Make sure it is present under the provided filepath. Exiting..."
    exit 1
  fi
  AZURE_CREDENTIALS_FILE="$1"
  AZURE_SUBSCRIPTION_TENANT=$(jq -r '.tenant_id' "$AZURE_CREDENTIALS_FILE")
  AZURE_SUBSCRIPTION_APP_ID=$(jq -r '.app_id' "$AZURE_CREDENTIALS_FILE")
  AZURE_SUBSCRIPTION_SECRET=$(jq -r '.secret' "$AZURE_CREDENTIALS_FILE")

  # login
  log::info "Logging in to Azure"
  az login --service-principal -u "${AZURE_SUBSCRIPTION_APP_ID}" -p "${AZURE_SUBSCRIPTION_SECRET}" --tenant "${AZURE_SUBSCRIPTION_TENANT}"
  log::info "Successfully logged-in!"
}

# az::set_subscription sets the subscription using provided subscription ID in the argument.
function az::set_subscription {
  if [[ -z "$1" ]]; then
    log::error "Azure Subscription ID is not found. Please provide azure subscription ID in the argument. Exiting..."
    exit 1
  fi
  log::info "Setting Azure subscription..."
  az account set \
    --subscription "$1"
}

# az::create_resource_group creates resource group in a given region
#
# Arguments
# $1 - resource group name to be created
# $2 - region in which group should be created
function az::create_resource_group {
  if [[ -z "$1" ]]; then
    log::error "Resource group name not present. Provide resource group name as 1st argument. Exiting..."
    exit 1
  fi
  if [[ -z "$2" ]]; then
    log::error "Region not present. Provide region as 2nd argument. Exiting..."
    exit 1
  fi
  local rsGroup
  local azRegion
  rsGroup=$1
  azRegion=$2

  log::info "Creating resouce group \"$rsGroup\" in a region \"$azRegion\""
  az group create --name "${rsGroup}" --location "${azRegion}"
  until [[ $(az group exists --name "${rsGroup}" -o json) == true ]]; do
		sleep 15
		counter=$(( counter + 1 ))
		if (( counter == 5 )); then
			log::error "\n---\nAzure resource group ${rsGroup} still not present after one minute wait.\n---"
			exit 1
		fi
	done
}

# az::provision_aks_cluster creates an AKS cluster
#
# Required exported variables
# RS_GROUP - azure resource group
# REGION - azure region
# CLUSTER_SIZE - azure cluster size
# AKS_CLUSTER_VERSION - desired k8s cluster version
# CLUSTER_ADDONS - addidional AKS addons
# AZURE_CREDENTIALS_FILE - credentials file, refer to az::login
#
# Arguments
# $1 - cluster name
function az::provision_aks_cluster {
  if [[ -z "$1" ]]; then
    log::error "Cluster name not present. Provide cluster name as an argument. Exiting..."
    exit 1
  fi
  local CLUSTER_NAME
  CLUSTER_NAME=$1

  log::info "Provisioning AKS cluster"
  AKS_CLUSTER_VERSION_PRECISE=$(az aks get-versions -l "${REGION}" | jq '.orchestrators|.[]|select(.orchestratorVersion | contains("'"${AKS_CLUSTER_VERSION}"'"))' | jq -s '.' | jq -r 'sort_by(.orchestratorVersion | split(".") | map(tonumber)) | .[-1].orchestratorVersion')
	log::info "Latest available version is: ${AKS_CLUSTER_VERSION_PRECISE}"

  az aks create \
      --resource-group "${RS_GROUP}" \
      --name "${CLUSTER_NAME}" \
      --node-count 3 \
      --node-vm-size "${CLUSTER_SIZE}" \
      --kubernetes-version "${AKS_CLUSTER_VERSION_PRECISE}" \
      --enable-addons "${CLUSTER_ADDONS}" \
      --service-principal "$(jq -r '.app_id' "$AZURE_CREDENTIALS_FILE")" \
      --client-secret "$(jq -r '.secret' "$AZURE_CREDENTIALS_FILE")" \
      --generate-ssh-keys \
      --zones 1 2 3

  # Schedule pod with oom finder.
  if [ "${DEBUG_COMMANDO_OOM}" = "true" ]; then
      # run oom debug pod
      utils::debug_oom
  fi

}

# az ::reserve_ip_address reserves IP address and returns it to STDOUT
#
# Required exported variables
# RS_GROUP - resource group (must match the cluster RS group)
# REGION - Azure region in which IP is reserved
#
# Arguments
# $1 - IP address name used for identification
function az::reserve_ip_address {
  if [[ -z "$1" ]]; then
    log::error "IP address name not present. Provide IP address name as an argument. Exiting..."
    exit 1
  fi
  local ipName
  ipName=$1

  if az network public-ip create -g "${RS_GROUP}" -n "${ipName}" -l "${REGION}" --allocation-method static; then
    az network public-ip show -g "${RS_GROUP}" -n "${ipName}" --query ipAddress -o tsv
  else
    log::error "Could not create IP address. Exiting..."
    exit 1
  fi
}

# This check will trigger everytime the file is sourced.
# This should allow easy checking for the related requirements without copying loads of code.
# If more checks will be needed we should add another function there.
az::verify_deps
