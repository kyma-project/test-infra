#!/bin/bash

set -o errexit

cleanup() {
    ARG=$?
    echo "
################################################################################
# Removing instance kyma-integration-test-${RANDOM_ID}
################################################################################
    "

    gcloud compute instances delete "kyma-integration-test-${RANDOM_ID}"
    exit $ARG
}

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"
authenticate

RANDOM_ID=$(< /dev/urandom tr -dc 'a-z0-9' | fold -w 8 | head -n 1)

LABELS=""
if [[ -z "${PULL_NUMBER}" ]]; then
    LABELS=(--labels "branch=$PULL_BASE_REF,job-name=kyma-integration")
else
    LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=kyma-integration")
fi

echo "
################################################################################
# Creating a new instance named kyma-integration-test-${RANDOM_ID}
################################################################################
"

gcloud compute instances create "kyma-integration-test-${RANDOM_ID}" \
    --metadata enable-oslogin=TRUE --image debian-9-stretch-v20181011 \
    --image-project debian-cloud --machine-type n1-standard-4 --boot-disk-size 20 "${LABELS[@]}"

trap cleanup exit

echo "
################################################################################
# Copying Kyma to the instance
################################################################################
"

for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && echo 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute scp --quiet --recurse /home/prow/go/src/github.com/kyma-project/kyma "kyma-integration-test-${RANDOM_ID}":~/kyma && break;
    [[ ${i} -ge 5 ]] && echo "Failed after $i attempts." && exit 1
done;

echo "
################################################################################
# Triggering the installation
################################################################################
"

gcloud compute ssh --quiet "kyma-integration-test-${RANDOM_ID}" -- ./kyma/prow/kyma-integration-on-debian/deploy-and-test-kyma.sh