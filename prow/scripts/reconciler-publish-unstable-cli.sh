#!/bin/bash

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_PROJECT_DIR=${KYMA_PROJECT_DIR:-"/home/prow/go/src/github.com/kyma-project"}

# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${SCRIPT_DIR}/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$SCRIPT_DIR/lib/gcp.sh"

cleanup() {
    ARG=$?
    log::info "Removing instance cli-integration-test-${RANDOM_ID}"
    gcloud compute instances delete --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" || true ### Workaround: not failing the job regardless of the vm deletion result
    exit $ARG
}

function testCustomImage() {
    CUSTOM_IMAGE="$1"
    log::info "Test custom image: ${CUSTOM_IMAGE}"
    IMAGE_EXISTS=$(gcloud compute images list --filter "name:${CUSTOM_IMAGE}" | tail -n +2 | awk '{print $1}')
    if [[ -z "$IMAGE_EXISTS" ]]; then
        log::error "${CUSTOM_IMAGE} is invalid, it is not available in GCP images list, the script will terminate ..." && exit 1
    fi
}

cd "${KYMA_PROJECT_DIR}/cli"

log::info "Bump reconciler version used by the Kyma CLI"
go get github.com/kyma-incubator/reconciler

make resolve
log::info "Run unit-tests for kyma kyma"
make test
log::info "Building Kyma CLI"
make build-linux

log::info "Committing reconciler bump"
git_status=$(git status --porcelain)
if [[ "${git_status}" != "" ]]; then
  git commit -am 'bump reconciler version'
fi

log::info "GCP Authentication"
gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"

RANDOM_ID=$(openssl rand -hex 4)

LABELS=""
if [[ -z "${PULL_NUMBER}" ]]; then
    LABELS=(--labels "branch=${PULL_BASE_REF/./-},job-name=cli-integration")
else
    LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=cli-integration")
fi

label_log="Labels for gcloud: "
for label in "${LABELS[@]}"
do
  label_log="${label_log} ${label}"
done
log::info "${label_log}"

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

for ZONE in ${EU_ZONES}; do
    log::info "Attempting to create a new instance named cli-integration-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
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

log::info "Wait for 30sec, to let VM start correctly"
sleep 30s
log::info "Create bin directory on VM"
gcloud compute ssh \
  --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" \
  --verbosity="${GCLOUD_SSH_LOG_LEVEL:-error}" \
  --quiet --zone="${ZONE}" \
  "cli-integration-test-${RANDOM_ID}" \
  --command="mkdir \$HOME/bin"

log::info "Copying Kyma CLI to the instance"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "cli-integration-test-${RANDOM_ID}" "${KYMA_PROJECT_DIR}/cli/bin/kyma-linux" "~/bin/kyma"
gcloud compute ssh \
  --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" \
  --verbosity="${GCLOUD_SSH_LOG_LEVEL:-error}" \
  --quiet \
  --zone="${ZONE}" \
  "cli-integration-test-${RANDOM_ID}" \
  --command="sudo cp \$HOME/bin/kyma /usr/local/bin/kyma"

log::info "Provisioning k3d Kubernetes runtime"
gcloud compute ssh \
  --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" \
  --verbosity="${GCLOUD_SSH_LOG_LEVEL:-error}" \
  --quiet \
  --zone="${ZONE}" \
  "cli-integration-test-${RANDOM_ID}" \
  --command="yes | sudo kyma provision k3d --ci"

log::info "Installing Kyma"
gcloud compute ssh \
  --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" \
  --verbosity="${GCLOUD_SSH_LOG_LEVEL:-error}" \
  --quiet \
  --zone="${ZONE}" \
  "cli-integration-test-${RANDOM_ID}" \
  --command="yes | sudo kyma deploy --ci ${SOURCE}"

log::info "Copying Kyma to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "cli-integration-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-project/kyma" "~/kyma"

log::info "Running fast-integration tests"
gcloud compute ssh\
  --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" \
  --verbosity="${GCLOUD_SSH_LOG_LEVEL:-error}" \
  --quiet \
  --zone="${ZONE}" \
  --ssh-flag="-o ServerAliveInterval=30" \
  "cli-integration-test-${RANDOM_ID}" \
  --command="cd ~/kyma/tests/fast-integration && sudo make ci"

log::info "Uninstalling Kyma"
gcloud compute ssh \
  --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" \
  --verbosity="${GCLOUD_SSH_LOG_LEVEL:-error}" \
  --quiet \
  --zone="${ZONE}" \
  "cli-integration-test-${RANDOM_ID}" \
  --command="sudo kyma undeploy --ci --timeout=10m0s"

log::info "Publishing new unstable builds to $KYMA_CLI_UNSTABLE_BUCKET"
make ci-main

log::success "all done"
