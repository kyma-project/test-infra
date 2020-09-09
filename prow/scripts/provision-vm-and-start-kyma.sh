#!/bin/bash

# This script is designed to provision a new vm and start kyma.It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisining vms. If no flag is provided, the latest custom image is used.

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"
readonly TMP_DIR=$(mktemp -d)
readonly JUNIT_REPORT_PATH="${ARTIFACTS:-${TMP_DIR}}/junit_kyma_octopus-test-suite.xml"

# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    shout "Execute Job Guard"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

cleanup() {
    ARG=$?
    shout "Removing instance kyma-integration-test-${RANDOM_ID}"
    gcloud compute instances delete --zone="${ZONE}" "kyma-integration-test-${RANDOM_ID}" || true ### Workaround: not failing the job regardless of the vm deletion result
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
    shout "Attempting to create a new instance named kyma-integration-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
    gcloud compute instances create "kyma-integration-test-${RANDOM_ID}" \
        --metadata enable-oslogin=TRUE \
        --image "${IMAGE}" \
        --machine-type n1-standard-4 \
        --zone "${ZONE}" \
        --boot-disk-size 30 "${LABELS[@]}" &&\
    shout "Created kyma-integration-test-${RANDOM_ID} in zone ${ZONE}" && break
    shout "Could not create machine in zone ${ZONE}"
done || exit 1

trap cleanup exit INT

shout "Copying Kyma to the instance"

for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && echo 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute scp --quiet --recurse --zone="${ZONE}" /home/prow/go/src/github.com/kyma-project/kyma "kyma-integration-test-${RANDOM_ID}":~/kyma && break;
    [[ ${i} -ge 5 ]] && echo "Failed after $i attempts." && exit 1
done;

# FIXME (@Ressetkk): Remove this part when the kyma CLI 1.16.0 is released.
shout "WORKAROUND: Compile kyma-cli from source and push it to the VM"
git clone https://github.com/kyma-project/cli.git &> /dev/null
pushd "./cli"
echo " Compiling Kyma CLI"
make build-linux &> /dev/null
code=$?
if [[ $code != 0 ]]; then
  echo "X An error occured during build."
  exit 1
fi
echo " SCP binary file to a VM"
gcloud compute scp --quiet --recurse --zone="${ZONE}" ./bin/kyma-linux "kyma-integration-test-${RANDOM_ID}":~/kyma-linux
gcloud compute ssh --quiet --zone="${ZONE}" "kyma-integration-test-${RANDOM_ID}" -- "sudo mv ./kyma-linux /usr/local/bin/kyma && sudo chmod +x /usr/local/bin/kyma"
popd
echo " Done"
#----

shout "Triggering the installation"
gcloud compute ssh --quiet --zone="${ZONE}" --command="sudo bash" "kyma-integration-test-${RANDOM_ID}" < "${SCRIPT_DIR}/cluster-integration/kyma-integration-minikube.sh"

shout "Fetch JUnit test results and store them in job artifacts"
gcloud compute scp --quiet --zone="${ZONE}" "kyma-integration-test-${RANDOM_ID}:junit_kyma_octopus-test-suite.xml" "${JUNIT_REPORT_PATH}"

shout "all done"
