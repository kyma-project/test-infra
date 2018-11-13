#!/usr/bin/env bash

#Description: Removes DNS record with given subdomain from the DNS Zone.
#
#Expected vars:
# - GCLOUD_PROJECT_NAME: name of a GCP project containing the Zone with the record.
# - DNS_ZONE_NAME: Name of the existing DNS zone in the project (NOT it's DNS name!)
# - DNS_SUBDOMAIN: a subdomain in the Zone.
# - IP_ADDRESS: v4 IP Address of the DNS record.
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

DNS_DOMAIN="$(gcloud dns managed-zones describe ${DNS_ZONE_NAME} --format="value(dnsName)")"
DNS_FULL_NAME="${DNS_SUBDOMAIN}.${DNS_DOMAIN}"

gcloud dns --project="${GCLOUD_PROJECT_NAME}" record-sets transaction start --zone="${DNS_ZONE_NAME}"

gcloud dns record-sets transaction remove "${IP_ADDRESS}" --zone="${DNS_ZONE_NAME}" --name="${DNS_FULL_NAME}" --type=A --ttl=300

gcloud dns --project="${GCLOUD_PROJECT_NAME}" record-sets transaction execute --zone="${DNS_ZONE_NAME}"

echo "DNS Record deleted, but it can be visible for some time due to DNS caches"
