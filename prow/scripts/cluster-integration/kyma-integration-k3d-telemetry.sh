#!/usr/bin/env bash

set -o errexit
set -o pipefail

export KYMA_SOURCES_DIR="./kyma"

function prereq_test() {
  command -v node >/dev/null 2>&1 || { echo >&2 "node not found"; exit 1; }
  command -v npm >/dev/null 2>&1 || { echo >&2 "npm not found"; exit 1; }
  command -v jq >/dev/null 2>&1 || { echo >&2 "jq not found"; exit 1; }
  command -v kubectl >/dev/null 2>&1 || { echo >&2 "kubectl not found"; exit 1; }
  command -v k3d >/dev/null 2>&1 || { echo >&2 "k3d not found"; exit 1; }
  command -v go >/dev/null 2>&1 || { echo >&2 "go not found"; exit 1; }
}

cleanup() {
  rm "teltest.yaml" || true
}

trap cleanup EXIT SIGINT

function load_env() {
  ENV_FILE=".env"
  if [ -f "${ENV_FILE}" ]; then
    read -ra args < <(xargs < "${ENV_FILE}")
    export "${args[@]}"
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

  echo "Kubernetes client and server version:"

  kubectl version --output=yaml

  local deploy
  local deploy_commands
  local deploy_dryrun
  
  deploy_commands=" --ci
   --source=local
   --workspace ${KYMA_SOURCES_DIR}
   --value=telemetry.operator.controllers.tracing.enabled=true
   --components-file kyma-integration-k3d-telemetry-components.yaml"
  deploy="kyma deploy -p evaluation"
  deploy_dryrun="kyma deploy --dry-run  -p production"

  ls ${KYMA_SOURCES_DIR}/components/telemetry-operator/config/crd/
  echo "Copy tracepipeline CRD"
  cp ${KYMA_SOURCES_DIR}/components/telemetry-operator/config/crd/bases/telemetry.kyma-project.io_tracepipelines.yaml ${KYMA_SOURCES_DIR}/installation/resources/crds/telemetry/tracepipelines.crd.yaml

  deploy+="$deploy_commands"
  deploy_dryrun+="$deploy_commands"

  $deploy
  $deploy_dryrun 1>/dev/null

  kubectl get pods -A
}


function run_tests() {
  pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
  make telemetry
  kubectl logs --tail=-1 -l control-plane=telemetry-operator -n kyma-system -c manager
  popd
}

load_env
prereq_test
install_cli
deploy_kyma
run_tests
