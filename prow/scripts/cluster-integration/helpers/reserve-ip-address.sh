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

attempts=3
for ((i=1; i<=attempts; i++)); do
    gcloud beta compute --project="${CLOUDSDK_CORE_PROJECT}" addresses create "${IP_ADDRESS_NAME}" --region="${CLOUDSDK_COMPUTE_REGION}" --network-tier=PREMIUM
    if [[ $? -eq 0 ]]; then
        GATEWAY_IP_ADDRESS="$(gcloud compute addresses list --filter="name=${IP_ADDRESS_NAME}" --format="value(ADDRESS)")"
        echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"
        break
    elif [[ "${i}" -lt "${attempts}" ]]; then
        echo "Unable to create address: ${IP_ADDRESS_NAME}. Attempts ${i} of ${attempts}."
    else
        echo "Unable to create DNS record after ${attempts} attempts, giving up."
        exit 1
    fi
done
