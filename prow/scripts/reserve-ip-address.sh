#!/usr/bin/env bash

set -o errexit

if [ -z "$PROJECT" ]; then
    echo "\$PROJECT is empty"
    exit 1
fi

if [ -z "$GCLOUD_REGION" ]; then
    echo "\$GCLOUD_REGION is empty"
    exit 1
fi

if [ -z "$GCLOUD_IP_ADDRESS_NAME" ]; then
    echo "\$GCLOUD_IP_ADDRESS_NAME is empty"
    exit 1
fi

gcloud beta compute --project="${PROJECT}" addresses create "${GCLOUD_IP_ADDRESS_NAME}" --region="${GCLOUD_REGION}" --network-tier=PREMIUM

gcloud compute addresses list --filter="name=${GCLOUD_IP_ADDRESS_NAME}" --format="value(ADDRESS)"
