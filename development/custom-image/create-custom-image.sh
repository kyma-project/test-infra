#!/bin/bash

set -o errexit

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly ROOT_DIR=${CURRENT_DIR}/../../
# shellcheck source=prow/scripts/lib/log.sh
source "${ROOT_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${ROOT_DIR}/prow/scripts/lib/utils.sh"

cleanup() {   
    log::info "Removing instance kyma-deps-image-vm-${RANDOM_ID}"
    gcloud compute instances delete --quiet --zone "${ZONE}" "kyma-deps-image-vm-${RANDOM_ID}"
}


RANDOM_ID=$(openssl rand -hex 4)
DATE=$(date +v%Y%m%d)
DEFAULT=false

POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"

    case ${key} in
        --default)
            DEFAULT=true
            shift
            ;;
        --*)
            echo "Unknown flag ${1}"
            exit 1
            ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

if [[ "$DEFAULT" == true ]]; then
    LABELS=(--labels "default=yes")
fi

ZONE_LIMIT=${ZONE_LIMIT:-5}
EU_ZONES=$(gcloud compute zones list --filter="name~europe" --limit="${ZONE_LIMIT}" | tail -n +2 | awk '{print $1}')
for ZONE in ${EU_ZONES}; do
    log::info "Attempting to create a new instance named kyma-deps-image-vm-${RANDOM_ID} in zone ${ZONE} ..."
    gcloud compute instances create "kyma-deps-image-vm-${RANDOM_ID}" \
        --metadata enable-oslogin=TRUE \
        --machine-type n1-standard-4 \
        --image-family debian-9 \
        --image-project debian-cloud \
        --zone "${ZONE}" \
        --boot-disk-size 30 \
        --metadata-from-file startup-script=./machine-id-clean-up.sh  &&\
    log::info "Created kyma-deps-image-vm-${RANDOM_ID} in zone ${ZONE}" && break
    log::error "Could not create machine in zone ${ZONE}"
done

trap cleanup exit

log::info "Moving install-deps-debian.sh to kyma-deps-image-vm-${RANDOM_ID} in zone ${ZONE} ..."
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "kyma-deps-image-vm-${RANDOM_ID}" "./install-deps-debian.sh" "~/"

log::info "Running install-deps-debian.sh ..."
gcloud compute ssh --quiet --zone="${ZONE}" "kyma-deps-image-vm-${RANDOM_ID}" -- ./install-deps-debian.sh

log::info "Clearing kyma-deps-image-vm-${RANDOM_ID} machine-id ..."
gcloud compute ssh --zone "${ZONE}" "kyma-deps-image-vm-${RANDOM_ID}" --command "sudo sh -c 'echo "" > /etc/machine-id'"
gcloud compute ssh --zone "${ZONE}" "kyma-deps-image-vm-${RANDOM_ID}" --command "sudo sh -c 'echo "" > /var/lib/dbus/machine-id'"


log::info "Stopping kyma-deps-image-vm-${RANDOM_ID} in zone ${ZONE} ..."
gcloud compute instances stop --zone="${ZONE}" "kyma-deps-image-vm-${RANDOM_ID}"


log::info "Creating the new image kyma-deps-image-${DATE}..."
gcloud compute images create "kyma-deps-image-${DATE}" \
  --source-disk "kyma-deps-image-vm-${RANDOM_ID}" \
  --source-disk-zone "${ZONE}" \
  "${LABELS[@]}" \
  --family "custom-images"
