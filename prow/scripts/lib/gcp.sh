#!/usr/bin/env bash
LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${LIBDIR}/utils.sh"

# gcp::provision_k8s_cluster creates a GKE kubernetes cluster
#
# Arguments:
#
# Required arguments:
# c - GKE cluster name
# p - GCP project name
# v - GKE cluster version
# j - prowjob name
# J - prowjob id
#
# Optional arguments:
# l - optional additional labels for the cluster
# t - cluster ttl hours, default 3
# z - zone in which the new zonal cluster will be created
# m - machine type to use in a cluster, default n1-standard-4
# R - region in which the new regional cluster will be created
# n - cluster worker nodes count, for regional clusters it's per zone count
# N - gcp network name which use for new cluster, default default
# S - gcp subnet name which use for new cluster
# C - release channel to use for new cluster
# i - cluster node vm image type to use for new cluster
# g - gcp security group domain to use for new cluster GCLOUD_SECURITY_GROUP_DOMAIN
# r - it true provision regional cluster
# s - if true enable using stackdriver for new cluster
# D - if true enable using ssd disks for new cluster
# e - if true enable pod security policy for new cluster
# P - path to test-infra sources
function gcp::provision_k8s_cluster {

    local OPTIND
    # required arguments
    local clusterName
    local gcpProjectName
    local gkeClusterVersion
    local prowjobName
    local prowjobID
    # default values
    local ttlHours="3"
    local computeZone="europe-west4-b"
    local machineType="n1-standard-4"
    local computeRegion="europe-west4"
    local numNodes="3"
    local nodesPerZone="1"
    local networkName="default"
    local provisionRegionalCluster="false"
    local enableSSD="false"
    local enablePSP="false"
    local enableStackdriver="false"
    local currentTimestampReadableParam
    local currentTimestampParam
    readonly currentTimestampReadableParam=$(date +%Y%m%d)
    readonly currentTimestampParam=$(date +%s)
    local testInfraSourcesDir="/home/prow/go/src/github.com/kyma-project"


    while getopts ":c:p:v:l:t:z:m:R:n:N:S:C:i:g:r:s:D:e:P:j:J:" opt; do
        case $opt in
            c)
                clusterName="${OPTARG:0:40}" ;;
            p)
                gcpProjectName="$OPTARG" ;;
            v)
                gkeClusterVersion="$OPTARG" ;;
            j)
                prowjobName="$OPTARG";;
            J)
                prowjobID="$OPTARG";;
            t)
                ttlHours=${OPTARG:-$ttlHours} ;;
            z)
                computeZone=${OPTARG:-$computeZone} ;;
            m)
                machineType=${OPTARG:-$machineType} ;;
            R)
                computeRegion=${OPTARG:-$computeRegion} ;;
            N)
                networkName=${OPTARG:-$networkName} ;;
            r)
                provisionRegionalCluster=${OPTARG:-$provisionRegionalCluster} ;;
            s)
                enableStackdriver=${OPTARG:-$enableStackdriver} ;;
            D)
                enableSSD=${OPTARG:-$enableSSD} ;;
            e)
                enablePSP=${OPTARG:-$enablePSP} ;;
            P)
                testInfraSourcesDir=${OPTARG:-$testInfraSourcesDir} ;;
            l)
                if [ -n "$OPTARG" ]; then
                    local additionalLabels="$OPTARG"
                fi ;;
            n)
                if [ -n "$OPTARG" ]; then
                    local nodesCount="$OPTARG"
                fi ;;
            S)
                if [ -n "$OPTARG" ]; then
                    local subnetName="$OPTARG"
                fi ;;
            C)
                if [ -n "$OPTARG" ]; then
                    local gkeReleaseChannel="$OPTARG"
                fi ;;
            i)
                if [ -n "$OPTARG" ]; then
                    local imageType="$OPTARG"
                fi ;;
            g)
                if [ -n "$OPTARG" ]; then
                    local gcpSecurityGroupDomain="$OPTARG"
                fi ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done


    # check required arguments
    utils::check_empty_arg "$clusterName" "Cluster name not provided."
    utils::check_empty_arg "$gcpProjectName" "GCP project name not provided."
    utils::check_empty_arg "$gkeClusterVersion" "GKE cluster version not provided."
    utils::check_empty_arg "$prowjobName" "prowjob name not provided."
    utils::check_empty_arg "$prowjobID" "prowjob ID not provided."

    log::info "Replacing underscore with dashes in cluster name."
    clusterName=$(echo "$clusterName" | tr '_' '-')
    # Cluster name must be less than 40 characters
    if [ "${#clusterName}" -ge 40 ]; then
        log::error "Cluster name must be less than 40 characters"
        exit 1
    fi

    log::banner "Provision cluster: $clusterName"

    local kubeDnsPatchPath="$testInfraSourcesDir/prow/scripts/resources/kube-dns-stub-domains-patch.yaml"
    local params

    # mandatory labels
    local cleanerLabels="created-at=$currentTimestampParam,created-at-readable=$currentTimestampReadableParam,ttl=$ttlHours"
    local jobLabels="job=$prowjobName,job-id=$prowjobID"
    local clusterLabels="cluster=$clusterName,volatile=true"

    # optional labels
    if [ -n "$additionalLabels" ]; then
        additionalLabels=",$additionalLabels"
    fi

    # build labels parameter value
    local labels="$jobLabels,"
    labels+="$clusterLabels,"
    labels+="$cleanerLabels"
    labels+="$additionalLabels"

    # Resolving parameters
    # Mandatory parameters
    params+=("--project=$gcpProjectName")
    params+=("--cluster-version=$gkeClusterVersion")
    params+=("--labels=$labels")

    # Conditional parameters
    params+=("--machine-type=$machineType")
    if [ "$provisionRegionalCluster" = "true" ] ; then
        params+=("--region=$computeRegion")
        params+=("--num-nodes=${nodesCount:-$nodesPerZone}")
    else
        params+=("--zone=$computeZone")
        params+=("--num-nodes=${nodesCount:-$numNodes}")
    fi
    if [ "$networkName" ] && [ "$subnetName" ]; then
        params+=("--network=$networkName")
        params+=("--subnetwork=$subnetName")
    else
        params+=("--network=$networkName")
    fi

    # Optional parameters
    if [ "$gkeReleaseChannel" ]; then
        params+=("--release-channel=$gkeReleaseChannel")
    fi
    # serverless tests are failing when are running on a cluster with contianerD
    if [[ "$gkeReleaseChannel" == "rapid" ]]; then
        # set image type to the image that uses docker instead of containerD
        params+=("--image-type=cos")
    elif [ "$imageType" ]; then
        params+=("--image-type=$imageType")
    fi
    if [ "$enableStackdriver" = "true" ]; then
        params+=("--enable-stackdriver-kubernetes")
    fi
    if [ "$enableSSD" = "true" ]; then
        params+=("--disk-type=pd-ssd")
    fi
    if [ "$gcpSecurityGroupDomain" ]; then
        params+=("--security-group=gke-security-groups@$gcpSecurityGroupDomain")
    fi
    if [ "$enablePSP" = "true" ]; then
        params+=("--enable-pod-security-policy")
    fi

    log::info "Provisioning GKE cluster"
    gcloud beta container clusters create "$clusterName" "${params[@]}"
    log::info "Successfully created cluster $clusterName!"

    log::info "Patching kube-dns with stub domains"
    counter=0
    until [[ $(kubectl get cm kube-dns -n kube-system > /dev/null 2>&1; echo $?) == 0 ]]; do
        if (( counter == 5 )); then
            echo -e "kube-dns configmap not available after 5 tries, exiting"
            exit 1
        fi
        echo -e "Waiting for kube-dns to be available. Try $(( counter + 1 )) of 5"
        counter=$(( counter + 1 ))
        sleep 15
    done

    kubectl -n kube-system patch cm kube-dns --type merge --patch "$(cat "$kubeDnsPatchPath")"

    if [ "${DEBUG_COMMANDO_OOM}" = "true" ]; then
        # run oom debug pod
        utils::debug_oom
    fi
}


# gcp::authenticate authenticates to GCP.
# Arguments:
# required:
# c - google credentials file path
function gcp::authenticate {
    log::info "Preparing credentials"
    local OPTIND
    #required arguments
    local googleAppCredentials
    log::info "Check the provided credentials in the argument"
    while getopts ":c:" opt; do
        case $opt in
            c)
                googleAppCredentials="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$googleAppCredentials" "Missing account credentials, please provide proper credentials"

    log::info "Authenticating to gcloud"
    gcloud auth activate-service-account --key-file "$googleAppCredentials" || exit 1
}

# gcp::set_account activates already authenticated account
# Arguments:
# required:
# c - credentials to Google application
function gcp::set_account() {
    
    local OPTIND
    #required arguments
    local googleAppCredentials
    local clientEmail

    while getopts ":c:" opt; do
        case $opt in
            c)
                googleAppCredentials="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$googleAppCredentials" "Missing account credentials, please provide proper credentials"

    clientEmail=$(jq -r '.client_email' < "$googleAppCredentials")
    log::info "Activating account $clientEmail"
    gcloud config set account "${clientEmail}" || exit 1
}

# gcp::reserve_ip_address requests a new IP address from GCP and prints this value to STDOUT
#
# Arguments:
#
# required:
# n - name of the IP address to reserve
# p - GCP project name
#
# optional:
# r - GCP compute region, default europe-west4
#
# Returns:
# gcp_reserve_ip_address_return_ip_address - reserved ip address
function gcp::reserve_ip_address {

    local OPTIND
    local ipAddressName
    local gcpProjectName
    local computeRegion="europe-west4"
    local ipAddress

    while getopts ":n:p:r:" opt; do
        case $opt in
            n)
                ipAddressName="$OPTARG" ;;
            p)
                gcpProjectName="$OPTARG" ;;
            r)
                computeRegion=${OPTARG:-$computeRegion} ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$ipAddressName" "IP address name is empty. Exiting..."
    utils::check_empty_arg "$gcpProjectName" "gcp project name is empty. Exiting..."

    log::info "Replacing underscore with dashes in address name."
    ipAddressName=$(echo "$ipAddressName" | tr '_' '-')

    log::info "Reserve IP Address for $ipAddressName"
    local counter=0
    # Check if IP address reservation is present. Wait and retry for one minute to disappear.
    # If IP reservation was removed just before it need a few seconds to disappear.
    # Otherwise, creation will fail.
    ipAddress=$(gcloud compute addresses list \
        --filter="name=$ipAddressName" \
        --format="value(ADDRESS)")
    until [[ -z $ipAddress ]]; do
        sleep 15
        counter=$(( counter + 1 ))
        ipAddress=$(gcloud compute addresses list \
            --filter="name=$ipAddressName" \
            --format="value(ADDRESS)")
        if (( counter == 5 )); then
            # Fail after one minute wait.
            echo "$ipAddressName IP address is still present after one minute wait. Failing"
            return 1
        fi
    done

    gcloud compute addresses create "$ipAddressName" \
        --project="$gcpProjectName" \
        --region="$computeRegion" \
        --network-tier="PREMIUM"
    # shellcheck disable=SC2034
    gcp_reserve_ip_address_return_ip_address="$(gcloud compute addresses list \
        --filter="name=$ipAddressName" \
        --format="value(ADDRESS)")"
    log::info "Created IP Address for Ingressgateway: $ipAddressName"
}

# gcp::create_dns_record creates an A dns record for corresponding ip address
#
# Arguments:
# required:
# a - ip address to use for creating dns record
# h - hostname to use for creating dns record
# s - subdomain to use for creating dns record
# p - GCP project name
# z - GCP dns zone name
#
# Returns:
# gcp_create_dns_record_return_dns_domain - dns domain
function gcp::create_dns_record {

    local OPTIND
    local ipAddress
    local dnsSubDomain
    local dnsHostname
    local gcpProjectName
    local gcpDnsZoneName

    while getopts ":a:h:s:p:z:" opt; do
        case $opt in
            a)
                ipAddress="$OPTARG" ;;
            h)
                dnsHostname="$OPTARG" ;;
            s)
                dnsSubDomain="$OPTARG" ;;
            p)
                gcpProjectName="$OPTARG" ;;
            z)
                gcpDnsZoneName="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2; exit 1 ;;
        esac
    done


    utils::check_empty_arg "$gcpProjectName" "GCP project name is empty. Exiting..."
    utils::check_empty_arg "$gcpDnsZoneName" "GCP DNS zone name is empty. Exiting..."
    utils::check_empty_arg "$ipAddress" "IP address is empty. Exiting..."
    utils::check_empty_arg "$dnsHostname" "DNS hostname is empty. Exiting..."
    utils::check_empty_arg "$dnsSubDomain" "DNS subdomain is empty. Exiting..."

    dnsDomain="$(gcloud dns managed-zones describe "$gcpDnsZoneName" --format="value(dnsName)")"
    # set return value
    # shellcheck disable=SC2034
    gcp_create_dns_record_return_dns_domain=$dnsDomain
    # set return value
    # shellcheck disable=SC2034
    gcp_create_dns_record_return_dns_subdomain=$dnsSubDomain

    dnsFQDN="$dnsHostname.$dnsSubDomain.$dnsDomain"

    set +e
    local attempts=10
    local retryTimeInSec="5"
    for ((i=1; i<=attempts; i++)); do
      gcloud dns --project="$gcpProjectName" record-sets transaction start --zone="$gcpDnsZoneName" && \
      gcloud dns --project="$gcpProjectName" record-sets transaction add "${ipAddress}" --name="${dnsFQDN}" --ttl=60 --type=A --zone="$gcpDnsZoneName" && \
      if gcloud dns --project="$gcpProjectName" record-sets transaction execute --zone="$gcpDnsZoneName"; then
          break
      fi

      gcloud dns record-sets transaction abort --zone="$gcpDnsZoneName" --verbosity none

      if [[ "$i" -lt "$attempts" ]]; then
          echo "Unable to create DNS record, let's wait $retryTimeInSec seconds and retry. Attempts $i of $attempts."
      else
          echo "Unable to create DNS record after $attempts attempts, giving up."
          exit 1
      fi

      sleep $retryTimeInSec
    done

    set -e

    local SECONDS=0
    local endTime=$((SECONDS+600)) #600 seconds == 10 minutes

    while [ $SECONDS -lt $endTime ];do
      echo "Trying to resolve $dnsFQDN"
      sleep 10

      local resolvedIpAddress
      resolvedIpAddress=$(dig +short "$dnsFQDN")

      if [ "$resolvedIpAddress" = "$ipAddress" ]; then
          echo "Successfully resolved $dnsFQDN to $resolvedIpAddress!"
          return 0
      fi
    done

    echo "Cannot resolve $dnsFQDN to expected IP_ADDRESS: $ipAddress."
    log::warn "Continuing anyway... Kyma installation may fail!"
}


# gcp::delete_dns_record will delete dns record for given hostname and IP address
#
# Arguments:
#
# required:
# a - ip address of dns record to remove
# p - GCP project name
# h - DNS hostname to remove
# s - DNS subdomain name of dns record to remove
# z - GCP DNS zone name of dns record to remove
function gcp::delete_dns_record {

    local OPTIND
    local ipAddress
    local dnsSubDomain
    local dnsHostname
    local gcpProjectName
    local gcpDnsZoneName

    while getopts ":a:p:h:s:z:" opt; do
        case $opt in
            a)
                ipAddress="$OPTARG" ;;
            p)
                gcpProjectName="$OPTARG" ;;
            h)
                dnsHostname="$OPTARG" ;;
            s)
                dnsSubDomain="$OPTARG" ;;
            z)
                gcpDnsZoneName="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2; ;;
        esac
    done

    utils::check_empty_arg "$ipAddress" "IP address not provided"
    utils::check_empty_arg "$gcpProjectName" "Project name not provided"
    utils::check_empty_arg "$gcpDnsZoneName" "GCP DNS zone name is empty. Exiting..."
    utils::check_empty_arg "$dnsHostname" "DNS hostname is empty. Exiting..."
    utils::check_empty_arg "$dnsSubDomain" "DNS subdomain is empty. Exiting..."

    local dnsDomain
    dnsDomain="$(gcloud dns managed-zones describe "$gcpDnsZoneName" --format="value(dnsName)")"
    local dnsFQDN="$dnsHostname.$dnsSubDomain.$dnsDomain"

    log::info "Deleting DNS record $dnsFQDN"
    set +e

    local attempts=10
    local retryTimeInSec="5"
    for ((i=1; i<=attempts; i++)); do
        gcloud dns --project="$gcpProjectName" record-sets transaction start --zone="$gcpDnsZoneName" && \
        gcloud dns --project="$gcpProjectName" record-sets transaction remove "$ipAddress" --name="$dnsFQDN" --ttl=60 --type=A --zone="$gcpDnsZoneName" && \
        if gcloud dns --project="$gcpProjectName" record-sets transaction execute --zone="$gcpDnsZoneName"; then
            break
        fi

        gcloud dns record-sets transaction abort --zone="$gcpDnsZoneName" --verbosity none

        if [[ "$i" -lt "$attempts" ]]; then
            echo "Unable to delete DNS record, Retrying after $retryTimeInSec. Attempts $i of $attempts."
        else
            echo "Unable to delete DNS record after $attempts attempts, giving up."
            exit 1
        fi
        sleep $retryTimeInSec
    done

    log::info "DNS Record deleted, but it can be visible for some time due to DNS caches"
    set -e
}

# gcp::set_vars_for_network will generate network and subnetwork names
#
# Arguments:
#
# required:
# n - prowjob name, you can use environment variable $JOB_NAME set by Prow
#
# Return:
# gcp_set_vars_for_network_return_net_name - generated network name
# gcp_set_vars_for_network_return_subnet_name - generated subnetwork name
function gcp::set_vars_for_network {

    local OPTIND
    local jobName
    local networkName
    local subnetworkName

    while getopts ":n:" opt; do
        case $opt in
            n)
                jobName="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2; exit 1 ;;
        esac
    done

    utils::check_empty_arg "$jobName" "Job name is empty. Exiting..."

    if [[ "$jobName" =~ .*_test_of_prowjob_.* ]]; then
        jobName="pjtester"
    fi

    log::info "Replacing underscore with dashes in job name."
    jobName=$(echo "$jobName" | tr '_' '-')
    # Trim jobName to 54 chars to meet network and subnetwork name lenght requirement..
    if [ ${#jobName} -gt 54 ]; then
        jobName=${jobName:(-54):54}
    fi
    # Remove leading dash or dot from network and subnetwork names.
    jobName=${jobName#-}
    jobName=${jobName#.}
    # Add network and subnetwork suffix to prowjob name.
    networkName="$jobName-net"
    subnetworkName="$jobName-subnet"

    # variable hnew return value for calling process
    # shellcheck disable=SC2034
    gcp_set_vars_for_network_return_net_name="$networkName"
    # variable hnew return value for calling process
    # shellcheck disable=SC2034
    gcp_set_vars_for_network_return_subnet_name="$subnetworkName"
}

# gcp::create_network will create GCP network
#
# Arguments:
#
# required:
# n - GCP network name
# p - GCP project name
# s - GCP subnetowrk name
function gcp::create_network {

    local OPTIND
    local gcpProjectName
    local gcpNetworkName
    local gcpSubnetName

    while getopts ":n:p:s:" opt; do
        case $opt in
            n)
                gcpNetworkName="$OPTARG" ;;
            p)
                gcpProjectName="$OPTARG" ;;
            s)
                gcpSubnetName="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$gcpProjectName" "gcp project name is empty. Exiting..."
    utils::check_empty_arg "$gcpNetworkName" "gcp network name is empty. Exiting..."
    utils::check_empty_arg "$gcpSubnetName" "gcp subnet name is empty. Exiting..."

    log::info "Create $gcpNetworkName network with $gcpSubnetName subnet"

    if gcloud compute networks describe "$gcpNetworkName"; then
        log::warn "Network $gcpNetworkName already exists! Skipping network creation."
        return 0
    fi
    log::info "Creating network $gcpNetworkName"
    gcloud compute networks create "$gcpNetworkName" \
        --project="$gcpProjectName" \
        --subnet-mode=custom

    gcloud compute networks subnets create "$gcpSubnetName" \
        --network="$gcpNetworkName" \
        --range=10.0.0.0/22

   log::info "Successfully created network $gcpNetworkName"
}

# gcp::deprovision_k8s_cluster removes a GKE cluster
#
# Arguments:
#
# required:
# n - k8s cluster name to remove
# p - GCP project name
#
# optional:
# z - GCP compute zone, default europe-west4-b
# R - GCP compute region, default europe-west4
# r - if true clean regional cluster, default false
# d - if true clean cluster in async mode, default true
function gcp::deprovision_k8s_cluster {

    local OPTIND
    local clusterName
    local projectName
    local computeZone="europe-west4-b"
    local computeRegion="europe-west4"
    local cleanRegionalCluster="false"
    local asyncDeprovision="true"
    local params

    while getopts ":n:p:z:R:r:d:" opt; do
        case $opt in
            n)
                clusterName="$OPTARG" ;;
            p)
                projectName="$OPTARG" ;;
            z)
                computeZone=${OPTARG:-$computeZone} ;;
            R)
                computeRegion=${OPTARG:-$computeRegion} ;;
            r)
                cleanRegionalCluster=${OPTARG:-$cleanRegionalCluster} ;;
            d)
                asyncDeprovision=${OPTARG:-$asyncDeprovision} ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2; ;;
        esac
    done

    utils::check_empty_arg "$clusterName" "Cluster name not provided"
    utils::check_empty_arg "$projectName" "Project name not provided"

    log::info "Deprovisioning cluster $clusterName."

    params+=("--quiet")
    if [ "$cleanRegionalCluster" = "true" ]; then
        #Pass gke region name to delete command.
        params+=("--region=$computeRegion")
    else
        params+=("--zone=$computeZone")
    fi

    if [ "$asyncDeprovision" = "true" ]; then
        params+=("--async")
    fi

    if gcloud --project="$projectName" beta container clusters delete "$clusterName" "${params[@]}"; then
        log::info "Successfully removed cluster $clusterName!"
    fi
}

# gcp::delete_ip_address will delete ip address identified by it's name in GCP.
#
# Arguments:
#
# required:
# p - GCP project name
# n - IP address name
#
# optional:
# R - GCP compute region where IP address exists, default europe-west4
function gcp::delete_ip_address {

    local OPTIND
    local gcpProjectName
    local ipAddressName
    local gcpComputeRegion="europe-west4" # R - region in which the new regional cluster will be created

    while getopts ":p:R:n:" opt; do
        case $opt in
            p)
                gcpProjectName="$OPTARG" ;;
            n)
                ipAddressName="$OPTARG" ;;
            R)
                gcpComputeRegion=${OPTARG:-$computeRegion} ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2; ;;
        esac
    done

    utils::check_empty_arg "$ipAddressName" "IP address not provided"
    utils::check_empty_arg "$gcpProjectName" "Project name not provided"

    log::info "Removing IP address $ipAddressName."
    if gcloud compute addresses delete "$ipAddressName" --project="$gcpProjectName" --region="$gcpComputeRegion"; then
        log::info "Successfully removed IP $ipAddressName!"
    else
        log::error "Failed to remove IP $ipAddressName!"
        return 1
    fi
}


# gcp::delete_docker_image deletes Docker image
# Arguments:
# required:
# i - name of the Docker image
function gcp::delete_docker_image() {

    local OPTIND
    local imageName

    while getopts ":i:" opt; do
        case $opt in
            i)
                imageName="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2; exit 1 ;;
        esac
    done

    utils::check_empty_arg "$imageName" "Image name is empty. Exiting..."
    gcloud container images delete "$imageName" || \
    (
        log::error "Could not remove Docker image" && \
        exit 1
    )
}


# gcp::set_latest_cluster_version_for_channel checks for latest possible version in GKE_RELEASE_CHANNEL and updates GKE_CLUSTER_VERSION accordingly
# Arguments:
# required:
# C - release channel
# Returns
# gcp_set_latest_cluster_version_for_channel_return_cluster_version - latest cluster version for given channel
function gcp::set_latest_cluster_version_for_channel() {

    local OPTIND
    local releaseChannel
    local clusterVersion

    while getopts ":C:" opt; do
        case $opt in
            C)
                releaseChannel="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2; exit 1 ;;
        esac
    done

    utils::check_empty_arg "$releaseChannel" "Release channel is empty. Exiting..."

    clusterVersion=$(gcloud container get-server-config --zone europe-west4 --format="json" | jq -r '.channels|.[]|select(.channel | contains("'"${releaseChannel}"'"|ascii_upcase))|.validVersions|.[0]')
    log::info "Updating GKE_CLUSTER_VERSION to newest available in ${releaseChannel}: ${clusterVersion}"

    # shellcheck disable=SC2034
    gcp_set_latest_cluster_version_for_channel_return_cluster_version="$clusterVersion"
}

# gcp::encrypt encrypts text using Google KMS
#
# Arguments:
# required:
# t - plain text file to encrypt
# c - cipher text file
# e - encryption key
# k - keyring
# p - project
function gcp::encrypt {
    local OPTIND
    local plainText
    local cipherText
    local encryptionKey
    local keyring
    local project

    while getopts ":t:c:e:k:p:" opt; do
        case $opt in
            t)
                plainText="$OPTARG" ;;
            c)
                cipherText="$OPTARG" ;;
            e)
                encryptionKey=${OPTARG} ;;
            k)
                keyring="$OPTARG" ;;
            p)
                project=${OPTARG} ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2; ;;
        esac
    done

    utils::check_empty_arg "$plainText" "Plain text not provided"
    utils::check_empty_arg "$cipherText" "Cipher text not provided"
    utils::check_empty_arg "$encryptionKey" "Encryption key not provided"
    utils::check_empty_arg "$keyring" "keyring name not provided"
    utils::check_empty_arg "$project" "Project name not provided"



  log::info "Encrypting ${plainText} as ${cipherText}"
  gcloud kms encrypt --location global \
      --keyring "${keyring}" \
      --key "${encryptionKey}" \
      --plaintext-file "${plainText}" \
      --ciphertext-file "${cipherText}" \
      --project "${project}"
}

# gcp::encrypt encrypts text using Google KMS
#
# Arguments:
# required:
# t - encrypted text file to decrypt
# c - cipher text file
# e - encryption key
# k - keyring
# p - project
function gcp::decrypt {
  local OPTIND
    local plainText
    local cipherText
    local encryptionKey
    local keyring
    local project

    while getopts ":t:c:e:k:p:" opt; do
        case $opt in
            t)
                plainText="$OPTARG" ;;
            c)
                cipherText="$OPTARG" ;;
            e)
                encryptionKey=${OPTARG} ;;
            k)
                keyring="$OPTARG" ;;
            p)
                project=${OPTARG} ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2; ;;
        esac
    done

    utils::check_empty_arg "$plainText" "Plain text file not provided"
    utils::check_empty_arg "$cipherText" "Cipher text file not provided"
    utils::check_empty_arg "$encryptionKey" "Encryption key not provided"
    utils::check_empty_arg "$keyring" "keyring name not provided"
    utils::check_empty_arg "$project" "Project name not provided"


  log::info "Decrypting ${cipherText} to ${plainText}"

  gcloud kms decrypt --location global \
      --keyring "${keyring}" \
      --key "${encryptionKey}" \
      --ciphertext-file "${cipherText}" \
      --plaintext-file "${plainText}" \
      --project "${project}"
}

# gcp::get_cluster_kubeconfig gets kubeconfig for the chosen cluster
#
# Arguments:
#
# Required arguments:
# c - GKE cluster name
# p - GCP project name
#
# Optional arguments:
# z - zone in which the cluster is located
# R - region in which the cluster is located
# r - it true it is regional cluster
function gcp::get_cluster_kubeconfig {
    local OPTIND
    # required arguments
    local clusterName
    local gcpProjectName

    # default values
    local computeZone="europe-west4-b"
    local computeRegion="europe-west4"
    local provisionRegionalCluster="false"

    while getopts ":c:p:z:R:r:" opt; do
        case $opt in
            c)
                clusterName="${OPTARG:0:40}" ;;
            p)
                gcpProjectName="$OPTARG" ;;
            z)
                computeZone=${OPTARG:-$computeZone} ;;
            R)
                computeRegion=${OPTARG:-$computeRegion} ;;
            r)
                provisionRegionalCluster=${OPTARG:-$provisionRegionalCluster} ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2; ;;
        esac
    done
    
    utils::check_empty_arg "$clusterName" "Cluster name not provided."
    utils::check_empty_arg "$gcpProjectName" "GCP project name not provided."


    log::info "Getting kubeconfig for cluster $clusterName"
    local params

    if [ "$provisionRegionalCluster" = "true" ] ; then
        params+=("--region=$computeRegion")
    else
        params+=("--zone=$computeZone")
    fi

    params+=("--project=$gcpProjectName")

    gcloud container clusters get-credentials "$clusterName" "${params[@]}"
}
# (2025-03-04)