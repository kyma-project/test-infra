#!/usr/bin/env bash

# This script is designed to provision a new vm. It takes an optional positional parameter using --image flag
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
    export JOB_NAME_PATTERN="(pre-compass-components-.*)|(^pre-compass-tests$)"
    export JOBGUARD_TIMEOUT="30m"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

cleanup() {
    ARG=$?
    log::info "Removing instance compass-ord-test-${RANDOM_ID}"
    gcloud compute instances delete --zone="${ZONE}" "compass-ord-test-${RANDOM_ID}" || true ### Workaround: not failing the job regardless of the vm deletion result
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
    LABELS=(--labels "branch=$PULL_BASE_REF,job-name=compass-ord")
else
    LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=compass-ord")
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

echo "IMAGE = ${IMAGE}"

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
    log::info "Attempting to create a new instance named compass-ord-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
    gcloud compute instances create "compass-ord-test-${RANDOM_ID}" \
        --metadata enable-oslogin=TRUE \
        --image "${IMAGE}" \
        --machine-type n1-standard-4 \
        --zone "${ZONE}" \
        --boot-disk-size 200 "${LABELS[@]}" &&\
    log::info "Created compass-ord-test-${RANDOM_ID} in zone ${ZONE}" && break
    log::error "Could not create machine in zone ${ZONE}"
done || exit 1

trap cleanup exit INT

chmod -R 0777 /home/prow/go/src/github.com/kyma-incubator/compass/.git
mkdir -p /home/prow/go/src/github.com/kyma-incubator/compass/components/console/shared/build

log::info "Copying Compass to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "compass-ord-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-incubator/compass" "~/compass"

log::info "Get ORD commit ID"
ORD_PR_NUMBER=$(yq e .global.images.ord_service.version /home/prow/go/src/github.com/kyma-incubator/compass/chart/compass/values.yaml | cut -d '-' -f 2 | xargs)
log::info "ORD_PR_NUMBER PR is: ${ORD_PR_NUMBER}"

ORD_PR_DATA=$(curl -sS "https://api.github.com/repos/kyma-incubator/ord-service/pulls/${ORD_PR_NUMBER}")
log::info "ORD_PR_DATA is: ${ORD_PR_DATA}"

ORD_PR_COMMIT_HASH=$(jq -r '.merge_commit_sha' <<< "${ORD_PR_DATA}")
log::info "ORD_PR_COMMIT_HASH is: ${ORD_PR_COMMIT_HASH}"

log::info "Fetch ORD service sources"
git clone https://github.com/kyma-incubator/ord-service.git && cd ord-service && git checkout "${ORD_PR_COMMIT_HASH}" && cd ..

log::info "Copying ORD service sources to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "compass-ord-test-${RANDOM_ID}" "./ord-service" "~/ord-service"

log::info "Triggering the test"

gcloud compute ssh --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" --verbosity="${GCLOUD_SSH_LOG_LEVEL:-error}" --quiet --zone="${ZONE}" "compass-ord-test-${RANDOM_ID}" --command="yes | ./compass/installation/scripts/prow/ord-test.sh"

log::info "Copying test artifacts from VM"
utils::receive_from_vm "${ZONE}" "compass-ord-test-${RANDOM_ID}" "/var/log/prow_artifacts" "${ARTIFACTS}"
