#!/bin/bash

# This script is designed to provision a new vm and start kyma.It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisining vms. If no flag is provided, the latest custom image is used.

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"
export KYMA_SOURCES_DIR="/home/prow/go/src/github.com/kyma-project/kyma"

# TODO: move this to job definition? Less logs from gcloud itself
export GCLOUD_SSH_LOG_LEVEL="error"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
  log::info "Execute Job Guard"
  "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

cleanup() {
  # TODO - collect junit results

  log::info "Stopping instance kyma-integration-test-${RANDOM_ID}"
  log::info "It will be removed automatically by cleaner job"

  # do not fail the job regardless of the vm deletion result
  set +e

  #shellcheck disable=SC2088
  if [[ ! $ISTIO_INTEGRATION_ENABLED ]]; then
    utils::receive_from_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "~/kyma/tests/fast-integration/junit_kyma-fast-integration.xml" "${ARTIFACTS}"
  fi
  
  # gcloud compute instances stop --async --zone="${ZONE}" "kyma-integration-test-${RANDOM_ID}"

  log::info "End of cleanup"
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

export DOCKER_SKIP_GCR_AUTHENTICATION="true"
docker::start

RANDOM_ID=$(openssl rand -hex 4)

LABELS=""
if [[ -z "${PULL_NUMBER}" ]]; then
  LABELS=(--labels "branch=$PULL_BASE_REF,job-name=kyma-integration")
else
  LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=kyma-integration")
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
  log::info "Attempting to create a new instance named kyma-integration-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
  gcloud compute instances create "kyma-integration-test-${RANDOM_ID}" \
      --metadata enable-oslogin=TRUE \
      --image "${IMAGE}" \
      --machine-type n2-standard-4 \
      --zone "${ZONE}" \
      --boot-disk-size 200 "${LABELS[@]}" && \
  log::info "Created kyma-integration-test-${RANDOM_ID} in zone ${ZONE}" && break
  log::error "Could not create machine in zone ${ZONE}"
done || exit 1
ENDTIME=$(date +%s)
echo "VM creation time: $((ENDTIME - STARTTIME)) seconds."
export MACHINE_IP=$(gcloud compute instances describe "kyma-integration-test-${RANDOM_ID}" --zone "${ZONE}" --format='get(networkInterfaces[0].accessConfigs[0].natIP)')

trap cleanup exit INT

log::info "Preparing environment variables for the instance"

envVars=(
  COMPASS_TENANT
  COMPASS_HOST
  COMPASS_CLIENT_ID
  COMPASS_CLIENT_SECRET
  COMPASS_INTEGRATION_ENABLED
  CENTRAL_APPLICATION_CONNECTIVITY_ENABLED
  TELEMETRY_ENABLED
  ISTIO_INTEGRATION_ENABLED
  KYMA_MAJOR_VERSION
  KYMA_PROFILE
  K8S_VERSION
  MACHINE_IP
)
utils::save_env_file "${envVars[@]}"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" ".env" "~/.env"

# log::info "Copying Kyma to the instance"
# #shellcheck disable=SC2088
# utils::compress_send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-project/kyma" "~/kyma"

# if [[ -v COMPASS_INTEGRATION_ENABLED ]]; then
#   log::info "Copying components file for compass tests"
#   #shellcheck disable=SC2088
#   utils::send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d-compass-components.yaml" "~/kyma-integration-k3d-compass-components.yaml"
# fi

# if [[ -v TELEMETRY_ENABLED ]]; then
#   log::info "Copying components file for telemetry tests"
#   #shellcheck disable=SC2088
#   utils::send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d-telemetry-components.yaml" "~/kyma-integration-k3d-telemetry-components.yaml"
# fi

# if [[ -v ISTIO_INTEGRATION_ENABLED ]]; then
#   log::info "Copying components file for telemetry tests"
#   #shellcheck disable=SC2088
#   utils::send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d-istio-components.yaml" "~/kyma-integration-k3d-istio-components.yaml"
# fi

# log::info "Triggering the installation"
# utils::ssh_to_vm_with_script -z "${ZONE}" -n "kyma-integration-test-${RANDOM_ID}" -c "sudo bash" -p "${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d.sh"

log::info "Provision cluster"
utils::ssh_to_vm_with_script -z "${ZONE}" -n "kyma-integration-test-${RANDOM_ID}" -c "sudo bash" -p "${SCRIPT_DIR}/cluster-integration/helpers/set-up-vm-k3d-cluster.sh"
mkdir -p "$HOME/.kube"
utils::receive_from_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "~/kubeconfig.yaml" "$HOME/.kube/config"
export KUBECONFIG="$HOME/.kube/config"


function deploy_kyma() {
  log::info "Printing client and server version info"
  kubectl version

  local kyma_deploy_cmd
  kyma_deploy_cmd="kyma deploy -p evaluation --ci --source=local --workspace ${KYMA_SOURCES_DIR}"

  if [[ -v ISTIO_INTEGRATION_ENABLED ]]; then
    log::info "Installing Kyma with ${KYMA_PROFILE} profile"
    kyma_deploy_cmd="kyma deploy -p ${KYMA_PROFILE} --ci --source=local --workspace ${KYMA_SOURCES_DIR} --components-file ${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d-istio-components.yaml"
  fi

  if [[ -v CENTRAL_APPLICATION_CONNECTIVITY_ENABLED ]]; then
    kyma_deploy_cmd+=" --value application-connector.central_application_gateway.enabled=true"
  fi

  if [[ -v COMPASS_INTEGRATION_ENABLED ]]; then
    kyma_deploy_cmd+=" --value global.disableLegacyConnectivity=true"
    kyma_deploy_cmd+=" --value compass-runtime-agent.compassRuntimeAgent.config.skipAppsTLSVerification=true"
    kyma_deploy_cmd+=" --components-file ${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d-compass-components.yaml"
  fi

  if [[ -v TELEMETRY_ENABLED ]]; then
    kyma_deploy_cmd+=" --value=global.telemetry.enabled=true"
    kyma_deploy_cmd+=" --components-file ${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d-telemetry-components.yaml"
  fi

  $kyma_deploy_cmd

  kubectl get pods -A
}


function run_tests() {
  pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
  if [[ -v COMPASS_INTEGRATION_ENABLED && -v CENTRAL_APPLICATION_CONNECTIVITY_ENABLED ]]; then
    make ci-application-connectivity-2-compass
  elif [[ -v COMPASS_INTEGRATION_ENABLED ]]; then
    make ci-compass
  elif [[ -v TELEMETRY_ENABLED ]]; then
    npm install
    npm run test-telemetry
  elif [[ -v ISTIO_INTEGRATION_ENABLED ]]; then
    pushd "../components/istio"
    go install github.com/cucumber/godog/cmd/godog@latest
    make test
    popd
  else
    make ci
  fi
  popd
}

kyma::install_cli
deploy_kyma
run_tests

log::success "all done"
