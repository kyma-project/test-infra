#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${LIBDIR}/utils.sh"

# az::verify_deps checks if all required commands are available
function az::verify_deps {
  log::info "Verify dependencies checks if all required commands are available"
    if ! [[ -x $(command -v az) ]]; then
        log::error "'az' command not found in \$PATH. Exiting..."
        exit 1
    else
        log::info "Azure CLI Version:"
        az version
    fi
    if ! [[ -x $(command -v jq) ]]; then
        log::error "'jq' command not found in \$PATH. Exiting..."
        exit 1
    else
        log::info "jq version:"
        jq --version
    fi
}

# az::authenticate logs in to the azure service using provided credentials file in the function argument.
# Arguments:
# required:
# f - Azure login credentials file
# Function accepts JSON file formatted below:
# {
#   "tenant_id": "tenant_id",
#   "app_id": "subscription_app_id",
#   "secret": "subscription_secret"
# }
function az::authenticate {
    local OPTIND
    local azureSubscriptionTenant
    local azureSubscriptionAppID
    local azureSubscriptionSecret
    local azureCredentialsFile

    # Check the provided credentials in the argument.
    # Use arguments to avoid exporting sensitive values.
    log::info "Check the provided credentials in the argument"
    while getopts ":f:" opt; do
        case $opt in
            f)
                azureCredentialsFile="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done
    utils::check_empty_arg "$azureCredentialsFile" "Missing account credentials, please provide proper credentials"

    azureSubscriptionTenant=$(jq -r '.tenant_id' "$azureCredentialsFile")
    azureSubscriptionAppID=$(jq -r '.app_id' "$azureCredentialsFile")
    azureSubscriptionSecret=$(jq -r '.secret' "$azureCredentialsFile")

    # login
    log::info "Logging in to Azure"
    az login \
        --service-principal \
        -u "${azureSubscriptionAppID}" \
        -p "${azureSubscriptionSecret}" \
        --tenant "${azureSubscriptionTenant}"
    log::info "Successfully logged-in!"
}

# az::set_subscription sets the subscription using provided subscription ID in the argument.
# Arguments:
# required:
# s - subscription ID
function az::set_subscription {

    local OPTIND
    local azureSubscription

    while getopts ":s:" opt; do
        case $opt in
            s)
                azureSubscription="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done
    utils::check_empty_arg "$azureSubscription"  "missing Azure Subscription ID, please provide proper azure subscription ID in the argument. Exiting..."
    log::info "Setting Azure subscription..."
    az account set \
        --subscription "$azureSubscription"
}

# az::create_resource_group creates resource group in a given region
#
# Arguments:
# required:
# g - resource group name to be created
# r - region in which group should be created
# optional:
# t - tags, can be used mutiple times to pass an array
function az::create_resource_group {

    local OPTIND
    local resourceGroup
    local azureRegion
    local groupTags
    log::info "Check the provided group name, region and tags in the argument"
    while getopts ":g:r:t:" opt; do
        case $opt in
            g)
                resourceGroup="$OPTARG" ;;
            r)
                azureRegion="$OPTARG" ;;
            t)
                if [ -n "$OPTARG" ]; then
                    groupTags+=("$OPTARG")
                fi ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$resourceGroup" "Resource group name is empty. Exiting..."
    utils::check_empty_arg "$azureRegion" "Azure region name is empty. Exiting..."

    if [[ $(az group exists --name "${resourceGroup}" -o json) == true ]]; then
        log::info "Azure Resource Group ${resourceGroup} exists"
        return
    fi

    log::info "Creating resouce group \"$resourceGroup\" in a region \"$azureRegion\""
    if [ ${#groupTags[@]} != 0 ]; then
        az group create \
        --name "${resourceGroup}" \
        --location "${azureRegion}" \
        --tags "${groupTags[@]}"
    else
        az group create \
        --name "${resourceGroup}" \
        --location "${azureRegion}"
    fi

    until [[ $(az group exists --name "${resourceGroup}" -o json) == true ]]; do
        sleep 15
		counter=$(( counter + 1 ))
		if (( counter == 5 )); then
			log::error "\\n---\\nAzure resource group ${resourceGroup} still not present after one minute wait.\\n---"
			exit 1
		fi
	done
    log::info "Resource Group created"
}

# az::delete_resource_group deletes resource group in a given region
#
# Arguments:
# required:
# g - resource group name to be deleted
#
function az::delete_resource_group {

    local OPTIND
    local resourceGroup

    local deleteReturnCode

    while getopts ":g:" opt; do
        case $opt in
            g)
                resourceGroup="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$resourceGroup" "Resource group name is empty. Exiting..."

    log::info "Remove group"
    if [[ $(az group exists --name "${RS_GROUP}" -o json) == true ]]; then
        az group delete \
            -n "${resourceGroup}" \
            -y
        deleteReturnCode=$?
        if [[ ${deleteReturnCode} -ne 0 ]]; then
            log::error "Failed to delete ResourceGrouop : ${resourceGroup}"
            exit 1
        else
            log::info "ResourceGroup deleted : ${resourceGroup}"
        fi
    else
		log::error "Azure group does not exist"
        return 1
	fi
}

# az::create_storage_account creates storage accont resource group in a given group
#
# Arguments:
# required:
# n - storage account name to be created
# g - resource group name
# optional:
# t - tags, can be used mutiple times to pass an array
function az::create_storage_account {

    local OPTIND
    local resourceGroup
    local accountName
    local groupTags

    while getopts ":g:n:t:" opt; do
        case $opt in
            g)
                resourceGroup="$OPTARG" ;;
            n)
                accountName="$OPTARG" ;;
            t)
                if [ -n "$OPTARG" ]; then
                    groupTags+=("$OPTARG")
                fi ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$resourceGroup" "Resource group name is empty. Exiting..."
    utils::check_empty_arg "$accountName" "Account name is empty. Exiting..."

    log::info "Creating ${accountName} Storage Account"
    if [ ${#groupTags[@]} != 0 ]; then
        az storage account create \
            --name "${accountName}" \
            --resource-group "${resourceGroup}" \
            --tags "${groupTags[@]}"
    else
        az storage account create \
            --name "${accountName}" \
            --resource-group "${resourceGroup}"
    fi
    log::info "Storage Account created"
}

# az::delete_storage_account deletes storage accont resource group in a given group
#
# Arguments:
# required:
# n - storage account name to be deleted
# g - resource group name
function az::delete_storage_account {

    local OPTIND
    local resourceGroup
    local accountName

    while getopts ":g:n:" opt; do
        case $opt in
            g)
                resourceGroup="$OPTARG" ;;
            n)
                accountName="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$resourceGroup" "Resource group name is empty. Exiting..."
    utils::check_empty_arg "$accountName" "Account name is empty. Exiting..."

    log::info "Deleting ${accountName} Storage Account"

    az storage account delete \
        --name "${accountName}" \
        --resource-group "${resourceGroup}" \
        --yes
    log::info "Storage Account deleted"
}

# az::provision_k8s_cluster creates an AKS cluster
#
# Arguments:
# required:
# c - cluster name

# g - azure resource group
# r - azure region

# s - azure cluster size
# v - desired k8s cluster version
# a - addidional AKS addons
# f - credentials file, refer to az::authenticate
#
function az::provision_k8s_cluster {

    local OPTIND
    local clusterName
    local resourceGroup
    local azureRegion
    local clusterSize
    local clusterVersion
    local clusterVersionPrecise
    local aksAddons
    local credentialsFile

    # Check the provided credentials in the argument.
    # Use arguments to avoid exporting sensitive values.
    while getopts ":c:g:r:s:v:a:f:" opt; do
        case $opt in
            c)
                clusterName="$OPTARG" ;;
            g)
                resourceGroup="$OPTARG" ;;
            r)
               azureRegion="$OPTARG" ;;
            s)
                clusterSize="$OPTARG" ;;
            v)
                clusterVersion="$OPTARG" ;;
            a)
                aksAddons="$OPTARG" ;;
            f)
                credentialsFile="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$clusterName" "Missing cluster name, please provide proper cluster name"
    utils::check_empty_arg "$resourceGroup" "Missing resource group name, please provide proper resource group name"
    utils::check_empty_arg "$azureRegion" "Missing Azure region name, please provide proper Azure region name"
    utils::check_empty_arg "$clusterSize" "Missing cluster size, please provide proper cluster size"
    utils::check_empty_arg "$clusterVersion" "Missing cluster name, please provide proper cluster name"
    utils::check_empty_arg "$aksAddons" "Missing AKS addons, please provide proper AKS addons"
    utils::check_empty_arg "$credentialsFile" "Missing credentials file name, please provide proper credentials file name"


    log::info "Provisioning AKS cluster"
    clusterVersionPrecise=$(az aks get-versions -l "${azureRegion}" | jq '.orchestrators|.[]|select(.orchestratorVersion | contains("'"${clusterVersion}"'"))' | jq -s '.' | jq -r 'sort_by(.orchestratorVersion | split(".") | map(tonumber)) | .[-1].orchestratorVersion')
    log::info "Latest available version is: ${clusterVersionPrecise}"

    az aks create \
        --resource-group "${resourceGroup}" \
        --name "${clusterName}" \
        --node-count 3 \
        --node-vm-size "${clusterSize}" \
        --kubernetes-version "${clusterVersionPrecise}" \
        --enable-addons "${aksAddons}" \
        --service-principal "$(jq -r '.app_id' "$credentialsFile")" \
        --client-secret "$(jq -r '.secret' "$credentialsFile")" \
        --generate-ssh-keys \
        --zones 1 2 3

    # run oom debug pod
    if [ "${DEBUG_COMMANDO_OOM}" = "true" ]; then
        # run oom debug pod
        utils::debug_oom
    fi
}

# az::deprovision_k8s_cluster removes an AKS cluster
#
# Arguments:
# required:
# c - cluster name
# g - azure resource group
#
function az::deprovision_k8s_cluster {

    local OPTIND
    local clusterName
    local resourceGroup

    local deleteReturnCode

    while getopts ":c:g:" opt; do
        case $opt in
            c)
                clusterName="$OPTARG" ;;
            g)
                resourceGroup="$OPTARG" ;;

            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$clusterName" "Missing cluster name, please provide proper cluster name"
    utils::check_empty_arg "$resourceGroup" "Missing resource group name, please provide proper resource group name"

    log::info "Deprovisioning AKS cluster"
    az aks delete \
        -g "${resourceGroup}" \
        -n "${clusterName}" \
        -y

    deleteReturnCode=$?
    if [[ ${deleteReturnCode} -ne 0 ]]; then
        log::error "Failed delete cluster : ${clusterName}"
        return 1
    else
        log::info "Cluster and IP address for Ingressgateway deleted"
    fi
}

# az ::reserve_ip_address reserves IP address
#
# Arguments:
# required:
# g - resource group (must match the cluster RS group)
# n - IP address name used for identification
# r - Azure region in which IP is reserved
#
# Returns:
# az_reserve_ip_address_return_ip_address - reserved ip address -> GATEWAY_IP_ADDRESS
#
function az::reserve_ip_address {

    local OPTIND
    local resourceGroup
    local ipAddressName
    local azureRegion

    while getopts ":g:n:r:" opt; do
        case $opt in
            g)
                resourceGroup="$OPTARG" ;;
            n)
                ipAddressName="$OPTARG" ;;
            r)
                azureRegion="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$resourceGroup" "Resource group name is empty. Exiting..."
    utils::check_empty_arg "$ipAddressName" "IP address name is empty. Exiting..."
    utils::check_empty_arg "$azureRegion" "Azure region name is empty. Exiting..."

    log::info "Reserve IP Address for Ingressgateway"
    if az network public-ip create -g "${resourceGroup}" -n "${ipAddressName}" -l "${azureRegion}" --allocation-method static --sku Standard; then
        # shellcheck disable=SC2034
        az_reserve_ip_address_return_ip_address=$(az network public-ip show -g "${resourceGroup}" -n "${ipAddressName}" --query ipAddress -o tsv)
        log::success "Created IP Address for Ingressgateway: ${az_reserve_ip_address_return_ip_address}"
    else
        log::error "Could not create IP address. Exiting..."
        exit 1
    fi
}

# Arguments:
# required:
# r - resource group
# c - cluster name
#
# Returns:
# az_get_cluster_resource_group_return_resource_group - name of the cluster resource group
#
function az::get_cluster_resource_group {

    local OPTIND
    local resourceGroup
    local clusterName
    local clusterResourceGroup
    local tmpStatus

    while getopts ":r:c:" opt; do
        case $opt in
            r)
                resourceGroup="$OPTARG" ;;
            c)
                clusterName="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$resourceGroup" "Resource group name is empty. Exiting..."
    utils::check_empty_arg "$clusterName" "Cluster name is empty. Exiting..."

    clusterResourceGroup=$(az aks show -g "${resourceGroup}" -n "${clusterName}" --query nodeResourceGroup -o tsv)
    tmpStatus=$?
    if [[ ${tmpStatus} -ne 0 ]]; then
        log::error "Failed to get nodes resource group."
        return 1
    fi
    # shellcheck disable=SC2034
    az_get_cluster_resource_group_return_resource_group="$clusterResourceGroup"
}

# This check will trigger everytime the file is sourced.
# This should allow easy checking for the related requirements without copying loads of code.
# If more checks will be needed we should add another function there.
az::verify_deps
# (2025-03-04)