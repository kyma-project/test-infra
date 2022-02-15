#!/bin/bash

# This script is designed to provision a new vm and start kyma.It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisining vms. If no flag is provided, the latest custom image is used.

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"

export CONTROL_PLANE_RECONCILER_DIR="/home/prow/go/src/github.com/kyma-project/control-plane/tools/reconciler"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
  log::info "Execute Job Guard"
  "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

cleanup() {
  # TODO - collect junit results
  log::info "Stopping instance ${INSTANCE_NAME}"
  log::info "It will be removed automatically by cleaner job"

  # do not fail the job regardless of the vm deletion result
  set +e

  gcloud compute instances stop --async --zone="${ZONE}" "${INSTANCE_NAME}"

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
INSTANCE_NAME="reconciler-fi-test-${RANDOM_ID}"

LABELS=""
if [[ -z "${PULL_NUMBER}" ]]; then
  LABELS=(--labels "branch=$PULL_BASE_REF,job-name=reconciler-integration")
else
  LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=reconciler-integration")
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
  log::info "Attempting to create a new instance named ${INSTANCE_NAME} in zone ${ZONE} using image ${IMAGE}"
  gcloud compute instances create "${INSTANCE_NAME}" \
      --metadata enable-oslogin=TRUE \
      --image "${IMAGE}" \
      --machine-type n2-standard-4 \
      --zone "${ZONE}" \
      --boot-disk-size 200 "${LABELS[@]}" && \
  log::info "Created ${INSTANCE_NAME} in zone ${ZONE}" && break
  log::error "Could not create machine in zone ${ZONE}"
done || exit 1
ENDTIME=$(date +%s)
echo "VM creation time: $((ENDTIME - STARTTIME)) seconds."

trap cleanup exit INT

# Fetch latest Kyma2 release
kyma::get_last_release_version -t "${BOT_GITHUB_TOKEN}"
export KYMA_UPGRADE_SOURCE="${kyma_get_last_release_version_return_version:?}"
log::info "### Reading release version from RELEASE_VERSION file, got: ${KYMA_UPGRADE_SOURCE}"

log::info "Preparing environment variables for the instance"
envVars=(
  KYMA_PROJECT_DIR
  KYMA_UPGRADE_SOURCE
)

log::info "Copying Kyma to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "${INSTANCE_NAME}" "/home/prow/go/src/github.com/kyma-project/kyma" "${KYMA_PROJECT_DIR}/kyma"

log::info "Copying control-plane to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "${INSTANCE_NAME}" "/home/prow/go/src/github.com/kyma-project/control-plane" "${KYMA_PROJECT_DIR}/control-plane"

log::info "Copying Test-infra to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "${INSTANCE_NAME}" "/home/prow/go/src/github.com/kyma-project/test-infra" "${KYMA_PROJECT_DIR}/test-infra"

log::info "Copying envVars to the instance"
utils::save_env_file "${envVars[@]}"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "${INSTANCE_NAME}" ".env" "~/.env"

log::info "Triggering the installation and tests"
gcloud compute ssh --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" --verbosity="${GCLOUD_SSH_LOG_LEVEL:-error}" --quiet --zone="${ZONE}" --command="sudo bash" --ssh-flag="-o ServerAliveInterval=30" "${INSTANCE_NAME}" < "${SCRIPT_DIR}/cluster-integration/reconciler-fast-integration-k3d.sh"

log::success "all done"
