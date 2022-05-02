#!/bin/bash

set -o errexit

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly ROOT_DIR=${CURRENT_DIR}/../../
# shellcheck source=prow/scripts/lib/log.sh
source "${ROOT_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${ROOT_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "${ROOT_DIR}/prow/scripts/lib/gcp.sh"
cleanup() {   
    log::info "Removing instance $VM_NAME"
    gcloud compute instances delete --quiet --zone "${ZONE}" "$VM_NAME"
    if [ "$JOB_TYPE" == "presubmit" ]; then
      log::info "Removing image $IMAGE"
      gcloud compute images delete "$IMAGE"
    fi
}

if [ "$CI" == "true" ]; then
  gcp::authenticate \
    -c "$GOOGLE_APPLICATION_CREDENTIALS"
fi


RANDOM_ID=$(openssl rand -hex 4)
VM_NAME="kyma-deps-image-vm-${RANDOM_ID}"
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
    log::info "Attempting to create a new instance named $VM_NAME in zone ${ZONE} ..."
    gcloud compute instances create "$VM_NAME" \
        --metadata enable-oslogin=TRUE \
        --machine-type n1-standard-4 \
        --image-family debian-10 \
        --image-project debian-cloud \
        --zone "${ZONE}" \
        --boot-disk-size 200 \
        --metadata-from-file startup-script="$CURRENT_DIR"/machine-id-clean-up.sh  &&\
    log::info "Created $VM_NAME in zone ${ZONE}" && break
    log::error "Could not create machine in zone ${ZONE}"
done || exit 1

trap cleanup exit

log::info "Moving install-deps-debian.sh to $VM_NAME in zone ${ZONE} ..."
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "$VM_NAME" "$CURRENT_DIR/install-deps-debian.sh" "~/"

log::info "Running install-deps-debian.sh ..."
utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_NAME}" -c "./install-deps-debian.sh"

log::info "Clearing $VM_NAME machine-id ..."
utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_NAME}" -c "sudo sh -c 'echo "" > /etc/machine-id'"
utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_NAME}" -c "sudo sh -c 'echo "" > /var/lib/dbus/machine-id'"
utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_NAME}" -c "sudo sh -c 'echo \"RateLimitInterval=30s\" > /etc/systemd/journald.conf'"
utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_NAME}" -c "sudo sh -c 'echo \"RateLimitBurst=1500\" >> /etc/systemd/journald.conf'"
utils::send_to_vm "${ZONE}" "$VM_NAME" "$CURRENT_DIR/resources/dbus-1_system-local.conf" "/tmp/system-local.conf"
utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_NAME}" -c "sudo sh -c 'mv /tmp/system-local.conf /etc/dbus-1/system-local.conf'"


log::info "Stopping $VM_NAME in zone ${ZONE} ..."
gcloud compute instances stop --zone="${ZONE}" "$VM_NAME"

if [ "$JOB_TYPE" == "presubmit" ]; then
  IMAGE="$VM_NAME"
else
  IMAGE="kyma-deps-image-${DATE}-${PULL_BASE_SHA::6}"
fi

log::info "Creating the new image $IMAGE..."
gcloud compute images create "$IMAGE" \
  --source-disk "$VM_NAME" \
  --source-disk-zone "${ZONE}" \
  "${LABELS[@]}" \
  --family "custom-images"
