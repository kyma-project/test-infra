#!/bin/bash

# This script is designed to provision a new vm and start kyma.It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisining vms. If no flag is provided, the latest custom image is used.

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"
readonly KYMA_PROJECT_DIR="$(cd "${SCRIPT_DIR}/../../../" && pwd)"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/cluster-integration/helpers/integration-tests.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/integration-tests.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
  log::info "Execute Job Guard"
  "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

cleanup() {
  # TODO - collect junit results
  log::info "Stopping instance reconciler-istio-integration-test-${RANDOM_ID}"
  log::info "It will be removed automatically by cleaner job"

  # do not fail the job regardless of the vm deletion result
  set +e

  gcloud compute instances stop --async --zone="${ZONE}" "reconciler-istio-integration-test-${RANDOM_ID}"

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
  log::info "Attempting to create a new instance named reconciler-istio-integration-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
  gcloud compute instances create "reconciler-istio-integration-test-${RANDOM_ID}" \
      --metadata enable-oslogin=TRUE \
      --image "${IMAGE}" \
      --machine-type n2-standard-4 \
      --zone "${ZONE}" \
      --boot-disk-size 200 "${LABELS[@]}" && \
  log::info "Created reconciler-istio-integration-test-${RANDOM_ID} in zone ${ZONE}" && break
  log::error "Could not create machine in zone ${ZONE}"
done || exit 1
ENDTIME=$(date +%s)
echo "VM creation time: $((ENDTIME - STARTTIME)) seconds."

trap cleanup exit INT

# Determine Kyma version from the latest release
if [[ ! $KYMA_VERSION ]]; then
    # Fetch latest Kyma2 release
    kyma::get_last_release_version -t "${BOT_GITHUB_TOKEN}"
    export KYMA_VERSION="${kyma_get_last_release_version_return_version:?}"
    log::info "Reading latest Kyma release version, got: ${KYMA_VERSION}"
fi
# Determine Istio version based on Kyma version
istio::get_version
export ISTIO_VERSION="${istio_version:?}"
log::info "Reading Istio version from ${KYMA_VERSION}, got: ${ISTIO_VERSION}"

log::info "Preparing environment variables for the instance"
envVars=(
  TEST_NAME
  EXECUTION_PROFILE
  KYMA_VERSION
  ISTIO_VERSION
)
utils::save_env_file "${envVars[@]}"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "reconciler-istio-integration-test-${RANDOM_ID}" ".env" "~/.env"

log::info "Copying Reconciler to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "reconciler-istio-integration-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-incubator/reconciler" "~/reconciler"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "reconciler-istio-integration-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-project/test-infra" "~/test-infra"

log::info "Triggering the installation"
utils::ssh_to_vm_with_script -z "${ZONE}" -n "reconciler-istio-integration-test-${RANDOM_ID}" -c "sudo bash" -p "${SCRIPT_DIR}/cluster-integration/reconciler-istio-integration.sh"

log::success "all done"
