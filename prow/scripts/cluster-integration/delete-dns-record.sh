#!/usr/bin/env bash

set -o errexit

if [ -z "$PROJECT" ]; then
    echo "\$PROJECT is empty"
    exit 1
fi

if [ -z "$DNS_ZONE" ]; then
    echo "\$DNS_ZONE is empty"
    exit 1
fi

if [ -z "$IP_ADDRESS" ]; then
    echo "\$IP_ADDRESS is empty"
    exit 1
fi

if [ -z "$DNS_NAME" ]; then
    echo "\$DNS_NAME is empty"
    exit 1
fi

gcloud dns --project="${PROJECT}" record-sets transaction start --zone="${DNS_ZONE}"

gcloud dns record-sets transaction remove ${IP_ADDRESS} --zone=${DNS_ZONE} --name=${DNS_NAME} --type=A --ttl=300

gcloud dns --project=${PROJECT} record-sets transaction execute --zone=${DNS_ZONE}
