#!/bin/bash

# This script is designed to provision a new vm and start serverless chart. It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisining vms. If no flag is provided, the latest custom image is used.

set -o errexit

date

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"
readonly KYMA_PROJECT_DIR="$(cd "${SCRIPT_DIR}/../../../" && pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "$SCRIPT_DIR/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$SCRIPT_DIR/lib/gcp.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    log::info "Execute Job Guard"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

cleanup() {
    ARG=$?
    log::info "Removing instance kyma-integration-test-${RANDOM_ID}"
    gcloud compute instances delete --zone="${ZONE}" "kyma-integration-test-${RANDOM_ID}" || true ### Workaround: not failing the job regardless of the vm deletion result
    date
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
    LABELS=(--labels "branch=$PULL_BASE_REF,job-name=kyma-integration")
else
    LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=kyma-integration")
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

date

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
        --boot-disk-size 200 "${LABELS[@]}" &&\
    log::info "Created kyma-integration-test-${RANDOM_ID} in zone ${ZONE}" && break
    log::error "Could not create machine in zone ${ZONE}"
done || exit 1
ENDTIME=$(date +%s)
echo "VM creation time: $((ENDTIME - STARTTIME)) seconds."

trap cleanup exit INT

# apply overrides if we are not using the default test suite
if [[ ${INTEGRATION_SUITE} == "git-auth-integration" ]]; then
    log::info "Creating Serverless git-auth-integration envs"
    mkdir -p "${KYMA_PROJECT_DIR}/overrides"
    cat <<EOF >> "${KYMA_PROJECT_DIR}/overrides/git-auth.env"
export APP_TEST_GITHUB_SSH_AUTH_KEY="${GH_AUTH_PRIVATE_KEY}"
export APP_TEST_AZURE_BASIC_AUTH_USERNAME="${AZURE_DEVOPS_AUTH_USERNAME}"
export APP_TEST_AZURE_BASIC_AUTH_PASSWORD="${AZURE_DEVOPS_AUTH_PASSWORD}"
EOF
fi

log::info "Copying Kyma to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "${KYMA_PROJECT_DIR}" "~/"

log::info "Triggering the installation"
utils::ssh_to_vm_with_script -z "${ZONE}" -n "kyma-integration-test-${RANDOM_ID}" -c "sudo bash ~/test-infra/prow/scripts/cluster-integration/kyma-serverless-integration-k3d.sh ${INTEGRATION_SUITE}"

log::success "all done"
