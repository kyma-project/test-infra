#!/usr/bin/env bash

################################################################################
# Removes disks from gcloud
################################################################################

set +e
echo "Removing remaining PVC disks"

for DISK in ${DISKS}
do
    echo "TODO: REMOVE, DEBUG: ${DISK}"
    DISK_NAME=$(gcloud compute disks list --filter="name~${DISK}" --format="value(name)")
    echo "Removing disk: ${DISK_NAME}"
    gcloud compute disks delete ${DISK_NAME} --quiet
done

