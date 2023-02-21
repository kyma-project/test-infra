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
  command -v go >/dev/null 2>&1 || { echo >&2 "go not found"; exit 1; }
}

function load_env() {
  ENV_FILE=".env"
  if [ -f "${ENV_FILE}" ]; then
    # shellcheck disable=SC2046
    export $(xargs < "${ENV_FILE}")
  fi
  export PATH="${PATH}:/usr/local/go/bin"
  export PATH="${PATH}:~/go/bin"
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

function deploy_kyma() {
  k3d version

  if [[ -v K8S_VERSION ]]; then
    echo "Creating k3d with kuberenetes version: ${K8S_VERSION}"
    kyma provision k3d --ci -k "${K8S_VERSION}"
  else
    kyma provision k3d --ci
  fi

  echo "Printing client and server version info"

  kubectl version

  local kyma_deploy_cmd
  kyma_deploy_cmd="kyma deploy -p evaluation --ci --source=local --workspace ${KYMA_SOURCES_DIR}"

  if [[ -v API_GATEWAY_INTEGRATION ]]; then
    echo "Executing API-gateway tests on k3d"
    kyma_deploy_cmd+=" --components-file kyma-integration-k3d-api-gateway-components.yaml"
  fi

  if [[ -v ISTIO_INTEGRATION_ENABLED ]]; then
    echo "Installing Kyma with ${KYMA_PROFILE} profile"
    kyma_deploy_cmd="kyma deploy -p ${KYMA_PROFILE} --ci --source=local --workspace ${KYMA_SOURCES_DIR} --components-file kyma-integration-k3d-istio-components.yaml"
  fi

  if [[ -v CENTRAL_APPLICATION_CONNECTIVITY_ENABLED ]]; then
    kyma_deploy_cmd+=" --value application-connector.central_application_gateway.enabled=true"
  fi

  if [[ -v COMPASS_INTEGRATION_ENABLED ]]; then
    kyma_deploy_cmd+=" --value global.disableLegacyConnectivity=true"
    kyma_deploy_cmd+=" --value compass-runtime-agent.compassRuntimeAgent.config.skipAppsTLSVerification=true"
    kyma_deploy_cmd+=" --components-file kyma-integration-k3d-compass-components.yaml"
  fi

  if [[ -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_VALIDATOR || -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_RUNTIME_AGENT ]]; then
    kyma_deploy_cmd+=" --value global.disableLegacyConnectivity=true"
    kyma_deploy_cmd+=" --value compass-runtime-agent.compassRuntimeAgent.config.skipAppsTLSVerification=true"
    kyma_deploy_cmd+=" --components-file kyma-integration-k3d-app-connector-components-skr.yaml"
  fi

  if [[ -v  APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_GATEWAY ]]; then
    kyma_deploy_cmd+=" --components-file kyma-integration-k3d-app-connector-components-os.yaml"
  fi

  if [[ -v TELEMETRY_TRACING_ENABLED ]]; then
    kyma_deploy_cmd+=" --value=telemetry.operator.controllers.tracing.enabled=true"
  fi

  if [[ -v TELEMETRY_ENABLED ]]; then
    kyma_deploy_cmd+=" --value=global.telemetry.enabled=true"
    kyma_deploy_cmd+=" --components-file kyma-integration-k3d-telemetry-components.yaml"
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

  elif [[ -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_GATEWAY || -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_VALIDATOR || -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_RUNTIME_AGENT ]]; then
      pushd "../components/application-connector"
      export EXPORT_RESULT="true"
      go install github.com/jstemmer/go-junit-report/v2@latest

      if [[ -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_GATEWAY ]]; then
        make test -f Makefile.test-application-gateway
      elif [ -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_VALIDATOR ]; then
        make test -f Makefile.test-application-conn-validator
      elif [ -v APPLICATION_CONNECTOR_COMPONENT_TESTS_ENABLED_RUNTIME_AGENT ]; then
        make test -f Makefile.test-compass-runtime-agent
      fi

      popd
  elif [[ -v ISTIO_INTEGRATION_ENABLED ]]; then
    pushd "../components/istio"
    export EXPORT_RESULT="true"
    go install github.com/cucumber/godog/cmd/godog@latest
    make test
    popd
  elif [[ -v API_GATEWAY_INTEGRATION ]]; then
    pushd "../components/api-gateway"
    export EXPORT_RESULT="true"
    export TEST_CONCURENCY="8"
    export KYMA_DOMAIN="local.kyma.dev"
    export TEST_DOMAIN="local.kyma.dev"
    export TEST_HYDRA_ADDRESS="https://oauth2.local.kyma.dev"
    go install github.com/cucumber/godog/cmd/godog@latest
    make test-k3d
    popd
  else
    make ci
  fi
  popd
}

load_env
prereq_test
install_cli
deploy_kyma
run_tests
