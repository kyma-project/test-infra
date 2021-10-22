#!/usr/bin/env bash

# This script is designed to provision a new vm and start kyma with cli. It takes the following optional positional parameters:
# custom VM image --image flag: Use this flag to specify the custom image for provisioning vms. If no flag is provided, the latest custom image is used.
# Kyma version to install --kyma-version flag: Use this flag to indicate which Kyma version the CLI should install (default: same as the CLI)

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_PROJECT_DIR=${KYMA_PROJECT_DIR:-"/home/prow/go/src/github.com/kyma-project"}

# shellcheck source=prow/scripts/lib/gcloud.sh
source "${SCRIPT_DIR}/lib/gcloud.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${SCRIPT_DIR}/lib/utils.sh"

cleanup() {
    ARG=$?
    log::info "Removing instance cli-integration-test-${RANDOM_ID}"
    date
    gcloud compute instances delete --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" || true ### Workaround: not failing the job regardless of the vm deletion result
    exit $ARG
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
    LABELS=(--labels "branch=${PULL_BASE_REF/./-},job-name=cli-integration")
else
    LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=cli-integration")
fi

# Support configuration via ENV vars (can be be overwritten by CLI args)
KUBERNETES_RUNTIME="${KUBERNETES_RUNTIME:=minikube}"
# Either use the default Kyma install command or the new alpha command.
INSTALLATION="${INSTALLATION:=default}"

POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"

    case ${key} in
        --image|-i)
            IMAGE="$2"
            testCustomImage "${IMAGE}"
            shift 2
            ;;
        --kyma-version|-kv)
            SOURCE="--source $2"
            shift 2
            ;;
        --kubernetes-runtime|-kr)
            KUBERNETES_RUNTIME="$2"
            shift 2
            ;;
        --installation)
            INSTALLATION="$2"
            shift 2
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
    # Newer versions use k3d v5, we want older version here
    IMAGE="kyma-deps-image-v20210930-03ea1d"
 fi

ZONE_LIMIT=${ZONE_LIMIT:-5}
EU_ZONES=$(gcloud compute zones list --filter="name~europe" --limit="${ZONE_LIMIT}" | tail -n +2 | awk '{print $1}')

for ZONE in ${EU_ZONES}; do
    log::info "Attempting to create a new instance named cli-integration-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
    date
    gcloud compute instances create "cli-integration-test-${RANDOM_ID}" \
        --metadata enable-oslogin=TRUE \
        --image "${IMAGE}" \
        --machine-type n1-standard-4 \
        --zone "${ZONE}" \
        --boot-disk-size 200 "${LABELS[@]}" &&\
    log::info "Created cli-integration-test-${RANDOM_ID} in zone ${ZONE}" && break
    log::error "Could not create machine in zone ${ZONE}"
done || exit 1

trap cleanup exit INT

log::info "Building Kyma CLI"
date
cd "${KYMA_PROJECT_DIR}/cli"
make build-linux
gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "mkdir \$HOME/bin"

log::info "Copying Kyma CLI to the instance"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "cli-integration-test-${RANDOM_ID}" "${KYMA_PROJECT_DIR}/cli/bin/kyma-linux" "~/bin/kyma"
gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "sudo cp \$HOME/bin/kyma /usr/local/bin/kyma"

# Provision Kubernetes runtime
log::info "Provisioning Kubernetes runtime '$KUBERNETES_RUNTIME'"
date
if [ "$KUBERNETES_RUNTIME" = 'minikube' ]; then
    gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "yes | sudo kyma provision minikube --non-interactive"
else
    # this line is used to fix k3s provisioning, see https://github.com/kyma-project/test-infra/issues/4033
    gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "sudo sysctl -w net/netfilter/nf_conntrack_max=131072"
    gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "yes | sudo kyma alpha provision k3s --ci"
fi

# Install kyma
log::info "Installing Kyma"
date
if [ "$INSTALLATION" = 'alpha' ]; then
    gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "yes | sudo kyma alpha deploy --ci ${SOURCE}"
else
    gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "yes | sudo kyma install --non-interactive ${SOURCE}"
fi

# Run test suite
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/lib/clitests.sh"

# ON alpha installation there is no dex, therefore skipping the test
if [ "$INSTALLATION" = 'alpha' ]; then
    if clitests::testSuiteExists "test-version"; then
        clitests::execute "test-version" "${ZONE}" "cli-integration-test-${RANDOM_ID}" "$SOURCE"
    else
        log::error "Test file 'test-version.sh' not found"
    fi
    if clitests::testSuiteExists "test-function"; then
        clitests::execute "test-function" "${ZONE}" "cli-integration-test-${RANDOM_ID}" "$SOURCE"
    else
        log::error "Test file 'test-function.sh' not found"
    fi
else
    if clitests::testSuiteExists "test-all"; then
        clitests::execute "test-all" "${ZONE}" "cli-integration-test-${RANDOM_ID}" "$SOURCE"
    else
        log::error "Test file 'test-all.sh' not found"
    fi
fi
