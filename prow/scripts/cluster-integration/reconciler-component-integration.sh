#!/usr/bin/env bash

set -o errexit
set -o pipefail

readonly RECONCILER_DIR="./reconciler"
readonly TEST_INFRA_DIR="./test-infra"
readonly GO_VERSION=1.17.5
export KYMA_SOURCES_DIR="./kyma"
export KYMA_VERSION="main"
export KUBECONFIG="${HOME}/.kube/config"
export ISTIOCTL_VERSION="1.11.4"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_DIR}/prow/scripts/lib/log.sh"

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

function install_prereq() {
  log::info "Installing Kyma CLI, Go and Istioctl"

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

  wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && export PATH=$PATH:/usr/local/go/bin && go version

  wget -q https://github.com/istio/istio/releases/download/${ISTIOCTL_VERSION}/istioctl-${ISTIOCTL_VERSION}-linux-amd64.tar.gz   && sudo tar -C /usr/local/bin -xzf istioctl-${ISTIOCTL_VERSION}-linux-amd64.tar.gz && export PATH=$PATH:/usr/local/bin/istioctl && istioctl version --remote=false && export ISTIOCTL_PATH=/usr/local/bin/istioctl
}

function provision_k3d() {
  log::info "Provisioning k3d cluster"

  k3d version
  kyma provision k3d --ci
  log::success "K3d cluster provisioned"
}

function ory::prepare_components_file() {
  log::info "Preparing Kyma installation with Ory and prerequisites"

cat << EOF > "$PWD/components.yaml"
defaultNamespace: kyma-system
prerequisites:
  - name: "cluster-essentials"
  - name: "istio-configuration"
    namespace: "istio-system"
components:
  - name: "ory"
EOF
}

function deploy_kyma() {
  log::info "Building reconciler from sources"

  pushd "${RECONCILER_DIR}"
  make build-linux

  local kyma_deploy_cmd
  kyma_deploy_cmd="./bin/mothership-linux local --kubeconfig ${KUBECONFIG} --value global.ingress.domainName=kyma.local,global.domainName=kyma.local --version ${KYMA_VERSION} --profile ${EXECUTION_PROFILE}"

  if [[ $TEST_NAME == ory ]]; then
    ory::prepare_components_file
    kyma_deploy_cmd+=" --components-file $PWD/components.yaml"
  fi
  log::info "Deploying Kyma components from version ${KYMA_VERSION}"

  $kyma_deploy_cmd

  popd

  log::success "Kyma components were deployed successfully"
  kubectl get pods -A
}

function run_tests() {
  log::info "Running tests"

  pushd "${RECONCILER_DIR}"
  make test-ory
  popd
}

prereq_test
load_env
install_prereq
provision_k3d
deploy_kyma
run_tests
