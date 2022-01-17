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
  kyma provision k3d --ci

  kyma deploy -p evaluation --ci --source=local --workspace ${KYMA_SOURCES_DIR}
}

function install_operator() {
  helm install -n kyma-system telemetry ${KYMA_SOURCES_DIR}/resources/telemetry

  kubectl get daemonset -n kyma-system
}

function install_mockserver() {
  local mock_namespace="mockserver"
  kubectl create namespace ${mock_namespace}

  helm install -n ${mock_namespace} mockserver ${KYMA_SOURCES_DIR}/tests/fast-integration/telemetry-test/helm/mockserver
  
  kubectl rollout -n ${mock_namespace} status deployments mockserver

  helm install -n ${mock_namespace} mockserver-config ${KYMA_SOURCES_DIR}/tests/fast-integration/telemetry-test/helm/mockserver-config

  kubectl -n mockserver port-forward svc/mockserver 1080:1080 &
}

function run_test() {
  kubectl apply -f ${KYMA_SOURCES_DIR}/tests/fast-integration/telemetry-test/log-pipeline.yaml
  sleep 10
  kubectl describe ds -n kyma-system telemetry-fluent-bit 
  kubectl describe po -n kyma-system --selector=control-plane=telemetry-operator-controller-manager

  POD_NAME=$(kubectl get po -n kyma-system --no-headers=true --selector=app.kubernetes.io/instance=telemetry,app.kubernetes.io/name=fluent-bit -o custom-columns=:metadata.name  | head -n 1)
  kubectl logs ${POD_NAME} -n kyma-system
  # pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
  # npm install
  # DEBUG=true npm run test-telemetry
  # popd
}

prereq_test
install_cli
deploy_kyma
install_operator
install_mockserver
run_test
