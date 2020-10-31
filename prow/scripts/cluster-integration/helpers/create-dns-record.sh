#!/usr/bin/env bash

#Description: Adds new type "A" DNS entry for given subdomain and IP Address
#
#Expected vars:
# - CLOUDSDK_CORE_PROJECT: name of a GCP project where new DNS record is created.
# - CLOUDSDK_DNS_ZONE_NAME: Name of an existing DNS zone in the project (NOT its DNS name!)
# - DNS_FULL_NAME: DNS name
# - IP_ADDRESS: v4 IP Address for the DNS record.
#
#Permissions: In order to run this script you need to use a service account with "DNS Administrator" role

set -o errexit

SCRIPTS_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../.."
#shellcheck source=lib/log.sh
#shellcheck disable=SC1090
source "${SCRIPTS_PATH}/lib/log.sh"

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

attempts=10
retryTimeInSec="5"
function createDNSWithRetries() {
    set +e

    for ((i=1; i<=attempts; i++)); do
        gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction start --zone="${CLOUDSDK_DNS_ZONE_NAME}" && \
        gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction add "${IP_ADDRESS}" --name="${DNS_FULL_NAME}" --ttl=60 --type=A --zone="${CLOUDSDK_DNS_ZONE_NAME}" && \
        gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction execute --zone="${CLOUDSDK_DNS_ZONE_NAME}"

        if [[ $? -eq 0 ]]; then
            break
        fi

        gcloud dns record-sets transaction abort --zone="${CLOUDSDK_DNS_ZONE_NAME}" --verbosity none

        if [[ "${i}" -lt "${attempts}" ]]; then
            echo "Unable to create DNS record, let's wait ${retryTimeInSec} seconds and retry. Attempts ${i} of ${attempts}."
        else
            echo "Unable to create DNS record after ${attempts} attempts, giving up."
            exit 1
        fi

        sleep ${retryTimeInSec}
    done

    set -e
}

createDNSWithRetries

SECONDS=0
END_TIME=$((SECONDS+600)) #600 seconds == 10 minutes

while [ ${SECONDS} -lt ${END_TIME} ];do
    echo "Trying to resolve ${DNS_FULL_NAME}"
    sleep 10

    RESOLVED_IP_ADDRESS=$(dig +short "${DNS_FULL_NAME}")

    if [ "${RESOLVED_IP_ADDRESS}" = "${IP_ADDRESS}" ]; then
        echo "Successfully resolved ${DNS_FULL_NAME} to ${RESOLVED_IP_ADDRESS}"
        exit 0
    fi


    log::banner "Debugging DNS issues"
    log::date
    log::info "trace DNS response for ${DNS_FULL_NAME}"
    dig +trace "${DNS_FULL_NAME}"
    log::info "query authoritative servers directly"
    log::info "ns-cloud-b1.googledomains.com."
    dig "${DNS_FULL_NAME}" @ns-cloud-b1.googledomains.com.
    log::info "ns-cloud-b2.googledomains.com."
    dig "${DNS_FULL_NAME}" @ns-cloud-b2.googledomains.com.
    log::info "ns-cloud-b3.googledomains.com."
    dig "${DNS_FULL_NAME}" @ns-cloud-b3.googledomains.com.
    log::info "ns-cloud-b4.googledomains.com."
    dig "${DNS_FULL_NAME}" @ns-cloud-b4.googledomains.com.


done

echo "Cannot resolve ${DNS_FULL_NAME} to expected IP_ADDRESS: ${IP_ADDRESS}."
exit 1
