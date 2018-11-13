#!/usr/bin/env bash

set -o errexit

discoverUnsetVar=false

for var in CLOUDSDK_COMPUTE_REGION GCLOUD_IP_ADDRESS_NAME; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done

if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

gcloud compute addresses delete "${GCLOUD_IP_ADDRESS_NAME}" --region "${CLOUDSDK_COMPUTE_REGION}" --quiet
