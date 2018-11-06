#!/usr/bin/env bash

set -o errexit

if [ -z "$GCLOUD_PROJECT_NAME" ]; then
    echo "\$GCLOUD_PROJECT_NAME is empty"
    exit 1
fi

if [ -z "$GCLOUD_REGION" ]; then
    echo "\$GCLOUD_IP_ADDRESS_NAME is empty"
    exit 1
fi

if [ -z "$GCLOUD_IP_ADDRESS_NAME" ]; then
    echo "\$GCLOUD_IP_ADDRESS_NAME is empty"
    exit 1
fi

gcloud beta compute --project=${GCLOUD_PROJECT_NAME} addresses create ${GCLOUD_IP_ADDRESS_NAME} --region=${GCLOUD_REGION} --network-tier=PREMIUM

gcloud compute addresses list --filter="name=${GCLOUD_IP_ADDRESS_NAME}" --format="value(ADDRESS)"
