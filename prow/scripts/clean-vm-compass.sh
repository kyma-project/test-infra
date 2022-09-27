#!/usr/bin/env bash

# This script is designed to clean all vms starged from integration tests. 
# It require mandatory parameter --vm-prefix that defines the VM name porefix that is used to select running GCloud images that will be cleaned up.  


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
    log::info "The job should not be executed on PRs"
    exit 1
elif [[ "${BUILD_TYPE}" == "master" ]] ; then
    log::info "Main branch - cleanup should be done."
    log::info "Execute Job Guard"
    export JOB_NAME_PATTERN="(^pre-main-compass-integration$)|(^pre-main-compass-integration-no-dump$)"
    export JOBGUARD_TIMEOUT="30m"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"

RANDOM_ID=$(openssl rand -hex 4)

LABELS=""
if [[ -z "${PULL_NUMBER}" ]]; then
    LABELS=(--labels "branch=$PULL_BASE_REF,job-name=compass-integration")
else
    LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=compass-integration")
fi

POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"

    case ${key} in
        --vm-prefix)
            VM_PREFIX="$2"
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

if [[ -z "$VM_PREFIX" ]]; then
    log::error "There are no --vm-prefix specified, the script will exit ..." && exit 1
fi

log::info "VM Prefix set to: ${VM_PREFIX}"

VMS_RESPONSE=$(gcloud compute instances list --sort-by "~creationTimestamp" --filter="name~${VM_PREFIX}" --format=json)
VMS_COUNT=$(echo -E "${VMS_RESPONSE}" | jq length)
log::info "VMs Found: ${VMS_COUNT}"

OPEN_PRS_RESPONSE=$(curl -sS "https://api.github.com/repos/kyma-incubator/compass/pulls?state=open")
OPEN_PRS_COUNT=$(echo -E "${OPEN_PRS_RESPONSE}" | jq length)
log::info "Open PRs Found: ${OPEN_PRS_COUNT}"

for i in $(echo -E "${VMS_RESPONSE}" | jq -r '.[].name'); do
    CURRENT_VM_NAME="${i}"
    CURRENT_VM_SUFFIX="${CURRENT_VM_NAME##*-}"
    CURRENT_VM_ZONE_URL=$(echo -E "${VMS_RESPONSE}" | jq -r --arg key "${CURRENT_VM_NAME}" '.[] | select(.name == $key).zone')
    CURRENT_VM_ZONE="${CURRENT_VM_ZONE_URL##*/}"

    log::info "VM Name: ${CURRENT_VM_NAME}"
    log::info "VM PR: ${CURRENT_VM_SUFFIX}"
    log::info "VM Zone: ${CURRENT_VM_ZONE}"

    PR_ID_FOR_SUFFIX=$(echo -E "${OPEN_PRS_RESPONSE}" | jq -r --argjson prnumber "${CURRENT_VM_SUFFIX}" '.[] | select(.number == $prnumber) | .number')
    if [[ -z "$PR_ID_FOR_SUFFIX" ]]; then
        log::info "The PR number: ${PR_ID_FOR_SUFFIX} is not open. Cleaning up the VM: ${CURRENT_VM_NAME}" 
        gcloud compute instances delete --zone="${CURRENT_VM_ZONE}" "${CURRENT_VM_NAME}" || true ### Workaround: not failing the job regardless of the vm deletion result
    else
        log::info "The PR number: ${PR_ID_FOR_SUFFIX} is still open. The VM: ${CURRENT_VM_NAME} will not be cleaned." 
    fi

done

log::info "Cleaning VMs complete"