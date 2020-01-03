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

# Export variable used in subshell.
export IP_ADDRESS_NAME

counter=0
# Check if IP address reservation is present. Wait and retry for one minute to disappear. If IP reservation was removed just before it need few seconds to disappear.
# Otherwise, creation will fail.
IP_ADDRESS=$(gcloud compute addresses list --filter="name=${IP_ADDRESS_NAME}" --format="value(ADDRESS)")
until [[ -z ${IP_ADDRESS} ]]; do
	sleep 15
	counter=$(( counter + 1 ))
	IP_ADDRESS=$(gcloud compute addresses list --filter="name=${IP_ADDRESS_NAME}" --format="value(ADDRESS)")
	if (( counter == 5 )); then
		# Fail after one minute wait.
		echo "${IP_ADDRESS_NAME} IP address is still present after one minute wait. Failing"
		exit 1
	fi
done

gcloud compute addresses create "${IP_ADDRESS_NAME}" --project="${CLOUDSDK_CORE_PROJECT}" --region="${CLOUDSDK_COMPUTE_REGION}" --network-tier=PREMIUM
# Print reserved IP address on stdout as it's consumed by calling process and used for next steps.
gcloud compute addresses list --filter="name=${IP_ADDRESS_NAME}" --format="value(ADDRESS)"
