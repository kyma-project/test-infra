#!/usr/bin/env bash

# Description: Creates a GCP network for cluster

# Expected vars:
# - GCLOUD_NETWORK_NAME - name for the new GCP network
# - GCLOUD_SUBNET_NAME - name for the subnet of the network
# - GCLOUD_PROJECT_NAME - name of GCP project

set -o errexit

discoverUnsetVar=false

for var in GCLOUD_NETWORK_NAME GCLOUD_SUBNET_NAME GCLOUD_PROJECT_NAME; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

gcloud compute networks create "${GCLOUD_NETWORK_NAME}" \
 --project="${GCLOUD_PROJECT_NAME}" \
 --subnet-mode=custom

gcloud compute networks subnets create "${GCLOUD_SUBNET_NAME}" \
 --network="${GCLOUD_NETWORK_NAME}" \
 --range=10.0.0.0/24
