#!/usr/bin/env bash

set -o errexit

discoverUnsetVar=false

for var in CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION IP_ADDRESS_NAME; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done

if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

gcloud beta compute --project="${CLOUDSDK_CORE_PROJECT}" addresses create "${IP_ADDRESS_NAME}" --region="${CLOUDSDK_COMPUTE_REGION}" --network-tier=PREMIUM

gcloud compute addresses list --filter="name=${IP_ADDRESS_NAME}" --format="value(ADDRESS)"
