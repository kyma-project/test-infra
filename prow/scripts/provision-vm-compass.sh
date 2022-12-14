#!/usr/bin/env bash

# This script is designed to provision a new vm and start kyma with compass. It takes an optional positional parameter using --image flag
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

 # Download yq
YQ_VERSION="v4.25.1"
log::info "Downloading yq version: $YQ_VERSION"
curl -fsSL "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64" -o yq
chmod +x yq

# Flag to recycle or recreate VM on each commit execution
RECYCLE_VM="no"
# Flag that shows if the mandatory provisioning (kyma installation) was completed
IS_KYMA_INSTALLED="no"

get_schema_migrator_version() {
    local PATH_TO_VALUES_YAML="/home/prow/go/src/github.com/kyma-incubator/compass/chart/compass/values.yaml"
    local SCHEMA_MIGRATOR_VERSION=$( yq eval '.global.images.schema_migrator.version' "${PATH_TO_VALUES_YAML}")
    echo "${SCHEMA_MIGRATOR_VERSION}"
}

cleanup() {
    local ARG=$?
    if [[ "${RECYCLE_VM}" == "yes" && "${IS_KYMA_INSTALLED}" == "yes" ]]; then
        log::info "Triggering the compass uninstallation"
        utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_PREFIX}${SUFFIX}" -c "yes | ~/compass/installation/scripts/prow/uninstall-compass.sh"
    else
        log::info "Removing instance ${VM_PREFIX}${SUFFIX}"
        gcloud compute instances delete --zone="${ZONE}" "${VM_PREFIX}${SUFFIX}" || true ### Workaround: not failing the job regardless of the vm deletion result
    fi
    exit $ARG
}

LABELS=""

if [[ -z "${PULL_NUMBER}" ]]; then
    LABELS=(--labels "branch=${PULL_BASE_REF},job-name=compass-integration")
    SUFFIX="${PULL_BASE_REF}"
else
    LABELS=(--labels "pull-number=${PULL_NUMBER},job-name=compass-integration")
    SUFFIX="${PULL_NUMBER}"
fi

read -r SCHEMA_MIGRATOR_VERSION <<< "$( get_schema_migrator_version )"
if [[ "${BUILD_TYPE}" == "pr" ]]; then
    log::info "Execute Job Guard"
    export JOB_NAME_PATTERN="(pull-.*)"
    export JOBGUARD_TIMEOUT="60m"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"

    CURRENT_PR_LABEL="PR-${PULL_NUMBER}"
    if [[ "${SCHEMA_MIGRATOR_VERSION}" != "${CURRENT_PR_LABEL}" ]]; then
        # Recycle VM only on PR execution and when schema migrator is not changed as part of this PR
        RECYCLE_VM="yes"
    fi
fi

gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"


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
        --dump-db)
            DUMP_DB="--dump-db"
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

VM_PREFIX="compass-integration-test-no-dump-"
if [[ ${DUMP_DB} ]]; then
    VM_PREFIX="compass-integration-test-with-dump-"
fi

# Check if the VM need to be recreated - default is no
RECREATE_VM_BEFORE_START="no"
if [[ -z "${PULL_NUMBER}" ]]; then
    # VM is recreated always when job is executed on main branch
    RECREATE_VM_BEFORE_START="yes"
else
    # On PR VM is recreated only when requested with comment
    FIVE_MINUTES_AGO=$(date -u +"%Y-%m-%dT%H:%M:%SZ" --date='5 minutes ago')
    PR_COMMENTS_RESPONSE=$(curl -sS "https://api.github.com/repos/kyma-incubator/compass/issues/${PULL_NUMBER}/comments?since=${FIVE_MINUTES_AGO}")
    COMMENTS_COUNT=$(echo -E "${PR_COMMENTS_RESPONSE}" | jq length)
    if [ $? -ne 0 ]; then
        log::error "The PRs Comments response cannot be parsed: ${PR_COMMENTS_RESPONSE}" && exit 1
    fi

    for i in $(echo -E "${PR_COMMENTS_RESPONSE}" | jq -r '.[].url'); do
        COMMENT_URL="${i}"
        COMMENT_BODY=$(echo -E "${PR_COMMENTS_RESPONSE}" | jq -r --arg key "${COMMENT_URL}" '.[] | select(.url == $key).body')

        if [[ "${COMMENT_BODY}" =~ "reset-vm" ]]; then
            RECREATE_VM_BEFORE_START="yes"
        fi
    done
fi


VMS_RESPONSE=$(gcloud compute instances list --sort-by "~creationTimestamp" --filter="name~${VM_PREFIX}" --format=json)
VM_FOR_PREFIX_AND_SUFFIX=$(echo -E "${VMS_RESPONSE}" | jq -r --arg vmname "${VM_PREFIX}${SUFFIX}" '.[] | select(.name == $vmname) | .name')

if [[ -n "${VM_FOR_PREFIX_AND_SUFFIX}" && "${RECREATE_VM_BEFORE_START}" == "yes" ]]; then
    VM_ZONE_URL=$(echo -E "${VMS_RESPONSE}" | jq -r --arg key "${VM_FOR_PREFIX_AND_SUFFIX}" '.[] | select(.name == $key).zone')
    ZONE="${VM_ZONE_URL##*/}"

    log::info "Removing VM ${VM_PREFIX}${SUFFIX} as it was requested to be reset"
    gcloud compute instances delete --zone="${ZONE}" "${VM_PREFIX}${SUFFIX}" || true ### Workaround: not failing the job regardless of the vm deletion result
    
    # Refetch the VMs information again
    VMS_RESPONSE=$(gcloud compute instances list --sort-by "~creationTimestamp" --filter="name~${VM_PREFIX}" --format=json)
    VM_FOR_PREFIX_AND_SUFFIX=$(echo -E "${VMS_RESPONSE}" | jq -r --arg vmname "${VM_PREFIX}${SUFFIX}" '.[] | select(.name == $vmname) | .name')
fi

if [[ -z "${VM_FOR_PREFIX_AND_SUFFIX}" ]]; then
    log::info "The VM with name:  ${VM_PREFIX}${SUFFIX} is missing - will be created..." 

    if [[ -z "${IMAGE}" ]]; then
        log::info "Provisioning vm using the latest default custom image ..."
        
        IMAGE=$(gcloud compute images list --sort-by "~creationTimestamp" \
            --filter "family:custom images AND labels.default:yes" --limit=1 | tail -n +2 | awk '{print $1}')
        
        if [[ -z "${IMAGE}" ]]; then
        log::error "There are no default custom images, the script will exit ..." && exit 1
        fi   
    fi

    ZONE_LIMIT=${ZONE_LIMIT:-5}
    EU_ZONES=$(gcloud compute zones list --filter="name~europe" --limit="${ZONE_LIMIT}" | tail -n +2 | awk '{print $1}')

    for ZONE in ${EU_ZONES}; do
        log::info "Attempting to create a new instance named ${VM_PREFIX}${SUFFIX} in zone ${ZONE} using image ${IMAGE}"
        gcloud compute instances create "${VM_PREFIX}${SUFFIX}" \
            --metadata enable-oslogin=TRUE \
            --image "${IMAGE}" \
            --machine-type n1-standard-8 \
            --zone "${ZONE}" \
            --boot-disk-size 200 "${LABELS[@]}" &&\
        log::info "Created ${VM_PREFIX}${SUFFIX} in zone ${ZONE}" && break
        log::error "Could not create machine in zone ${ZONE}"
    done || exit 1

    trap cleanup exit INT

    chmod -R 0777 /home/prow/go/src/github.com/kyma-incubator/compass/.git
    mkdir -p /home/prow/go/src/github.com/kyma-incubator/compass/components/console/shared/build

    log::info "Copying Compass to the instance"
    #shellcheck disable=SC2088
    utils::compress_send_to_vm "${ZONE}" "${VM_PREFIX}${SUFFIX}" "/home/prow/go/src/github.com/kyma-incubator/compass" "~/compass"


    KYMA_CLI_VERSION="2.2.0"
    log::info "Installing Kyma CLI version: $KYMA_CLI_VERSION"

    PREV_WD=$(pwd)
    git clone https://github.com/kyma-project/cli.git && cd cli && git checkout $KYMA_CLI_VERSION
    make build-linux && cd ./bin && mv ./kyma-linux ./kyma
    chmod +x kyma

    utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_PREFIX}${SUFFIX}" -c "mkdir \$HOME/bin"

    #shellcheck disable=SC2088
    utils::send_to_vm "${ZONE}" "${VM_PREFIX}${SUFFIX}" "kyma" "~/bin/kyma"

    utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_PREFIX}${SUFFIX}" -c "sudo cp \$HOME/bin/kyma /usr/local/bin/kyma"

    cd "$PREV_WD"
    log::info "Successfully installed Kyma CLI version: $KYMA_CLI_VERSION"

    log::info "Installing yq version to VM: $YQ_VERSION"
    #shellcheck disable=SC2088
    utils::send_to_vm "${ZONE}" "${VM_PREFIX}${SUFFIX}" "yq" "~/bin/yq"
    utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_PREFIX}${SUFFIX}" -c "sudo cp \$HOME/bin/yq /usr/local/bin/yq"

    log::info "Triggering the full installation"
    utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_PREFIX}${SUFFIX}" -c "yes | ~/compass/installation/scripts/prow/provision.sh ${DUMP_DB}"
    log::info "Full provisioning done"
    IS_KYMA_INSTALLED="yes"
else
    IS_KYMA_INSTALLED="yes"
    trap cleanup exit INT

    log::info "The VM with name:  ${VM_FOR_PREFIX_AND_SUFFIX} is available - will be reused..." 
    VM_ZONE_URL=$(echo -E "${VMS_RESPONSE}" | jq -r --arg key "${VM_FOR_PREFIX_AND_SUFFIX}" '.[] | select(.name == $key).zone')
    ZONE="${VM_ZONE_URL##*/}"

    chmod -R 0777 /home/prow/go/src/github.com/kyma-incubator/compass/.git

    log::info "Clear old Compass sources from VM"
    utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_PREFIX}${SUFFIX}" -c "sudo rm -fR ~/compass"

    log::info "Copying new Compass sources to the VM"
    #shellcheck disable=SC2088
    utils::compress_send_to_vm "${ZONE}" "${VM_PREFIX}${SUFFIX}" "/home/prow/go/src/github.com/kyma-incubator/compass" "~/compass"

    log::info "Triggering the compass only installation"
    utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_FOR_PREFIX_AND_SUFFIX}" -c "yes | ~/compass/installation/scripts/prow/update-compass.sh"
    log::info "Compass only installation done"
fi

log::info "Triggering the tests"

utils::ssh_to_vm_with_script -z "${ZONE}" -n "${VM_PREFIX}${SUFFIX}" -c "yes | ~/compass/installation/scripts/prow/execute-tests.sh ${DUMP_DB}"

log::info "Copying test artifacts from VM"
utils::receive_from_vm "${ZONE}" "${VM_PREFIX}${SUFFIX}" "/var/log/prow_artifacts" "${ARTIFACTS}"

log::info "Test execution completed"