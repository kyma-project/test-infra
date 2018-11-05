#!/usr/bin/env bash

#In order to run this script you need to use a service account with DNS Administrator role

set -o errexit

if [ -z "$IP_ADDRESS" ]; then
    echo "\$IP_ADDRESS is empty"
    exit 1
fi

if [ -z "$PROJECT" ]; then
    echo "\$PROJECT is empty"
    exit 1
fi

if [ -z "$DNS_ZONE" ]; then
    echo "\$DNS_ZONE is empty"
    exit 1
fi

if [ -z "$DOMAIN" ]; then
    echo "\$DOMAIN is empty"
    exit 1
fi

RANDOM_STRING=$(cat /dev/urandom | env LC_CTYPE=C tr -dc a-z0-9 | head -c 16; echo)
DNS_NAME="pull-${RANDOM_STRING}.${DOMAIN}"

gcloud dns --project=${PROJECT} record-sets transaction start --zone=${DNS_ZONE}

gcloud dns --project=${PROJECT} record-sets transaction add ${IP_ADDRESS} --name=${DNS_NAME} --ttl=300 --type=A --zone=${DNS_ZONE}

gcloud dns --project=${PROJECT} record-sets transaction execute --zone=${DNS_ZONE}

SECONDS=0
END_TIME=$((SECONDS+600)) #600 seconds == 10 minutes

while [ ${SECONDS} -lt ${END_TIME} ];do
    echo "Trying to resolve ${DNS_NAME}"
    sleep 10

    RESOLVED_IP_ADDRESS=$(dig +short ${DNS_NAME})

    if [ ! -z "${RESOLVED_IP_ADDRESS}" ]; then
        echo "Successfully resolved ${DNS_NAME} to ${RESOLVED_IP_ADDRESS}"
        break
    fi
done

if [ -z ${RESOLVED_IP_ADDRESS} ]; then
    echo "Can't resolve ${DNS_NAME}"
    exit 1
fi

if [ "${RESOLVED_IP_ADDRESS}" != "${IP_ADDRESS}" ]; then
    echo "Error. Resolved hostname (${RESOLVED_IP_ADDRESS}) doesn't match input IP address (${IP_ADDRESS})."
    exit 1
fi
