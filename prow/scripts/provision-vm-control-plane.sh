#!/usr/bin/env bash

# This script is designed to provision a new vm and start kyma with control-plane. It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisining vms. If no flag is provided, the latest custom image is used.

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"
KYMA_PROJECT_DIR=${KYMA_PROJECT_DIR:-"/home/prow/go/src/github.com/kyma-project"}

# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${SCRIPT_DIR}/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$SCRIPT_DIR/lib/gcp.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    log::info "Execute Job Guard"
    export JOB_NAME_PATTERN="(pre-control-plane-components-.*)|(pre-control-plane-tests-.*)"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

cleanup() {
    ARG=$?
    log::info "Removing instance control-plane-integration-test-${RANDOM_ID}"
    gcloud compute instances delete --zone="${ZONE}" "control-plane-integration-test-${RANDOM_ID}" || true ### Workaround: not failing the job regardless of the vm deletion result
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
    LABELS=(--labels "branch=$PULL_BASE_REF,job-name=control-plane-integration")
else
    LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=control-plane-integration")
fi

POSITIONAL=()
while [[ $# -gt 0 ]]
do

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

for ZONE in ${EU_ZONES}; do
    log::info "Attempting to create a new instance named control-plane-integration-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
    gcloud compute instances create "control-plane-integration-test-${RANDOM_ID}" \
        --metadata enable-oslogin=TRUE \
        --image "${IMAGE}" \
        --machine-type n1-standard-4 \
        --zone "${ZONE}" \
        --boot-disk-size 200 "${LABELS[@]}" &&\
    log::info "Created control-plane-integration-test-${RANDOM_ID} in zone ${ZONE}" && break
    log::error "Could not create machine in zone ${ZONE}"
done || exit 1

trap cleanup exit INT

log::info "Copying control-plane to the instance"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "control-plane-integration-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-project/control-plane" "~/control-plane"

log::info "Download stable Kyma CLI"
curl -Lo kyma https://storage.googleapis.com/kyma-cli-stable/kyma-linux
chmod +x kyma

gcloud compute ssh --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" --quiet --zone="${ZONE}" "control-plane-integration-test-${RANDOM_ID}" -- "mkdir \$HOME/bin"

#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "control-plane-integration-test-${RANDOM_ID}" "kyma" "~/bin/kyma"

gcloud compute ssh --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" --quiet --zone="${ZONE}" "control-plane-integration-test-${RANDOM_ID}" -- "sudo cp \$HOME/bin/kyma /usr/local/bin/kyma"

log::info "Triggering the installation"

gcloud compute ssh --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" --quiet --zone="${ZONE}" "control-plane-integration-test-${RANDOM_ID}" -- "yes | ./control-plane/installation/scripts/prow/deploy-and-test.sh"

log::info "Copying test artifacts from VM"
utils::receive_from_vm "${ZONE}" "control-plane-integration-test-${RANDOM_ID}" "/var/log/prow_artifacts" "${ARTIFACTS}"
