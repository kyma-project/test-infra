#!/usr/bin/env bash
set -e
set -o pipefail

if [ -z "$CLOUDSDK_CORE_PROJECT" ]; then
    echo "Environment variable CLOUDSDK_CORE_PROJECT is empty"
    exit 1
fi

if [ -z "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
    echo "Environment variable GOOGLE_APPLICATION_CREDENTIALS is empty"
    exit 1
fi

if [ -z "$CLOUDSDK_COMPUTE_ZONE" ]; then
    echo "Environment variable CLOUDSDK_COMPUTE_ZONE is empty"
    exit 1
fi

if [ -z "$CLOUDSDK_COMPUTE_REGION" ]; then
    echo "Environment variable CLOUDSDK_COMPUTE_REGION is empty"
    exit 1
fi

echo "Authenticating to Google Cloud..."
gcloud config set project ${CLOUDSDK_CORE_PROJECT}
gcloud auth activate-service-account --key-file "${GOOGLE_APPLICATION_CREDENTIALS}"

# Get list of ssh-keys, remove header line and print only first column which is key
out=$(gcloud compute os-login ssh-keys list | sed '1 d' | awk -F "\t" '{print $1}')

for id in ${out}; do
    echo "Removing key ${id} ..."
    gcloud compute os-login ssh-keys remove --key ${id}
done;

echo "DONE"