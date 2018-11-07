#!/usr/bin/env bash
set -e
set -o pipefail

echo "Authenticating to Google Cloud..."
export CLOUDSDK_COMPUTE_ZONE="europe-west3-a"
export CLOUDSDK_COMPUTE_REGION="europe-west3"
gcloud config set project kyma-project
gcloud auth activate-service-account --key-file "${GOOGLE_APPLICATION_CREDENTIALS}"

# Get list of ssh-keys, remove header line and print only first column which is key
out=$(gcloud compute os-login ssh-keys list | sed '1 d' | awk -F "\t" '{print $1}')

for id in ${out}; do
    echo "Removing key ${id} ..."
    gcloud compute os-login ssh-keys remove --key ${id}
done;

echo "DONE"