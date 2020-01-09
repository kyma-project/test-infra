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

for var in GCLOUD_SERVICE_KEY_PATH GCLOUD_PROJECT_NAME CLUSTER_NAME GCLOUD_COMPUTE_ZONE TEST_INFRA_SOURCES_DIR TEST_INFRA_SOURCES_DIR; do
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
declare -a GCLOUD_PARAMS

TTL_HOURS_PARAM="3"
CLUSTER_VERSION_PARAM="--cluster-version=1.13"
MACHINE_TYPE_PARAM="--machine-type=n1-standard-4"
NUM_NODES_PARAM="--num-nodes=3"
NETWORK_PARAM="--network=default"
if [ "${TTL_HOURS}" ]; then TTL_HOURS_PARAM="${TTL_HOURS}"; fi
CLEANER_LABELS_PARAM="created-at=${CURRENT_TIMESTAMP_PARAM},created-at-readable=${CURRENT_TIMESTAMP_READABLE_PARAM},ttl=${TTL_HOURS_PARAM}"

GCLOUD_PARAMS+=("${CLUSTER_NAME}")
if [ "${CLUSTER_VERSION}" ]; then GCLOUD_PARAMS+=("--cluster-version=${CLUSTER_VERSION}"); else GCLOUD_PARAMS+=("${CLUSTER_VERSION_PARAM}"); fi
if [ "${MACHINE_TYPE}" ]; then GCLOUD_PARAMS+=("--machine-type=${MACHINE_TYPE}"); else GCLOUD_PARAMS+=("${MACHINE_TYPE_PARAM}"); fi
if [ "${NUM_NODES}" ]; then GCLOUD_PARAMS+=("--num-nodes=${NUM_NODES}"); else GCLOUD_PARAMS+=("${NUM_NODES_PARAM}"); fi
if [ "${GCLOUD_NETWORK_NAME}" ] && [ "${GCLOUD_SUBNET_NAME}" ]; then GCLOUD_PARAMS+=("--network=${GCLOUD_NETWORK_NAME}" "--subnetwork=${GCLOUD_SUBNET_NAME}"); else GCLOUD_PARAMS+=("${NETWORK_PARAM}"); fi
if [ "${STACKDRIVER_KUBERNETES}" ];then GCLOUD_PARAMS+=("--enable-stackdriver-kubernetes"); fi
if [ "${CLUSTER_USE_SSD}" ];then GCLOUD_PARAMS+=("--disk-type=pd-ssd"); fi
#Must be added at the end to override --num-nodes parameter value set earlier.
if [ "${PROVISION_REGIONAL_CLUSTER}" ] && [ "${CLOUDSDK_COMPUTE_REGION}" ] && [ "${NODES_PER_ZONE}" ];then GCLOUD_PARAMS+=("--region=${CLOUDSDK_COMPUTE_REGION}" "--num-nodes=${NODES_PER_ZONE}"); fi

APPENDED_LABELS=""
if [ "${ADDITIONAL_LABELS}" ]; then APPENDED_LABELS=(",${ADDITIONAL_LABELS}") ; fi
GCLOUD_PARAMS+=("--labels=job=${JOB_NAME},job-id=${PROW_JOB_ID},cluster=${CLUSTER_NAME},volatile=true${APPENDED_LABELS[@]},${CLEANER_LABELS_PARAM}")

command -v gcloud

gcloud auth activate-service-account --key-file="${GCLOUD_SERVICE_KEY_PATH}"
gcloud config set project "${GCLOUD_PROJECT_NAME}"
gcloud config set compute/zone "${GCLOUD_COMPUTE_ZONE}"

echo -e "\n---> Creating cluster with follwing parameters."
echo "${GCLOUD_PARAMS[@]}"
echo -e "\n---> Creating cluster"
gcloud beta container clusters create "${GCLOUD_PARAMS[@]}"

kubectl -n kube-system patch cm kube-dns --type merge --patch \
  "$(cat "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/resources/kube-dns-stub-domains-patch.yaml)"
