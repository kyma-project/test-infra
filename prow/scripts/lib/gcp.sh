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
# z - zone in which the new zonal cluster will be created, default europe-west4-b
# m - machine type to use in a cluster, default n1-standard-4
# R - region in which the new regional cluster will be created, default europe-west4
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

    # default values
    local clusterName
    local gcpProjectName
    local gkeClusterVersion
    local additionalLabels
    local prowjobName
    local prowjobID
    local ttlHours="3"
    local computeZone="europe-west4-b"
    local machineType="n1-standard-4"
    local computeRegion="europe-west4"
    local numNodes="3"
    local nodesPerZone="1"
    local networkNameDefault="default"
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
                clusterName="$OPTARG" ;;
            p)
                gcpProjectName="$OPTARG" ;;
            v)
                gkeClusterVersion="$OPTARG" ;;
            l)
                additionalLabels="$OPTARG" ;;
            j)
                prowjobName="$OPTARG";;
            J)
                prowjobID="$OPTARG";;
            t)
                ttlHours="$OPTARG" ;;
            z)
                computeZone="$OPTARG" ;;
            m)
                machineType="$OPTARG" ;;
            R)
                computeRegion="$OPTARG" ;;
            n)
                local nodesCount="$OPTARG" ;;
            N)
                local networkName="$OPTARG" ;;
            S)
                local subnetName="$OPTARG" ;;
            C)
                local gkeReleaseChannel="$OPTARG" ;;
            i)
                local imageType="$OPTARG" ;;
            g)
                local gcpSecurityGroupDomain="$OPTARG" ;;
            r)
                provisionRegionalCluster="$OPTARG" ;;
            s)
                enableStackdriver="$OPTARG" ;;
            D)
                enableSSD="$OPTARG" ;;
            e)
                enablePSP="$OPTARG" ;;
            P)
                testInfraSourcesDir="$OPTARG" ;;
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

    local kubeDnsPatchPath="$testInfraSourcesDir/prow/scripts/resources/kube-dns-stub-domains-patch.yaml"
    local params

    # mandatory labels
    local cleanerLabels="created-at=$currentTimestampParam,created-at-readable=$currentTimestampReadableParam,ttl=$ttlHours"
    local jobLabels="job=$prowjobName,job-id=$prowjobID"
    local clusterLabels="cluster=$clusterName,volatile=true"

    # optional labels
    if [ "$additionalLabels" ]; then
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
        params+=("--network=${networkName:-$networkNameDefault}")
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
function gcp::authenticate() {
    log::info "Authenticating to gcloud"
    if [[ -z "$1" ]]; then
      log::error "Missing account credentials, please provide proper credentials"
    fi
    gcloud auth activate-service-account --key-file "${1}" || exit 1
}
