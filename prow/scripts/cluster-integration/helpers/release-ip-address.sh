#!/usr/bin/env bash

#Description: Releases an IP Address
#
#Expected vars:
# - CLOUDSDK_COMPUTE_REGION: Region of the IP Address (e.g. europe-west3)
# - IP_ADDRESS_NAME: Name for the IP Address object (NOT an actual IP Address)
#
#Permissions: In order to run this script you need to use a service account with "Compute Network Admin" role

set -o errexit

discoverUnsetVar=false

for var in CLOUDSDK_COMPUTE_REGION IP_ADDRESS_NAME; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done

if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

gcloud compute addresses delete "${IP_ADDRESS_NAME}" --region "${CLOUDSDK_COMPUTE_REGION}"
