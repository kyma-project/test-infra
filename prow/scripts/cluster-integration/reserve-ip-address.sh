#!/usr/bin/env bash

set -o errexit

discoverUnsetVar=false

for var in GCLOUD_PROJECT_NAME GCLOUD_REGION GCLOUD_IP_ADDRESS_NAME; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done

if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

gcloud beta compute --project="${GCLOUD_PROJECT_NAME}" addresses create "${GCLOUD_IP_ADDRESS_NAME}" --region="${GCLOUD_REGION}" --network-tier=PREMIUM

gcloud compute addresses list --filter="name=${GCLOUD_IP_ADDRESS_NAME}" --format="value(ADDRESS)"
