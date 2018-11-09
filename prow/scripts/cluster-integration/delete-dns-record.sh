#!/usr/bin/env bash

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
