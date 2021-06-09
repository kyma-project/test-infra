#!/usr/bin/env bash
LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${LIBDIR}/utils.sh"

# gcloud::provision_gke_cluster creates a GKE cluster
# For switch parameters look up the code below.
#
# Required exported variables:
# GCLOUD_COMPUTE_ZONE - zone in which the new cluster will be created
# GCLOUD_PROJECT_NAME - name of GCP project
# GKE_CLUSTER_VERSION - GKE cluster version
#
# Arguments:
# Required arguments:
# $1 - GKE cluster name
# $2 - GCP project name
# $3 - GKE cluster version
# $ - path to test-infra sources
#
# Optional arguments:
# $3
# $2 - optional additional labels for the cluster
function gcp::provision_gke_cluster {
    # check required arguments
    utils::check_empty_arg "$1" "Cluster name not provided."
    utils::check_empty_arg "$2" "GCP project name not provided."
    utils::check_empty_arg "$3" "GKE cluster version not provided."
    utils::check_empty_arg "$" "Path to test-infra repo sources not provided."


    # name arguments
    local clusterName="$1"
    local gcpProjectName="$2"
    local gkeClusterVersion="$3"
    local testInfraSourcesDir="$"
    local additionalLabels=$

    # default values
    local currentTimestampReadableParam
    local currentTimestampParam
    readonly currentTimestampReadableParam=$(date +%Y%m%d)
    readonly currentTimestampParam=$(date +%s)
    local ttlHoursParam="3"
    local machineTypeParam="n1-standard-4"
    local numNodesParam="3"
    local nodesPerZoneParam="1"
    local networkParam="default"
    local computeZoneParam="europe-west4-b"
    local computeRegionParam="europe-west4"
    local kubeDnsPatchPath="$testInfraSourcesDir/prow/scripts/resources/kube-dns-stub-domains-patch.yaml"
    local params

    # mandatory labels
    local cleanerLabels="created-at=$currentTimestampParam,created-at-readable=$currentTimestampReadableParam,ttl=${TTL_HOURS:-$ttlHoursParam}"
    local jobLabels="job=$JOB_NAME,job-id=$PROW_JOB_ID"
    local clusterLabels="cluster=$clusterName,volatile=true"

    # optional labels
    if [ "${additionalLabels}" ]; then
        additionalLabels=",${additionalLabels}"
    fi

    # build lables parameter value
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
    params+=("--machine-type=${MACHINE_TYPE:-$machineTypeParam}")
    if [ "${PROVISION_REGIONAL_CLUSTER}" ] ; then
        params+=("--region=${CLOUDSDK_COMPUTE_REGION:-$computeRegionParam}")
        params+=("--num-nodes=${NODES_PER_ZONE:-$nodesPerZoneParam}")
    else
        params+=("--zone=${GCLOUD_COMPUTE_ZONE:-$computeZoneParam}")
        params+=("--num-nodes=${NUM_NODES:-$numNodesParam}")
    fi
    if [ "${GCLOUD_NETWORK_NAME}" ] && [ "${GCLOUD_SUBNET_NAME}" ]; then
        params+=("--network=${GCLOUD_NETWORK_NAME}")
        params+=("--subnetwork=${GCLOUD_SUBNET_NAME}")
    else
        params+=("--network=${GCLOUD_NETWORK_NAME:-$networkParam}")
    fi

    # Optional parameters
    if [ "${GKE_RELEASE_CHANNEL}" ]; then
        params+=("--release-channel=${GKE_RELEASE_CHANNEL}")
    fi
    # serverless tests are failing when are running on a cluster with contianerD
    if [[ "${GKE_RELEASE_CHANNEL}" == "rapid" ]]; then
        # set image type to the image that uses docker instead of containerD
        params+=("--image-type=cos")
    elif [ "${IMAGE_TYPE}" ]; then
        params+=("--image-type=${IMAGE_TYPE}")
    fi
    if [ "${STACKDRIVER_KUBERNETES}" ]; then
        params+=("--enable-stackdriver-kubernetes")
    fi
    if [ "${CLUSTER_USE_SSD}" ]; then
        params+=("--disk-type=pd-ssd")
    fi
    if [ "${GCLOUD_SECURITY_GROUP_DOMAIN}" ]; then
        params+=("--security-group=gke-security-groups@${GCLOUD_SECURITY_GROUP_DOMAIN}")
    fi
    if [ "${GKE_ENABLE_POD_SECURITY_POLICY}" ]; then
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
