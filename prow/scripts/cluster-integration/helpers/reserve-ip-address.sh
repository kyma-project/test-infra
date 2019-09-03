#!/usr/bin/env bash

#Description: Reserves new IP Address and returns it on stdout
#
#Expected vars:
# - CLOUDSDK_CORE_PROJECT: name of a GCP project where IP Address is reserved
# - CLOUDSDK_COMPUTE_REGION: Region for the IP Address (e.g. europe-west3)
# - IP_ADDRESS_NAME: Name for the IP Address object (NOT an actual IP Address)
#
#Permissions: In order to run this script you need to use a service account with "Compute Network Admin" role

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

echo "Running gcloud beta compute --project=${CLOUDSDK_CORE_PROJECT} addresses create ${IP_ADDRESS_NAME} --region=${CLOUDSDK_COMPUTE_REGION} --network-tier=PREMIUM"
gcloud beta compute --project="${CLOUDSDK_CORE_PROJECT}" addresses create "${IP_ADDRESS_NAME}" --region="${CLOUDSDK_COMPUTE_REGION}" --network-tier=PREMIUM

gcloud compute addresses list --filter="name=${IP_ADDRESS_NAME}" --format="value(ADDRESS)"
