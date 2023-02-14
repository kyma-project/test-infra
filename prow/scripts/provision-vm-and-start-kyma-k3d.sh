#!/bin/bash

# This script is designed to provision a new vm and start kyma.It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisining vms. If no flag is provided, the latest custom image is used.

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
  log::info "Execute Job Guard"
  "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

function cleanup() {
  # TODO - collect junit results

  log::info "Stopping instance kyma-integration-test-${RANDOM_ID}"
  log::info "It will be removed automatically by cleaner job"

  # do not fail the job regardless of the vm deletion result
  set +e

  #shellcheck disable=SC2088
  if [[ "$ISTIO_INTEGRATION_ENABLED" == "true" ]]; then
    utils::receive_from_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "~/kyma/tests/components/istio/junit-report.xml" "${ARTIFACTS}"
    utils::receive_from_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "~/kyma/tests/components/istio/reports/*.html" "${ARTIFACTS}"
  elif [[ -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_GATEWAY ]] || [[ -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_VALIDATOR ]] || [[ -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_RUNTIME_AGENT ]]; then
    utils::receive_from_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "~/kyma/tests/components/application-connector/junit-report.xml" "${ARTIFACTS}"
  elif [[ "$API_GATEWAY_INTEGRATION" == "true" ]]; then
    utils::receive_from_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "~/kyma/tests/components/api-gateway/junit-report.xml" "${ARTIFACTS}"
    utils::receive_from_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "~/kyma/tests/components/api-gateway/reports/*.html" "${ARTIFACTS}/report.html"
  else
    utils::receive_from_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "~/kyma/tests/fast-integration/junit_kyma-fast-integration.xml" "${ARTIFACTS}"
  fi

  gcloud compute instances stop --async --zone="${ZONE}" "kyma-integration-test-${RANDOM_ID}"

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

trap cleanup exit INT

log::info "Preparing environment variables for the instance"

export GARDENER_ZONE=${ZONE}
export CLUSTER_NAME=kyma-integration-test-${RANDOM_ID}

envVars=(
  COMPASS_TENANT
  COMPASS_HOST
  COMPASS_CLIENT_ID
  COMPASS_CLIENT_SECRET
  COMPASS_INTEGRATION_ENABLED
  CENTRAL_APPLICATION_CONNECTIVITY_ENABLED
  APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_GATEWAY
  APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_VALIDATOR
  APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_RUNTIME_AGENT
  TELEMETRY_ENABLED
  TELEMETRY_TRACING_ENABLED
  ISTIO_INTEGRATION_ENABLED
  API_GATEWAY_INTEGRATION
  GARDENER_ZONE
  CLUSTER_NAME
  KYMA_PROFILE
  RECONCILATION_TEST
  K8S_VERSION
)
utils::save_env_file "${envVars[@]}"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" ".env" "~/.env"

log::info "Copying Kyma to the instance"
#shellcheck disable=SC2088
utils::compress_send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "/home/prow/go/src/github.com/kyma-project/kyma" "~/kyma"

if [[ -v API_GATEWAY_INTEGRATION ]]; then
  log::info "Copying components file for API-gateway tests"
  #shellcheck disable=SC2088
  utils::send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d-api-gateway-components.yaml" "~/kyma-integration-k3d-api-gateway-components.yaml"
fi

if [[ -v COMPASS_INTEGRATION_ENABLED ]]; then
  log::info "Copying components file for compass tests"
  #shellcheck disable=SC2088
  utils::send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d-compass-components.yaml" "~/kyma-integration-k3d-compass-components.yaml"
fi

if [[ -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_GATEWAY ]]; then
  log::info "Copying components file for application connector component tests for Kyma OS setup"
  #shellcheck disable=SC2088
  utils::send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d-app-connector-components-os.yaml" "~/kyma-integration-k3d-app-connector-components-os.yaml"
fi

if [[ -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_VALIDATOR ]] || [[ -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_RUNTIME_AGENT ]]; then
  log::info "Copying components file for application connector component tests for Kyma SKR setup"
  #shellcheck disable=SC2088
  utils::send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d-app-connector-components-skr.yaml" "~/kyma-integration-k3d-app-connector-components-skr.yaml"
fi

if [[ -v TELEMETRY_ENABLED ]]; then
  log::info "Copying components file for telemetry tests"
  #shellcheck disable=SC2088
  utils::send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d-telemetry-components.yaml" "~/kyma-integration-k3d-telemetry-components.yaml"
  log::info "Triggering the installation"
  utils::ssh_to_vm_with_script -z "${ZONE}" -n "kyma-integration-test-${RANDOM_ID}" -c "sudo bash" -p "${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d-telemetry.sh"
  log::success "all done"
  exit 0
fi

if [[ -v ISTIO_INTEGRATION_ENABLED ]]; then
  log::info "Copying components file for istio tests"
  #shellcheck disable=SC2088
  utils::send_to_vm "${ZONE}" "kyma-integration-test-${RANDOM_ID}" "${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d-istio-components.yaml" "~/kyma-integration-k3d-istio-components.yaml"
fi

log::info "Triggering the installation"
utils::ssh_to_vm_with_script -z "${ZONE}" -n "kyma-integration-test-${RANDOM_ID}" -c "sudo bash" -p "${SCRIPT_DIR}/cluster-integration/kyma-integration-k3d.sh"

log::success "all done"
