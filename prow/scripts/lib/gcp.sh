#!/usr/bin/env bash
LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${LIBDIR}/utils.sh"

# gcloud::provision_gke_cluster creates a GKE cluster
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
function gcp::provision_gke_cluster {

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
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done


    # check required arguments
    utils::check_empty_arg "$clusterName" "Cluster name not provided."
    utils::check_empty_arg "$gcpProjectName" "GCP project name not provided."
    utils::check_empty_arg "$gkeClusterVersion" "GKE cluster version not provided."
    utils::check_empty_arg "$prowjobName" "prowjob name not provided."
    utils::check_empty_arg "$prowjobID" "prowjob ID not provided."

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
            hexit 1
        fi
        echo -e "Waiting for kube-dns to be available. Try $(( counter + 1 )) of 5"
        counter=$(( counter + 1 ))
        sleep 15
    done

    kubectl -n kube-system patch cm kube-dns --type merge --patch "$(cat "$kubeDnsPatchPath")"

    # run oom debug pod
    utils::debug_oom
}


# gcloud::authenticate authenticates to gcloud.
# Arguments:
# $1 - google login credentials
function gcp::authenticate {

    local OPTIND
    #required arguments
    local googleAppCredentials

    while getopts ":c:" opt; do
        case $opt in
            c)
                googleAppCredentials="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$googleAppCredentials" "Missing account credentials, please provide proper credentials"

    log::info "Authenticating to gcloud"
    gcloud auth activate-service-account --key-file "$googleAppCredentials" || exit 1
}

# gcloud::reserve_ip_address requests a new IP address from gcloud and prints this value to STDOUT
# Required exported variables:
# CLOUDSDK_CORE_PROJECT - gcp project
# CLOUDSDK_COMPUTE_REGION - gcp region
# Arguments:
# $1 - name of the IP address to be set in gcp COMMON_NAME
# Returns:
# gcloud::reserve_ip_address_return_1 - reserved ip address -> GATEWAY_IP_ADDRESS
# TODO: add support for setting CLOUDSDK env vars from function args.
function gcp::reserve_ip_address {

    local OPTIND
    # required arguments
    local ipAddressName
    local gcpProjectName
    local computeRegion
    local ipAddress

    while getopts ":n:p:r:" opt; do
        case $opt in
            n)
                ipAddressName="$OPTARG" ;;
            p)
                gcpProjectName="$OPTARG" ;;
            r)
                computeRegion="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$ipAddressName" "IP address name is empty. Exiting..."
    utils::check_empty_arg "$gcpProjectName" "gcp project name is empty. Exiting..."
    utils::check_empty_arg "$computeRegion" "gcp compute region name is empty. Exiting..."

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

# gcloud::create_dns_record creates an A dns record for corresponding ip address
# Required exported variables:
# CLOUDSDK_CORE_PROJECT - gcp project
# CLOUDSDK_COMPUTE_REGION - gcp region
#
# Arguments:
# $1 - ip address
# $2 - domain name
function gcp::create_dns_record {

    local OPTIND
    local ipAddress
    local dnsSubDomain
    local dnsDomain
    local dnsHostname
    local gcpProjectName
    local gcpDnsZoneName

    while getopts ":a:h:s:d:p:z:" opt; do
        case $opt in
            a)
                ipAddress="$OPTARG" ;;
            h)
                dnsHostname="$OPTARG" ;;
            s)
                dnsSubDomain="$OPTARG" ;;
            d)
                dnsDomain="$OPTARG" ;;
            p)
                gcpProjectName="$OPTARG" ;;
            z)
                gcpDnsZoneName="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2; exit 1 ;;
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
    gcp_create_dns_record_dns_domain=$dnsDomain

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


# gcloud::delete_dns_record
# Required exported variables:
# CLOUDSDK_CORE_PROJECT - gcp project
# CLOUDSDK_COMPUTE_REGION - gcp region
#
# Arguments:
# $1 - ip address
# $2 - domain name
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
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2; ;;
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
    gcloud dns --project="$projectName" record-sets transaction start --zone="$gcpDnsZoneName" && \
    gcloud dns --project="$projectName" record-sets transaction remove "$ipAddress" --name="$dnsFQDN" --ttl=60 --type=A --zone="$gcpDnsZoneName" && \
    if gcloud dns --project="$projectName" record-sets transaction execute --zone="$gcpDnsZoneName"; then
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


# Arguments:
# n - $JOB_NAME
function gcp::set_vars_for_network {

    local OPTIND
    local jobName

    while getopts ":n:" opt; do
        case $opt in
            n)
                jobName="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2; exit 1 ;;
        esac
    done

    utils::check_empty_arg "$jobName" "Job name is empty. Exiting..."

    log::info "Replacing underscore with dashes in job name."
    jobName=$(echo "$jobName" | tr '_' '-')

    # variable hold return value for calling process
    # shellcheck disable=SC2034
    gcp_set_vars_for_network_net_name="$jobName-net"
    # variable hold return value for calling process
    # shellcheck disable=SC2034
    gcp_set_vars_for_network_subnet_name="$jobName-subnet"
}

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
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
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

# gcloud::deprovision_gke_cluster removes a GKE cluster
# Required exported variables:
# GCLOUD_COMPUTE_ZONE - zone in which the new cluster will be removed
# GCLOUD_PROJECT_NAME - name of GCP project
#
# Arguments:
# $1 - cluster name
function gcp::deprovision_gke_cluster {

    local OPTIND
    local clusterName
    local projectName
    local computeZone="europe-west4-b" # z - zone in which the new zonal cluster will be created
    local computeRegion="europe-west4" # R - region in which the new regional cluster will be created
    local asyncDeprovision="true" # A - deprovision cluster in async mode
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
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2; ;;
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
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2; ;;
        esac
    done

    utils::check_empty_arg "$ipAddressName" "IP address not provided"
    utils::check_empty_arg "$gcpProjectName" "Project name not provided"

    log::info "Removing IP address $ipAddressName."
    if gcloud compute addresses delete "$ipAddressName" --project="$gcpProjectName" --region="$gcpComputeRegion"; then
        log::info "Successfully removed IP $ipAddressName!"
    fi
}
