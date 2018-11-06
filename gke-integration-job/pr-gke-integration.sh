#!/usr/bin/env bash

set -o errexit

###
# Following script provisions GKE cluster and deprovision.
#
# INPUTS:
# - GCLOUD_SERVICE_KEY_PATH - content of service account credentials json file
# - GCLOUD_PROJECT_NAME - name of GCP project
# - GCLOUD_COMPUTE_ZONE - zone in which the new cluster will be provisioned
#
# REPO_OWNER, REPO_NAME and PULL_NUMBER are set by ProwJob
#
# OPTIONAL:
# - CLUSTER_VERSION - the k8s version to use for the master and nodes
# - MACHINE_TYPE - the type of machine to use for nodes
# - NUM_NODES - the number of nodes to be created
#
###

set -o errexit

ROOT_PATH="/home/prow/go/src/github.com/kyma-project/kyma"

discoverUnsetVar=false

for var in GCLOUD_SERVICE_KEY_PATH GCLOUD_PROJECT_NAME GCLOUD_COMPUTE_ZONE; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

echo "
################################################################################
# Provisioning gke cluster
################################################################################
"
CLUSTER_NAME="${REPO_OWNER}-${REPO_NAME}-${PULL_NUMBER}" ${ROOT_PATH}/prow/scripts/provision-gke-cluster.sh
