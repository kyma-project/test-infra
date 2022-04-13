#!/bin/bash

# This script is designed to provision a new vm and start kyma.It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisining vms. If no flag is provided, the latest custom image is used.

set -o errexit

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"
readonly TMP_DIR=$(mktemp -d)
readonly JUNIT_REPORT_PATH="${ARTIFACTS:-${TMP_DIR}}/junit_kyma_octopus-test-suite.xml"
readonly testSuiteScript=${TEST_SUITE_SCRIPT:-"cluster-integration/kyma-integration-minikube.sh"}
readonly TEST_SUITE=${TEST_SUITE:-"testsuite-all"}

if [[ ! -f ${SCRIPT_DIR}/${testSuiteScript} ]]; then
  echo "Fatal: selected script ${SCRIPT_DIR}/${testSuiteScript} does not exist"
  exit 1
fi

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
  log::info "Fetch JUnit test results and store them in job artifacts"
  gcloud compute scp --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" --verbosity="${GCLOUD_SCP_LOG_LEVEL:-error}" --quiet --zone="${ZONE}" "kyma-integration-test-${RANDOM_ID}:junit_kyma_octopus-test-suite.xml" "${JUNIT_REPORT_PATH}"
  ARG=$?
  log::info "Removing instance kyma-integration-test-${RANDOM_ID}"
  gcloud compute instances delete --zone="${ZONE}" "kyma-integration-test-${RANDOM_ID}" || true ### Workaround: not failing the job regardless of the vm deletion result
  exit $ARG
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
  LABELS=(--labels "commit-id=${PULL_BASE_SHA::8},job-name=$JOB_NAME")
else
  LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=kyma-integration")
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
  *) # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift              # past argument
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
  log::info "Attempting to create a new instance named kyma-integration-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
  gcloud compute instances create "kyma-integration-test-${RANDOM_ID}" \
    --metadata enable-oslogin=TRUE \
    --image "${IMAGE}" \
    --machine-type n1-standard-4 \
    --zone "${ZONE}" \
    --boot-disk-size 200 "${LABELS[@]}" &&
    log::info "Created kyma-integration-test-${RANDOM_ID} in zone ${ZONE}" && break
  log::error "Could not create machine in zone ${ZONE}"
done || exit 1
ENDTIME=$(date +%s)
echo "VM creation time: $((ENDTIME - STARTTIME)) seconds."

if [[ ${TEST_SUITE} == "testsuite-all" ]]; then
  trap cleanup exit INT
fi

log::info "Copying Kyma to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-project/kyma" "~/kyma"

log::info "Triggering the installation"
log::info "Running testsuite ${testSuiteScript}"
gcloud compute ssh --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" --verbosity="${GCLOUD_SSH_LOG_LEVEL:-error}" --strict-host-key-checking=no --quiet --zone="${ZONE}" \
  --command="sudo PULL_NUMBER=${PULL_NUMBER}  TEST_SUITE=${TEST_SUITE} bash" \
  --ssh-flag="-o ServerAliveInterval=10 -o TCPKeepAlive=no -o ServerAliveCountMax=60" "kyma-integration-test-${RANDOM_ID}" <"${SCRIPT_DIR}/${testSuiteScript}"

log::success "all done"
