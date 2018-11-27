#!/bin/bash

###
# Following script deprovisions GKE cluster.
#
# INPUTS:
# - GCLOUD_SERVICE_KEY_PATH - content of service account credentials json file
# - GCLOUD_PROJECT_NAME - name of GCP project
# - CLUSTER_NAME - name for the new cluster
# - GCLOUD_COMPUTE_ZONE - zone in which the new cluster will be deprovisioned
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

command -v gcloud

gcloud auth activate-service-account --key-file="${GCLOUD_SERVICE_KEY_PATH}"
gcloud config set project "${GCLOUD_PROJECT_NAME}"
gcloud config set compute/zone "${GCLOUD_COMPUTE_ZONE}"

#TODO: DEBUG

#gcloud compute instances list --filter="labels.job:kyma-gke-integration AND labels.cluster:${CLUSTER_NAME}"
VM_INSTANCES=$(gcloud compute instances list --filter="labels.job:kyma-gke-integration AND labels.cluster:${CLUSTER_NAME}" --format="value(name)")
echo "VM Instances: ${VM_INSTANCES}"

trap cleanup_vm EXIT
cleanup_vm() {
    sleep 5
		for vm in ${VM_INSTANCES}; do
        gcloud compute disks list --filter="name=$vm"
    done
#gcloud compute disks list --filter="name~'gke-gkeint-kyma-projec-pvc' AND labels.cluster:${CLUSTER_NAME}"
}


gcloud container clusters delete "${CLUSTER_NAME}" --quiet

