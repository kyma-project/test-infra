#!/usr/bin/env bash

set -o errexit
set -o pipefail

export KYMA_SOURCES_DIR="./kyma"

function prereq_test() {
  command -v node >/dev/null 2>&1 || { echo >&2 "node not found"; exit 1; }
  command -v npm >/dev/null 2>&1 || { echo >&2 "npm not found"; exit 1; }
  command -v jq >/dev/null 2>&1 || { echo >&2 "jq not found"; exit 1; }
  command -v helm >/dev/null 2>&1 || { echo >&2 "helm not found"; exit 1; }
  command -v kubectl >/dev/null 2>&1 || { echo >&2 "kubectl not found"; exit 1; }
  command -v k3d >/dev/null 2>&1 || { echo >&2 "k3d not found"; exit 1; }
}

function load_env() {
  ENV_FILE=".env"
  if [ -f "${ENV_FILE}" ]; then
    # shellcheck disable=SC2046
    export $(xargs < "${ENV_FILE}")
  fi
}

function install_cli() {
  local install_dir
  declare -r install_dir="/usr/local/bin"
  mkdir -p "$install_dir"

  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  if [[ -z "$os" || ! "$os" =~ ^(darwin|linux)$ ]]; then
    echo >&2 -e "Unsupported host OS. Must be Linux or Mac OS X."
    exit 1
  else
    readonly os
  fi

  pushd "$install_dir" || exit
  curl -Lo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}"
  chmod +x kyma
  popd

  kyma version --client
}

# Install Helm Charts required for the telemetry integration test 
function install_telemetry_helm_charts() {  
  helm install -n kyma-system telemetry ${KYMA_SOURCES_DIR}/resources/telemetry

  local mock_namespace="mockserver"
  kubectl create namespace ${mock_namespace}
  helm install -n ${mock_namespace} mockserver ${KYMA_SOURCES_DIR}/tests/fast-integration/telemetry-test/helm/mockserver
  helm install -n ${mock_namespace} mockserver-config ${KYMA_SOURCES_DIR}/tests/fast-integration/telemetry-test/helm/mockserver-config
}

function deploy_kyma() {
  k3d version

  if [[ -v K8S_VERSION ]]; then
    kyma provision k3d --ci -k "${K8S_VERSION}"
  else
    kyma provision k3d --ci
  fi

  local kyma_deploy_cmd
  kyma_deploy_cmd="kyma deploy -p evaluation --ci --source=local --workspace ${KYMA_SOURCES_DIR}"

  if [[ -v CENTRAL_APPLICATION_CONNECTIVITY_ENABLED ]]; then
    kyma_deploy_cmd+=" --value application-connector.central_application_gateway.enabled=true"
    kyma_deploy_cmd+=" --value global.centralApplicationConnectivityValidatorEnabled=true"
  fi

  if [[ -v COMPASS_INTEGRATION_ENABLED ]]; then
    kyma_deploy_cmd+=" --value global.disableLegacyConnectivity=true"
    kyma_deploy_cmd+=" --value compass-runtime-agent.compassRuntimeAgent.config.skipAppsTLSVerification=true"
    kyma_deploy_cmd+=" --components-file kyma-integration-k3d-compass-components.yaml"
  fi

  $kyma_deploy_cmd

  if [[ -v TELEMETRY_ENABLED ]]; then
    install_telemetry_helm_charts
  fi

  kubectl get pods -A
}


function run_tests() {
  pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
  if [[ -v COMPASS_INTEGRATION_ENABLED && -v CENTRAL_APPLICATION_CONNECTIVITY_ENABLED ]]; then
    make ci-application-connectivity-2-compass
  elif [[ -v COMPASS_INTEGRATION_ENABLED ]]; then
    make ci-compass
  elif [[ -v CENTRAL_APPLICATION_CONNECTIVITY_ENABLED ]]; then
    make ci-application-connectivity-2
  elif [[ -v TELEMETRY_ENABLED ]]; then
    npm install
    npm run test-telemetry
  else
    make ci
  fi
  popd
}

prereq_test
load_env
install_cli
deploy_kyma
run_tests
