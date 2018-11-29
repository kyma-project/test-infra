#!/usr/bin/env bash

#Description: Removes orphaned disks from gcloud
# Note: This script is executed AFTER cluster is deleted.
#
# The purpose of the script is to clean up disks that are provisioned in gcloud for PVCs used by Kyma.
# These disks are NOT removed along with the cluster and are left orphaned after cluster is deleted, causing resource leak.
# We're able to find those disks by label that propagate from the cluster.
#
#Expected vars:
# - CLUSTER_NAME: name of the cluster
#
#Permissions: In order to run this script you need to use a service account with "Compute Admin" role

set +e
echo "Removing remaining PVC disks"
DISKS_NAMES=$(gcloud compute disks list --filter="labels.cluster:${CLUSTER_NAME}" --format="value(name)")
echo "TODO: DEBUG: DISKS: $DISKS_NAMES"
echo

for DISK_NAME in ${DISKS_NAMES}
do
    echo "Removing disk: ${DISK_NAME}"
    gcloud compute disks delete "${DISK_NAME}" --quiet
done

