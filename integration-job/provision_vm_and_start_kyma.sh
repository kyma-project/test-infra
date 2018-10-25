#!/bin/bash

set -o errexit

cleanup() {
    ARG=$?
    echo "Removing instance kyma-integration-test-${RANDOM_ID}..."
    gcloud compute instances delete kyma-integration-test-${RANDOM_ID}
    exit $ARG
}

echo "Authenticating to Google Cloud..."
export CLOUDSDK_COMPUTE_ZONE="europe-west3-a"
export CLOUDSDK_COMPUTE_REGION="europe-west3"
gcloud config set project kyma-project
gcloud auth activate-service-account --key-file /var/run/secret/cloud.google.com/key.json

RANDOM_ID=$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 8 | head -n 1)

LABELS=""
if [[ -z "${PULL_NUMBER}" ]]; then
    LABELS="--labels branch=$PULL_BASE_REF,job-name=kyma-integration"
else
    LABELS="--labels pull-number=$PULL_NUMBER,job-name=kyma-integration"
fi

echo "Creating a new instance named kyma-integration-test-${RANDOM_ID}..."
gcloud compute instances create kyma-integration-test-${RANDOM_ID} \
    --metadata enable-oslogin=TRUE --image debian-9-stretch-v20181009 \
    --image-project debian-cloud --machine-type n1-standard-4 --boot-disk-size 20 $LABELS

trap cleanup exit

echo "Copying the installation script to the instance..."
gcloud compute scp --quiet install_deps_and_run_kyma.sh kyma-integration-test-${RANDOM_ID}:~/install_deps_and_run_kyma.sh

echo "Triggering the installation script..."
gcloud compute ssh --quiet kyma-integration-test-${RANDOM_ID} -- ./install_deps_and_run_kyma.sh --repo-owner $REPO_OWNER --branch $PULL_BASE_REF --pr-number $PULL_NUMBER