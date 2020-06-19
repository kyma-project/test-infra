#!/usr/bin/env bash

# This script is designed to provision a new vm and start kyma with compass. It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisining vms. If no flag is provided, the latest custom image is used.

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"
KYMA_PROJECT_DIR=${KYMA_PROJECT_DIR:-"/home/prow/go/src/github.com/kyma-project"}

# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    shout "Execute Job Guard"
    export JOB_NAME_PATTERN="(pre-compass-components-.*)|(pre-compass-tests-.*)"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

cleanup() {
    ARG=$?
    shout "Removing instance compass-integration-test-${RANDOM_ID}"
    gcloud compute instances delete --zone="${ZONE}" "compass-integration-test-${RANDOM_ID}" || true ### Workaround: not failing the job regardless of the vm deletion result
    exit $ARG
}

function testCustomImage() {
    CUSTOM_IMAGE="$1"
    IMAGE_EXISTS=$(gcloud compute images list --filter "name:${CUSTOM_IMAGE}" | tail -n +2 | awk '{print $1}')
    if [[ -z "$IMAGE_EXISTS" ]]; then
        shout "${CUSTOM_IMAGE} is invalid, it is not available in GCP images list, the script will terminate ..." && exit 1
    fi
}

authenticate

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
    shout "Provisioning vm using the latest default custom image ..."   
    
    IMAGE=$(gcloud compute images list --sort-by "~creationTimestamp" \
         --filter "family:custom images AND labels.default:yes" --limit=1 | tail -n +2 | awk '{print $1}')
    
    if [[ -z "$IMAGE" ]]; then
       shout "There are no default custom images, the script will exit ..." && exit 1 
    fi   
 fi

ZONE_LIMIT=${ZONE_LIMIT:-5}
EU_ZONES=$(gcloud compute zones list --filter="name~europe" --limit="${ZONE_LIMIT}" | tail -n +2 | awk '{print $1}')

for ZONE in ${EU_ZONES}; do
    shout "Attempting to create a new instance named compass-integration-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
    gcloud compute instances create "compass-integration-test-${RANDOM_ID}" \
        --metadata enable-oslogin=TRUE \
        --image "${IMAGE}" \
        --machine-type n1-standard-4 \
        --zone "${ZONE}" \
        --boot-disk-size 30 "${LABELS[@]}" &&\
    shout "Created compass-integration-test-${RANDOM_ID} in zone ${ZONE}" && break
    shout "Could not create machine in zone ${ZONE}"
done || exit 1

trap cleanup exit INT

shout "Copying Compass to the instance"

for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && echo 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute scp --quiet --recurse --zone="${ZONE}" /home/prow/go/src/github.com/kyma-incubator/compass "compass-integration-test-${RANDOM_ID}":~/compass && break;
    [[ ${i} -ge 5 ]] && echo "Failed after $i attempts." && exit 1
done;

shout "Download stable Kyma CLI"
curl -Lo kyma https://storage.googleapis.com/kyma-cli-stable/kyma-linux
chmod +x kyma

gcloud compute ssh --quiet --zone="${ZONE}" "compass-integration-test-${RANDOM_ID}" -- "mkdir \$HOME/bin"

for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && echo 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute scp --quiet --zone="${ZONE}" "kyma" "compass-integration-test-${RANDOM_ID}":~/bin/kyma && break;
    [[ ${i} -ge 5 ]] && echo "Failed after $i attempts." && exit 1
done;

gcloud compute ssh --quiet --zone="${ZONE}" "compass-integration-test-${RANDOM_ID}" -- "sudo cp \$HOME/bin/kyma /usr/local/bin/kyma"

shout "Triggering the installation"

gcloud compute ssh --quiet --zone="${ZONE}" "compass-integration-test-${RANDOM_ID}" -- "yes | ./compass/installation/scripts/prow/deploy-and-test.sh"

shout "Copying test artifacts from VM"

for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && echo 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute scp --recurse --zone="${ZONE}" "compass-integration-test-${RANDOM_ID}":/var/log/prow_artifacts "${ARTIFACTS}"  && break;
    # TODO change exit code to 1 later
    [[ ${i} -ge 5 ]] && echo "Failed after $i attempts." && exit 0
done;

