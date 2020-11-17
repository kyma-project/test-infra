#!/usr/bin/env bash

set +e

gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction start --zone="${CLOUDSDK_DNS_ZONE_NAME}"
gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction remove "${IP_ADDRESS}" --name="${DNS_FULL_NAME}" --ttl=60 --type=A --zone="${CLOUDSDK_DNS_ZONE_NAME}"
gcloud dns --project="${CLOUDSDK_CORE_PROJECT}" record-sets transaction execute --zone="${CLOUDSDK_DNS_ZONE_NAME}"

set -e


