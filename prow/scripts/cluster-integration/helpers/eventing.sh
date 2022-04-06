#!/usr/bin/env bash

set -e

readonly BACKEND_SECRET_NAME=eventing-backend
readonly BACKEND_SECRET_NAMESPACE=default
readonly BACKEND_SECRET_LABEL_KEY=kyma-project.io/eventing-backend
readonly BACKEND_SECRET_LABEL_VALUE=NATS
readonly EVENTING_BACKEND_CR_NAME=eventing-backend
readonly EVENTING_BACKEND_CR_NAMESPACE=kyma-system

# Check if required vars are set or not
function eventing::check_required_vars() {
  if [[ -z ${CREDENTIALS_DIR} ]]; then
    echo "required variable CREDENTIALS_DIR is missing"
    exit 1
  fi
}

# Create a Kubernetes Secret which contains the EventMesh service key
function eventing::create_eventmesh_secret() {
  eventing::check_required_vars

  pushd "${CREDENTIALS_DIR}"

  SECRET_NAME=event-mesh
  SECRET_NAMESPACE=default

  SERVICE_KEY_VALUE=$(base64 -i serviceKey | tr -d '\n')

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: ${SECRET_NAME}
  namespace: ${SECRET_NAMESPACE}
  labels:
    kyma-project.io/event-mesh: "true"
data:
  serviceKey: "${SERVICE_KEY_VALUE}"
EOF

  popd
}

# Create a Kubernetes Secret which is needed by the Eventing Backend controller
function eventing::create_eventing_backend_secret() {
  eventing::check_required_vars

  pushd "${CREDENTIALS_DIR}"

  SECRET_NAME="${BACKEND_SECRET_NAME}"
  SECRET_NAMESPACE="${BACKEND_SECRET_NAMESPACE}"

  MANAGEMENT=$(jq -r  '.management' < serviceKey | tr -d '[:space:]' | base64 | tr -d '\n')
  MESSAGING=$(jq -r '.messaging' < serviceKey | tr -d '[:space:]' | base64 | tr -d '\n')
  NAMESPACE=$(jq -r '.namespace' < serviceKey | tr -d '[:space:]' | base64 | tr -d '\n')
  SERVICE_INSTANCE_ID=$(jq -r '.serviceinstanceid' < serviceKey | tr -d '[:space:]' | base64 | tr -d '\n')
  XS_APP_NAME=$(jq -r '.xsappname' < serviceKey | tr -d '[:space:]' | base64 | tr -d '\n')

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: ${SECRET_NAME}
  namespace: ${SECRET_NAMESPACE}
data:
  management: "${MANAGEMENT}"
  messaging: "${MESSAGING}"
  namespace: "${NAMESPACE}"
  serviceinstanceid: "${SERVICE_INSTANCE_ID}"
  xsappname: "${XS_APP_NAME}"
EOF

  popd
}

# Create a Kubernetes Secret which is needed by the Eventing Publisher and Subscription Controller
function eventing::create_eventing_secret() {
  eventing::check_required_vars

  pushd "${CREDENTIALS_DIR}"

  SECRET_NAME=eventing
  SECRET_NAMESPACE=kyma-system

  # delete the default Eventing secret
  kubectl delete secret -n ${SECRET_NAMESPACE} ${SECRET_NAME}

  HTTP_REST=$(jq -r '.messaging' < serviceKey | jq -c '.[] | select(.broker.type | contains("saprestmgw"))')
  BEB_NAMESPACE=$(jq -r '.namespace' < serviceKey | tr -d '[:space:]' | base64 | tr -d '\n')
  CLIENT_ID=$(echo "$HTTP_REST" | jq -r '.oa2.clientid' | tr -d '[:space:]' | base64 | tr -d '\n')
  CLIENT_SECRET=$(echo "$HTTP_REST" | jq -r '.oa2.clientsecret' | tr -d '[:space:]' | base64 | tr -d '\n')
  EMS_PUBLISH_URL=$(echo "$HTTP_REST" | jq -r '.uri' | tr -d '[:space:]' | base64 | tr -d '\n')
  TOKEN_ENDPOINT=$(echo "$HTTP_REST" | jq -r '.oa2.tokenendpoint' | tr -d '[:space:]' | base64 | tr -d '\n')

  # create Eventing secret with the proper values
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: ${SECRET_NAME}
  namespace: ${SECRET_NAMESPACE}
data:
  beb-namespace: "${BEB_NAMESPACE}"
  client-id: "${CLIENT_ID}"
  client-secret: "${CLIENT_SECRET}"
  ems-publish-url: "${EMS_PUBLISH_URL}"
  token-endpoint: "${TOKEN_ENDPOINT}"
EOF

  popd
}

# Switches the eventing backend based on the passed parameter (NATS or BEB).
# If there is no parameter passed, NATS is used as the default backend.
function eventing::switch_backend() {
  labelValue="$(echo "${1}" | tr -s '[:upper:]' '[:lower:]')"
  if [[ -z "${labelValue}" ]]; then
      labelValue="$(echo "${BACKEND_SECRET_LABEL_VALUE}" | tr -s '[:upper:]' '[:lower:]')"
  fi

  echo "label backend secret with ${BACKEND_SECRET_LABEL_KEY}=${labelValue}"
  kubectl label secret --namespace "${BACKEND_SECRET_NAMESPACE}" "${BACKEND_SECRET_NAME}" "${BACKEND_SECRET_LABEL_KEY}=${labelValue}" --overwrite
}

# Waits for Eventing backend to be ready by checking the EventingBackend custom resource status
function eventing::wait_for_backend_ready() {
  if [[ -z "${1}" ]]; then
    echo "backend type is missing"
    exit 1
  fi

  # wait for Eventing backend custom resource old status to be cleared
  sleep 10s

  retry=0
  maxRetires=20
  wantReady="$(echo "true" | tr -s '[:upper:]' '[:lower:]')"
  wantBackend="$(echo "${1}" | tr -s '[:upper:]' '[:lower:]')"

  while [[ ${retry} -lt ${maxRetires} ]]; do
      ready=$(kubectl get eventingbackends.eventing.kyma-project.io --namespace "${EVENTING_BACKEND_CR_NAMESPACE}" "${EVENTING_BACKEND_CR_NAME}" -ojsonpath="{.status.eventingReady}" | tr -s '[:upper:]' '[:lower:]')
      backend=$(kubectl get eventingbackends.eventing.kyma-project.io --namespace "${EVENTING_BACKEND_CR_NAMESPACE}" "${EVENTING_BACKEND_CR_NAME}" -ojsonpath="{.status.backendType}" | tr -s '[:upper:]' '[:lower:]')

      if [[ "${ready}" == "${wantReady}" && "${backend}" == "${wantBackend}" ]]; then
          echo "Eventing backend [${1}] is ready"
          kubectl get eventingbackends.eventing.kyma-project.io --namespace "${EVENTING_BACKEND_CR_NAMESPACE}" "${EVENTING_BACKEND_CR_NAME}"
          return 0
      fi

      echo "try $((retry + 1))/${maxRetires} waiting for Eventing backend ${1} to be ready - current backend status ${backend}/${ready}"
      retry=$((retry + 1))
      sleep 10
  done

  echo "Eventing backend [${1}] is not ready"
  kubectl get eventingbackends.eventing.kyma-project.io --namespace "${EVENTING_BACKEND_CR_NAMESPACE}" "${EVENTING_BACKEND_CR_NAME}"
  return 1
}

# Runs eventing specific fast-integration tests
function eventing::test_fast_integration_eventing() {
    log::info "Running Eventing E2E release tests"

    pushd /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration
    make ci-test-eventing
    popd

    log::success "Eventing tests completed"
}

# Runs eventing script to provision SKR
function eventing::test_fast_integration_provision_skr() {
    log::info "Running Eventing script to provision SKR"

    pushd /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration
    make ci-test-eventing-provision-skr
    popd

    log::success "Provision SKR completed"
}

# Runs eventing script to de-provision SKR
function eventing::test_fast_integration_deprovision_skr() {
    log::info "Running Eventing script to de-provision SKR"

    pushd /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration
    make ci-test-eventing-deprovision-skr
    popd

    log::success "De-provision SKR completed"
}

# Sets KUBECONFIG to ~/.kube/config
function eventing::set_default_kubeconfig_env() {
    log::info "Setting default KUBECONFIG ~/.kube/config"

    export KUBECONFIG="${HOME}/.kube/config"
}

function eventing::pre_upgrade_test_fast_integration() {
    log::info "Running pre upgrade Eventing E2E release tests"

    pushd /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration
    make ci-test-eventing-pre-upgrade
    popd

    log::success "Pre upgrade Eventing tests completed"
}
function eventing::fast_integration_tests() {
    log::info "Running only Eventing E2E release tests"

    pushd /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration
    make ci-test-eventing-tests
    popd

    log::success "Eventing tests completed"
}

function eventing::post_upgrade_test_fast_integration() {
    log::info "Running post upgrade Eventing E2E release tests and clean up the resources"

    pushd /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration
    make ci-test-eventing-post-upgrade
    popd

    log::success "Post upgrade Eventing tests completed"
}

function eventing::fast_integration_test_cleanup() {
    log::info "Running fast integration tests cleanup to remove the testing resources such as namespaces and compass scenario"

    pushd /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration
    npm run eventing-test-cleanup
    popd

    log::success "Fast integration tests cleanup completed"
}