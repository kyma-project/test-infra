#!/usr/bin/env bash

#Description: Removes orphaned disks from gcloud
# Note: This script is executed AFTER cluster is deleted.
#
# The purpose of the script is to clean up disks that are provisioned in gcloud for PVCs used by Kyma.
# These disks are NOT removed along with the cluster and are left orphaned after cluster is deleted, causing resource leak.
# We're able to find those disks by label that propagate from the cluster, until the cluster exists.
#
#Expected vars:
# - DISKS_NAMES: names of disks to delete
#
#Permissions: In order to run this script you need to use a service account with "Compute Admin" role

set +e
echo "TODO: DEBUG: DISKS: $DISKS"
echo

for NAMEPATTERN in ${DISKS}
do
    DISK_NAME=$(gcloud compute disks list --filter="name~${NAMEPATTERN}" --format="value(name)")
    echo "Removing disk: ${DISK_NAME}"
    gcloud compute disks delete "${DISK_NAME}" --quiet
done

