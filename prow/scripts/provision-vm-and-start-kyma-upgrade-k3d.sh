#!/bin/bash

# This script is designed to provision a new vm and start and upgrade kyma.It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisining vms. If no flag is provided, the latest custom image is used.

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
  log::info "Execute Job Guard"
  "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

cleanup() {
  log::info "Wait before cleanup - start"
  sleep 3h 30m
  log::info "Wait before cleanup - end"

  # TODO - collect junit results
  log::info "Stopping instance kyma-upgrade-test-${RANDOM_ID}"
  log::info "It will be removed automatically by cleaner job"

  # do not fail the job regardless of the vm deletion result
  set +e

  #shellcheck disable=SC2088
  utils::receive_from_vm "${ZONE}" "kyma-upgrade-test-${RANDOM_ID}" "~/kyma/tests/fast-integration/junit_kyma-fast-integration.xml" "${ARTIFACTS}"
  gcloud compute instances stop --async --zone="${ZONE}" "kyma-upgrade-test-${RANDOM_ID}"

  log::info "End of cleanup"
}

function testCustomImage() {
  CUSTOM_IMAGE="$1"
  IMAGE_EXISTS=$(gcloud compute images list --filter "name:${CUSTOM_IMAGE}" | tail -n +2 | awk '{print $1}')
  if [[ -z "$IMAGE_EXISTS" ]]; then
    log::error "${CUSTOM_IMAGE} is invalid, it is not available in GCP images list, the script will terminate ..." && exit 1
  fi
}

gcp::authenticate \
  -c "${GOOGLE_APPLICATION_CREDENTIALS}"

RANDOM_ID=$(openssl rand -hex 4)

LABELS=""
if [[ -z "${PULL_NUMBER}" ]]; then
  LABELS=(--labels "branch=$PULL_BASE_REF,job-name=kyma-upgrade")
else
  LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=kyma-upgrade")
fi

POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"

  case ${key} in
    --image)
      IMAGE="$2"
      testCustomImage "${IMAGE}"
      shift
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


if [[ -z "$IMAGE" ]]; then
  log::info "Provisioning vm using the latest default custom image ..."

  IMAGE=$(gcloud compute images list --sort-by "~creationTimestamp" \
       --filter "family:custom images AND labels.default:yes" --limit=1 | tail -n +2 | awk '{print $1}')

  if [[ -z "$IMAGE" ]]; then
    log::error "There are no default custom images, the script will exit ..." && exit 1
  fi
fi

ZONE_LIMIT=${ZONE_LIMIT:-5}
EU_ZONES=$(gcloud compute zones list --filter="name~europe" --limit="${ZONE_LIMIT}" | tail -n +2 | awk '{print $1}')
STARTTIME=$(date +%s)
for ZONE in ${EU_ZONES}; do
  log::info "Attempting to create a new instance named kyma-upgrade-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
  gcloud compute instances create "kyma-upgrade-test-${RANDOM_ID}" \
      --metadata enable-oslogin=TRUE \
      --image "${IMAGE}" \
      --machine-type n2-standard-4 \
      --zone "${ZONE}" \
      --boot-disk-size 200 "${LABELS[@]}" && \
  log::info "Created kyma-upgrade-test-${RANDOM_ID} in zone ${ZONE}" && break
  log::error "Could not create machine in zone ${ZONE}"
done || exit 1
ENDTIME=$(date +%s)
echo "VM creation time: $((ENDTIME - STARTTIME)) seconds."

trap cleanup exit INT

log::info "Preparing environment variables for the instance"
envVars=(
  KYMA_PROJECT_DIR
  BOT_GITHUB_TOKEN
)
utils::save_env_file "${envVars[@]}"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "kyma-upgrade-test-${RANDOM_ID}" ".env" "~/.env"

log::info "Copying Kyma to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "kyma-upgrade-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-project/kyma" "~/kyma"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "kyma-upgrade-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-project/test-infra" "~/test-infra"

set -o pipefail
log::info "Triggering the installation"

utils::ssh_to_vm_with_script -z "${ZONE}" -n "kyma-upgrade-test-${RANDOM_ID}" -c "sudo bash" -p "${SCRIPT_DIR}/cluster-integration/kyma-upgrade-k3d-kyma2-to-main.sh"


log::success "all done"
