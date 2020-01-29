#!/usr/bin/env bash

# This script is designed to provision a new vm and start kyma with cli. It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisioning vms. If no flag is provided, the latest custom image is used.

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"
KYMA_PROJECT_DIR=${KYMA_PROJECT_DIR:-"/home/prow/go/src/github.com/kyma-project"}

# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

cleanup() {
    ARG=$?
    shout "Removing instance cli-integration-test-${RANDOM_ID}"
    date
    gcloud compute instances delete --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" || true ### Workaround: not failing the job regardless of the vm deletion result
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
    LABELS=(--labels "branch=$PULL_BASE_REF,job-name=cli-integration")
else
    LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=cli-integration")
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
    date
    IMAGE=$(gcloud compute images list --sort-by "~creationTimestamp" \
         --filter "family:custom images AND labels.default:yes" --limit=1 | tail -n +2 | awk '{print $1}')

    if [[ -z "$IMAGE" ]]; then
       shout "There are no default custom images, the script will exit ..." && exit 1
    fi
 fi

ZONE_LIMIT=${ZONE_LIMIT:-5}
EU_ZONES=$(gcloud compute zones list --filter="name~europe" --limit="${ZONE_LIMIT}" | tail -n +2 | awk '{print $1}')

for ZONE in ${EU_ZONES}; do
    shout "Attempting to create a new instance named cli-integration-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
    date
    gcloud compute instances create "cli-integration-test-${RANDOM_ID}" \
        --metadata enable-oslogin=TRUE \
        --image "${IMAGE}" \
        --machine-type n1-standard-4 \
        --zone "${ZONE}" \
        --boot-disk-size 30 "${LABELS[@]}" &&\
    shout "Created cli-integration-test-${RANDOM_ID} in zone ${ZONE}" && break
    shout "Could not create machine in zone ${ZONE}"
done || exit 1

trap cleanup exit INT

shout "Building Kyma CLI"
date
cd "${KYMA_PROJECT_DIR}/cli"
make build-linux
gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "mkdir \$HOME/bin"

shout "Copying Kyma CLI to the instance"
date
for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && echo 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute scp --quiet --zone="${ZONE}" "${KYMA_PROJECT_DIR}/cli/bin/kyma-linux" "cli-integration-test-${RANDOM_ID}":~/bin/kyma && break;
    [[ ${i} -ge 5 ]] && echo "Failed after $i attempts." && exit 1
done;
gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "sudo cp \$HOME/bin/kyma /usr/local/bin/kyma"

shout "Provisioning Minikube"
date
gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "yes | sudo kyma provision minikube --non-interactive"

shout "Installing Kyma"
date
gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "yes | sudo kyma install --non-interactive"

shout "Checking the versions"
date
gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "sudo kyma version"

shout "Running a simple test on Kyma"
date
gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "sudo kyma test run dex-connection"

echo "Check if the test succeeds"
date
attempts=3
for ((i=1; i<=attempts; i++)); do
    result=$(gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "sudo kyma test status -o json" | jq '.status.results[0].status')
    if [[ "$result" == *"Succeeded"* ]]; then
        echo "The test succeeded"
        break
    elif [[ "${i}" == "${attempts}" ]]; then
        echo "ERROR: test result is ${result}"
        exit 1
    fi
    echo "Sleep for 15 seconds"
    sleep 15
done
