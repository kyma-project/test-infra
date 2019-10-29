#!/bin/bash

###
# Following script provisions GKE cluster.
#
# INPUTS:
# - GCLOUD_SERVICE_KEY_PATH - content of service account credentials json file
# - GCLOUD_PROJECT_NAME - name of GCP project
# - CLUSTER_NAME - name for the new cluster
# - GCLOUD_COMPUTE_ZONE - zone in which the new cluster will be provisioned
#
# OPTIONAL:
# - CLUSTER_VERSION - the k8s version to use for the master and nodes
# - MACHINE_TYPE - the type of machine to use for nodes
# - NUM_NODES - the number of nodes to be created
# - ADDITIONAL_LABELS - labels applied on the cluster
# - GCLOUD_NETWORK_NAME - network name for the cluster, must be specified together with $GCLOUD_NETWORK_NAME
# - GCLOUD_SUBNET_NAME - subnet name from $GCLOUD_NETWORK_NAME, must be specified together with $GCLOUD_NETWORK_NAME 
#
# REQUIREMENTS:
# - gcloud
###

set -o errexit

discoverUnsetVar=false

for var in GCLOUD_SERVICE_KEY_PATH GCLOUD_PROJECT_NAME CLUSTER_NAME GCLOUD_COMPUTE_ZONE; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

readonly CURRENT_TIMESTAMP_READABLE_PARAM=$(date +%Y%m%d)
readonly CURRENT_TIMESTAMP_PARAM=$(date +%s)

TTL_HOURS_PARAM="3"
CLUSTER_VERSION_PARAM="--cluster-version=1.13"
MACHINE_TYPE_PARAM="--machine-type=n1-standard-4"
NUM_NODES_PARAM="--num-nodes=3"
NETWORK_PARAM=(--network=default)
if [ "${TTL_HOURS}" ]; then TTL_HOURS_PARAM="${TTL_HOURS}"; fi
CLEANER_LABELS_PARAM="created-at=${CURRENT_TIMESTAMP_PARAM},created-at-readable=${CURRENT_TIMESTAMP_READABLE_PARAM},ttl=${TTL_HOURS_PARAM}"
STACKDRIVER_KUBERNETES_PARAM=""
DISK_TYPE_PARAM=""


if [ "${CLUSTER_VERSION}" ]; then CLUSTER_VERSION_PARAM="--cluster-version=${CLUSTER_VERSION}"; fi
if [ "${MACHINE_TYPE}" ]; then MACHINE_TYPE_PARAM="--machine-type=${MACHINE_TYPE}"; fi
if [ "${NUM_NODES}" ]; then NUM_NODES_PARAM="--num-nodes=${NUM_NODES}"; fi
if [ "${GCLOUD_NETWORK_NAME}" ] && [ "${GCLOUD_SUBNET_NAME}" ]; then NETWORK_PARAM=(--network="${GCLOUD_NETWORK_NAME}" --subnetwork="${GCLOUD_SUBNET_NAME}"); fi
if [ "${STACKDRIVER_KUBERNETES}" ];then STACKDRIVER_KUBERNETES_PARAM="--enable-stackdriver-kubernetes"; fi
if [ "${CLUSTER_USE_SSD}" ];then DISK_TYPE_PARAM="--disk-type=pd-ssd"; fi

APPENDED_LABELS=""
if [ "${ADDITIONAL_LABELS}" ]; then APPENDED_LABELS=(",${ADDITIONAL_LABELS}") ; fi
LABELS_PARAM=(--labels="job=${JOB_NAME},job-id=${PROW_JOB_ID},cluster=${CLUSTER_NAME},volatile=true${APPENDED_LABELS[@]},${CLEANER_LABELS_PARAM}")

command -v gcloud

gcloud auth activate-service-account --key-file="${GCLOUD_SERVICE_KEY_PATH}"
gcloud config set project "${GCLOUD_PROJECT_NAME}"
gcloud config set compute/zone "${GCLOUD_COMPUTE_ZONE}"

gcloud beta container clusters create "${CLUSTER_NAME}" "${CLUSTER_VERSION_PARAM}" "${MACHINE_TYPE_PARAM}" "${NUM_NODES_PARAM}" "${NETWORK_PARAM[@]}" "${LABELS_PARAM[@]}" "${STACKDRIVER_KUBERNETES_PARAM}" "${DISK_TYPE_PARAM}"
