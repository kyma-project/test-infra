#!/usr/bin/env bash
# Description: creates secret containing GCP service account credentials to use GCP buckets with minio gateway mode
# Expected vars:
# - GOOGLE_APPLICATION_CREDENTIALS   - GCP Service Account key file path
# - MINIO_GCS_GATEWAY_GCS_KEY_SECRET - the name of the secret object that will store service account credentials
set -o errexit

discoverUnsetVar=false

for var in MINIO_GCS_GATEWAY_GCS_KEY_SECRET GOOGLE_APPLICATION_CREDENTIALS; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done

if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

kubectl create secret generic "${MINIO_GCS_GATEWAY_GCS_KEY_SECRET}" --from-literal=service-account.json="${GOOGLE_APPLICATION_CREDENTIALS}" --namespace kyma-system