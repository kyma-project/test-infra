#!/bin/bash

set -o errexit

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly ROOT_DIR=${CURRENT_DIR}/../../
# shellcheck disable=SC1090
source "${ROOT_DIR}/prow/scripts/library.sh"

cleanup() {   
    shout "Removing instance kyma-deps-image-vm-${RANDOM_ID}"
    gcloud compute instances delete --quiet --zone "${ZONE}" "kyma-deps-image-vm-${RANDOM_ID}"
}


RANDOM_ID=$(openssl rand -hex 4)
DATE=$(date +v%Y%m%d)
ZONE=$(openssl rand -hex 4)

ZONE_LIMIT=${ZONE_LIMIT:-5}
EU_ZONES=$(gcloud compute zones list --filter="name~europe" --limit="${ZONE_LIMIT}" | tail -n +2 | awk '{print $1}')
for ZONE in ${EU_ZONES}; do
    shout "Attempting to create a new instance named kyma-deps-image-vm-${RANDOM_ID} in zone ${ZONE}"
    gcloud compute instances create "kyma-deps-image-vm-${RANDOM_ID}" \
        --metadata enable-oslogin=TRUE \
        --machine-type n1-standard-4 \
        --image debian-9-stretch-v20181011 \
        --image-project debian-cloud \
        --zone "${ZONE}" \
        --boot-disk-size 20 &&\
    shout "Created kyma-deps-image-vm-${RANDOM_ID} in zone ${ZONE}" && break
    shout "Could not create machine in zone ${ZONE}"
done

trap cleanup exit

shout "Moving install-deps-debian.sh to kyma-deps-image-vm-${RANDOM_ID} in zone ${ZONE} ..."
for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && echo 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute scp --quiet --zone="${ZONE}" --recurse  ./install-deps-debian.sh "kyma-deps-image-vm-${RANDOM_ID}":~/ && break;
    [[ ${i} -ge 5 ]] && echo "Failed after $i attempts." && exit 1
done;


shout "Running install-deps-debian.sh ..."
gcloud compute ssh --quiet --zone="${ZONE}" "kyma-deps-image-vm-${RANDOM_ID}" -- ./install-deps-debian.sh


shout "Stopping kyma-deps-image-vm-${RANDOM_ID} in zone ${ZONE}"
gcloud compute instances stop --zone="${ZONE}" "kyma-deps-image-vm-${RANDOM_ID}"


shout "Creating the new image kyma-deps-image-${DATE}..."
gcloud compute images create "kyma-deps-image-${DATE}" \
  --source-disk "kyma-deps-image-vm-${RANDOM_ID}" \
  --source-disk-zone "${ZONE}" \
  --family "custom-images"
