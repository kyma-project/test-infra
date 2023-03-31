#!/bin/bash

# This script is designed to provision a new vm and start kyma.It takes an optional positional parameter using --image flag
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
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
  export JOBGUARD_TIMEOUT="20m"
  export JOB_NAME_PATTERN="pre-main-reconciler-publish-pr-cli"

  log::info "Execute Job Guard to wait for job: ${JOB_NAME_PATTERN}"
  "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

cleanup() {
  log::banner "Job Cleanup"
  log::info "Stopping instance kyma-integration-test-${RANDOM_ID}"
  log::info "It will be removed automatically by cleaner job"

  # do not fail the job regardless of the vm deletion result
  set +e

  #shellcheck disable=SC2088
  utils::receive_from_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "~/kyma/tests/fast-integration/junit_kyma-fast-integration.xml" "${ARTIFACTS}"
  gcloud compute instances stop --async --zone="${ZONE}" "kyma-integration-test-${RANDOM_ID}"

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
  LABELS=(--labels "branch=$PULL_BASE_REF,job-name=kyma-integration")
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

log::banner "Provisioning VM on GCP"
ZONE_LIMIT=${ZONE_LIMIT:-5}
EU_ZONES=$(gcloud compute zones list --filter="name~europe" --limit="${ZONE_LIMIT}" | tail -n +2 | awk '{print $1}')
STARTTIME=$(date +%s)
for ZONE in ${EU_ZONES}; do
  log::info "Attempting to create a new instance named kyma-integration-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
  gcloud compute instances create "kyma-integration-test-${RANDOM_ID}" \
      --metadata enable-oslogin=TRUE \
      --image "${IMAGE}" \
      --machine-type n2-standard-4 \
      --zone "${ZONE}" \
      --boot-disk-size 200 "${LABELS[@]}" && \
  log::info "Created kyma-integration-test-${RANDOM_ID} in zone ${ZONE}" && break
  log::error "Could not create machine in zone ${ZONE}"
done || exit 1
ENDTIME=$(date +%s)
echo "VM creation time: $((ENDTIME - STARTTIME)) seconds."

trap cleanup exit INT


log::banner "Provision k3d, deploy Kyma and run fast-integration tests"

log::info "Define kyma version to deploy"
export KYMA_SOURCE="main"
if [[ "${KYMA_TEST_SOURCE}" == "latest-release" ]]; then
  # Fetch latest Kyma released version
  kyma_get_last_release_version_return_version=$(curl --silent --fail --show-error -H "Authorization: token ${BOT_GITHUB_TOKEN}" "https://api.github.com/repos/kyma-project/kyma/releases" \
      | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-1].target_commitish | split("/") | .[-1]')
  export KYMA_SOURCE="${kyma_get_last_release_version_return_version:?}"

  log::info "### Reading latest release version from RELEASE_VERSION file, got: ${KYMA_SOURCE}"
elif [[ "${KYMA_TEST_SOURCE}" == "previous-release" ]]; then
  # Fetch previous Kyma released version
  kyma_get_previous_release_version_return_version=$(curl --silent --fail --show-error -H "Authorization: token ${BOT_GITHUB_TOKEN}" "https://api.github.com/repos/kyma-project/kyma/releases" \
      | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-2].target_commitish | split("/") | .[-1]')
  export KYMA_SOURCE="${kyma_get_previous_release_version_return_version:?}"
  log::info "### Reading previous release version from RELEASE_VERSION file, got: ${KYMA_SOURCE}"
fi

### define Kyma version to upgrade to, if it is a upgrade test
if [[ "${KYMA_UPGRADE_SOURCE}" == "latest-release" ]]; then
  # Fetch latest Kyma released version
  kyma_get_last_release_version_return_version=$(curl --silent --fail --show-error -H "Authorization: token ${BOT_GITHUB_TOKEN}" "https://api.github.com/repos/kyma-project/kyma/releases" \
        | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-1].target_commitish | split("/") | .[-1]')
  export KYMA_UPGRADE_VERSION="${kyma_get_last_release_version_return_version:?}"
  log::info "### Reading latest release version from RELEASE_VERSION file, got: ${KYMA_UPGRADE_VERSION}"
fi

log::info "Preparing environment variables for the instance"
envVars=(
  PULL_NUMBER
  EXECUTION_PROFILE
  KYMA_SOURCE
  KYMA_UPGRADE_VERSION
)
utils::save_env_file "${envVars[@]}"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" ".env" "~/.env"

log::info "Copying Kyma to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-project/kyma" "~/kyma"

log::info "Copying Test-infra to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-project/test-infra" "~/test-infra"

log::info "Run the relevant script to deploy Kyma and run fast-integration tests"
if [[ "${KYMA_UPGRADE_VERSION}" ]]; then
  log::banner "Triggering the tests for Kyma upgrade scenario from version: ${KYMA_SOURCE} to version: ${KYMA_UPGRADE_VERSION}"
  utils::ssh_to_vm_with_script -z "${ZONE}" -n "kyma-integration-test-${RANDOM_ID}" -c "sudo bash" -p "${SCRIPT_DIR}/cluster-integration/reconciler-integration-with-cli-upgrade-k3d.sh"

else
  log::banner "Triggering the tests for Kyma deploy scenario for version: ${KYMA_SOURCE}"
  utils::ssh_to_vm_with_script -z "${ZONE}" -n "kyma-integration-test-${RANDOM_ID}" -c "sudo bash" -p "${SCRIPT_DIR}/cluster-integration/reconciler-integration-with-cli-k3d.sh"

fi

log::success "all done"