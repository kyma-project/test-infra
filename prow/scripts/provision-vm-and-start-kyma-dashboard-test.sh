#!/bin/bash

# This script is designed to provision a new vm and start kyma.It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisining vms. If no flag is provided, the latest custom image is used.

set -o errexit
set -o pipefail

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"
readonly TMP_DIR=$(mktemp -d)

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    BASE_REF=${PULL_PULL_SHA}
else
    BASE_REF=${PULL_BASE_SHA}
fi

log::info "Execute Job Guard"
/prow-tools/jobguard \
-github-endpoint=http://ghproxy \
-github-endpoint=https://api.github.com \
-github-token-path="/etc/github/token" \
-fail-on-no-contexts="false" \
-timeout="10m" \
-org="kyma-project" \
-repo="kyma-dashboard" \
-base-ref="$BASE_REF" \
-expected-contexts-regexp="(.*-dashboard-dev)"

if [ -n "${PULL_NUMBER}" ]; then
  echo "Building from PR"
  DOCKER_TAG="PR-${PULL_NUMBER}"
else
  # Build artifacts using short SHA for all branches postsubmits
  echo "Building as usual"
  DOCKER_TAG="${PULL_BASE_SHA::8}"
fi

export DOCKER_TAG
echo DOCKER_TAG "${DOCKER_TAG}"

cleanup() {
    
    # do not fail the job regardless of the vm deletion result
    set +e
    
    #shellcheck disable=SC2088
    utils::receive_from_vm "${ZONE}" "kyma-dashboard-test-${RANDOM_ID}" "~/kyma-dashboard-tests/cypress/screenshots" "${ARTIFACTS}"
    #shellcheck disable=SC2088
    utils::receive_from_vm "${ZONE}" "kyma-dashboard-test-${RANDOM_ID}" "~/kyma-dashboard-tests/cypress/videos" "${ARTIFACTS}"
    
    gcloud compute instances delete --zone="${ZONE}" "kyma-dashboard-test-${RANDOM_ID}"
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
    LABELS=(--labels "branch=$PULL_BASE_REF,job-name=kyma-dashboard-test")
else
    LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=kyma-dashboard-test")
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

trap cleanup EXIT HUP INT

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
    log::info "Attempting to create a new instance named busola-smoke-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
    gcloud compute instances create "kyma-dashboard-test-${RANDOM_ID}" \
    --metadata enable-oslogin=TRUE \
    --image "${IMAGE}" \
    --machine-type n2-highcpu-16 \
    --zone "${ZONE}" \
    --boot-disk-size 200 "${LABELS[@]}" && \
    log::info "Created kyma-dashboard-test-${RANDOM_ID} in zone ${ZONE}" && break
    log::error "Could not create machine in zone ${ZONE}"
done || exit 1
ENDTIME=$(date +%s)
echo "VM creation time: $((ENDTIME - STARTTIME)) seconds."

export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"
KYMA_CLUSTER_NAME="nkyma"
log::info "KYMA_CLUSTER_NAME=${KYMA_CLUSTER_NAME}"
kubectl get secrets "${KYMA_CLUSTER_NAME}.kubeconfig" -o jsonpath="{.data.kubeconfig}" | base64 -d > "${TMP_DIR}/kubeconfig-${KYMA_CLUSTER_NAME}.yaml"

log::info "Copying Kyma kubeconfig to the instance"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "kyma-dashboard-test-${RANDOM_ID}" "${TMP_DIR}/kubeconfig-${KYMA_CLUSTER_NAME}.yaml" "~/kubeconfig-kyma.yaml"

log::info "Copying Kyma-Dashboard 'tests' folder to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "kyma-dashboard-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-project/kyma-dashboard/tests" "~/kyma-dashboard-tests"

log::info "Preparing environment variables for the instance"
envVars=(
  DOCKER_TAG
)
utils::save_env_file "${envVars[@]}"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "kyma-dashboard-test-${RANDOM_ID}" ".env" "~/.env"

log::info "Launching the kyma-dashboard-test.sh script"
utils::ssh_to_vm_with_script -z "${ZONE}" -n "kyma-dashboard-test-${RANDOM_ID}" -c "sudo bash" -p "${SCRIPT_DIR}/cluster-integration/kyma-dashboard-test.sh"
log::success "all done"
