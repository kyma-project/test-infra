#!/usr/bin/env bash

#Description: Adds new type "A" DNS entry for given subdomain and IP Address
#
#Expected vars:
# - GCLOUD_PROJECT_NAME: name of a GCP project where new DNS record is created.
# - DNS_ZONE_NAME: Name of an existing DNS zone in the project (NOT it's DNS name!)
# - DNS_SUBDOMAIN: a subdomain to create entry for. Example: If DNS_ZONE_NAME referrs to a DNS Zone that controls "cool.dot.com" domain, use DNS_SUBDOMAIN="how" to create an entry: how.cool.dot.com
# - IP_ADDRESS: v4 IP Address for the DNS record.
#
#Permissions: In order to run this script you need to use a service account with "DNS Administrator" role

set -o errexit

discoverUnsetVar=false

for var in GCLOUD_PROJECT_NAME DNS_ZONE_NAME DNS_SUBDOMAIN IP_ADDRESS; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done

if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

DNS_DOMAIN="$(gcloud dns managed-zones describe "${DNS_ZONE_NAME}" --format="value(dnsName)")"
DNS_FULL_NAME="${DNS_SUBDOMAIN}.${DNS_DOMAIN}"

gcloud dns --project="${GCLOUD_PROJECT_NAME}" record-sets transaction start --zone="${DNS_ZONE_NAME}"

gcloud dns --project="${GCLOUD_PROJECT_NAME}" record-sets transaction add "${IP_ADDRESS}" --name="${DNS_FULL_NAME}" --ttl=300 --type=A --zone="${DNS_ZONE_NAME}"

gcloud dns --project="${GCLOUD_PROJECT_NAME}" record-sets transaction execute --zone="${DNS_ZONE_NAME}"

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

done

echo "Cannot resolve ${DNS_FULL_NAME} to expected IP_ADDRESS: ${IP_ADDRESS}."
exit 1
