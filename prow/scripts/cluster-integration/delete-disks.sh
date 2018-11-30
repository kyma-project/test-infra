#!/usr/bin/env bash

#Description: Removes orphaned disks from gcloud
# Note: This script is executed AFTER cluster is deleted.
#
# The purpose of the script is to clean up disks that are provisioned in gcloud for PVCs used by Kyma.
# These disks are NOT removed along with the cluster and are left orphaned after cluster is deleted, causing resource leak.
# We can get partial names of these objects from the cluster itself, while it's still running, like this:
# DISKS=$(kubectl get pvc --all-namespaces -o jsonpath="{.items[*].spec.volumeName}" | xargs -n1 echo)
# This allows to get a gcloud disk name suffix (not entire name), which is enough to find the proper object and delete it.
#
#Expected vars:
# - DISKS: names of disks to delete
#
#Permissions: In order to run this script you need to use a service account with "Compute Admin" role

set +e

for NAMEPATTERN in ${DISKS}
do
    DISK_NAME=$(gcloud compute disks list --filter="name~${NAMEPATTERN}" --format="value(name)")
    echo "Removing disk: ${DISK_NAME}"
    gcloud compute disks delete "${DISK_NAME}" --quiet
done

