#!/usr/bin/env bash

#Description: Removes DNS record with given subdomain from the DNS Zone.
#
#Expected vars:
# - CLOUDSDK_CORE_PROJECT: name of a GCP project containing the Zone with the record.
# - CLOUDSDK_DNS_ZONE_NAME: Name of the existing DNS zone in the project (NOT it's DNS name!)
# - DNS_FULL_NAME: DNS name
# - IP_ADDRESS: v4 IP Address of the DNS record.
#
#Permissions: In order to run this script you need to use a service account with "DNS Administrator" role
set -o errexit

discoverUnsetVar=false

for var in CLOUDSDK_CORE_PROJECT CLOUDSDK_DNS_ZONE_NAME DNS_FULL_NAME IP_ADDRESS; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done

if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

retries=3
retryTimeInSec="5"
function deleteDNSWithRetries() {
    set +e

    for ((i=1; i<=retries; i++)); do
        gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction start --zone="${CLOUDSDK_DNS_ZONE_NAME}" && \
        gcloud dns record-sets transaction remove "${IP_ADDRESS}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${DNS_FULL_NAME}" --type=A --ttl=300 && \
        gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction execute --zone="${CLOUDSDK_DNS_ZONE_NAME}"

        [[ $? -eq 0 ]] && break

        gcloud dns record-sets transaction abort --zone="${CLOUDSDK_DNS_ZONE_NAME}" --verbosity none

        if [[ "${i}" -lt "${retries}" ]]; then
            echo "Unable to delete DNS record, let's wait ${retryTimeInSec} seconds and retry [${i}/${retries}]."
        else
            echo "Unable to delete DNS record after ${retries} retries, giving up."
        fi

        sleep ${retryTimeInSec}
    done

    set -e
}

deleteDNSWithRetries
echo "DNS Record deleted, but it can be visible for some time due to DNS caches"