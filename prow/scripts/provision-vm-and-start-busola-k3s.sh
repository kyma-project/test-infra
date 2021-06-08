#!/bin/bash

# This script is designed to provision a new vm and start kyma.It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisining vms. If no flag is provided, the latest custom image is used.

set -o errexit
set -o pipefail

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"
readonly TMP_DIR=$(mktemp -d)


# shellcheck source=prow/scripts/lib/gcloud.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcloud.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    log::info "Execute Job Guard"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

cleanup() {
    
    # do not fail the job regardless of the vm deletion result
    set +e
    
    #shellcheck disable=SC2088
    utils::receive_from_vm "${ZONE}" "busola-ui-test-${RANDOM_ID}" "~/busola-tests/cypress/videos" "${ARTIFACTS}"
    
    gcloud compute instances stop --async --zone="${ZONE}" "busola-ui-test-${RANDOM_ID}"
    log::info "End of cleanup"
}

function testCustomImage() {
    CUSTOM_IMAGE="$1"
    IMAGE_EXISTS=$(gcloud compute images list --filter "name:${CUSTOM_IMAGE}" | tail -n +2 | awk '{print $1}')
    if [[ -z "$IMAGE_EXISTS" ]]; then
        log::error "${CUSTOM_IMAGE} is invalid, it is not available in GCP images list, the script will terminate ..." && exit 1
    fi
}

gcloud::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"

RANDOM_ID=$(openssl rand -hex 4)

LABELS=""
if [[ -z "${PULL_NUMBER}" ]]; then
    LABELS=(--labels "branch=$PULL_BASE_REF,job-name=busola-ui-test")
else
    LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=busola-ui-test")
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
    log::info "Attempting to create a new instance named busola-ui-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
    gcloud compute instances create "busola-ui-test-${RANDOM_ID}" \
    --metadata enable-oslogin=TRUE \
    --image "${IMAGE}" \
    --machine-type n2-standard-4 \
    --zone "${ZONE}" \
    --boot-disk-size 200 "${LABELS[@]}" && \
    log::info "Created busola-ui-test-${RANDOM_ID} in zone ${ZONE}" && break
    log::error "Could not create machine in zone ${ZONE}"
done || exit 1
ENDTIME=$(date +%s)
echo "VM creation time: $((ENDTIME - STARTTIME)) seconds."

trap cleanup exit INT

export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"
KYMA_CLUSTER_NAME="nkyma"
log::info "KYMA_CLUSTER_NAME=${KYMA_CLUSTER_NAME}"
kubectl get secrets "${KYMA_CLUSTER_NAME}.kubeconfig" -o jsonpath="{.data.kubeconfig}" | base64 -d > "${TMP_DIR}/kubeconfig-${KYMA_CLUSTER_NAME}.yaml"

log::info "Copying Kyma kubeconfig to the instance"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "busola-ui-test-${RANDOM_ID}" "${TMP_DIR}/kubeconfig-${KYMA_CLUSTER_NAME}.yaml" "~/kubeconfig-kyma.yaml"

log::info "Copying Busola 'tests' folder to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "busola-ui-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-project/busola/tests" "~/busola-tests"

log::info "Copying Busola 'resources' folder to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "busola-ui-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-project/busola/resources" "~/busola-resources"


log::info "Copying Kyma-Local to the instance"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "busola-ui-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-incubator/local-kyma" "~/local-kyma"


log::info "Launching the busola-ui-tests-k3s.sh script"
gcloud compute ssh --quiet --zone="${ZONE}" --command="sudo bash" --ssh-flag="-o ServerAliveInterval=30" "busola-ui-test-${RANDOM_ID}" < "${SCRIPT_DIR}/cluster-integration/busola-ui-tests-k3s.sh"

log::success "all done"
