#!/bin/bash

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

cleanup() {
    ARG=$?
    shout "Removing instance kyma-integration-test-${RANDOM_ID}"
    gcloud compute instances delete --zone="${ZONE}" "kyma-integration-test-${RANDOM_ID}"
    exit $ARG
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
    
    DEFAULT_IMAGES=$(gcloud compute images list --sort-by "~creationTimestamp" \
         --filter "family:custom images AND labels.default:yes" | tail -n +2 | awk '{print $1}')

    default_image_arr=($DEFAULT_IMAGES)
    IMAGE=${default_image_arr[0]}
    
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
        --boot-disk-size 20 "${LABELS[@]}" &&\
    shout "Created kyma-integration-test-${RANDOM_ID} in zone ${ZONE}" && break
    shout "Could not create machine in zone ${ZONE}"
done || exit 1

trap cleanup exit

shout "Copying Kyma to the instance"

for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && echo 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute scp --quiet --recurse --zone="${ZONE}" /home/prow/go/src/github.com/kyma-project/kyma "kyma-integration-test-${RANDOM_ID}":~/kyma && break;
    [[ ${i} -ge 5 ]] && echo "Failed after $i attempts." && exit 1
done;

shout "Triggering the installation"

gcloud compute ssh --quiet --zone="${ZONE}" "kyma-integration-test-${RANDOM_ID}" -- ./kyma/installation/scripts/prow/kyma-integration-on-debian/deploy-and-test-kyma.sh
