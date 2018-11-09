#!/usr/bin/env bash

set -o errexit

if [ -z "$GCLOUD_IP_ADDRESS_NAME" ]; then
    echo "\$GCLOUD_IP_ADDRESS_NAME is empty"
    exit 1
fi

if [ -z "$GCLOUD_REGION" ]; then
    echo "\$GCLOUD_REGION is empty"
    exit 1
fi

gcloud compute addresses delete "${GCLOUD_IP_ADDRESS_NAME}" --region "${GCLOUD_REGION}" --quiet
