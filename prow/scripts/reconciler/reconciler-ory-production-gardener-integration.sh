#!/usr/bin/env bash
set -o errexit

readonly GO_VERSION=1.19
export KYMA_SOURCES_DIR="./kyma"
export KUBECONFIG="${HOME}/.kube/config"
export CLUSTER_DOMAIN="local.kyma.dev"

function prereq_test() {
  command -v helm >/dev/null 2>&1 || { echo >&2 "helm not found"; exit 1; }
  command -v kubectl >/dev/null 2>&1 || { echo >&2 "kubectl not found"; exit 1; }
}

function load_env() {
  ENV_FILE=".env"
  if [ -f "${ENV_FILE}" ]; then
    # shellcheck disable=SC2046
    export $(xargs < "${ENV_FILE}")
  fi
}

function install_prereq() {
  log::info "Installing Go"

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

  wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && export PATH=$PATH:/usr/local/go/bin && go version
}

function ory::prepare_components_file() {
  log::info "Preparing Kyma installation with Ory and prerequisites"

cat << EOF > "$PWD/ory.yaml"
defaultNamespace: kyma-system
prerequisites:
  - name: "cluster-essentials"
  - name: "istio"
    namespace: "istio-system"
components:
  - name: "ory"
  - name: "istio-resources"
EOF
}

function deploy_kyma() {
  log::info "Building reconciler from sources"
  pushd "${RECONCILER_DIR}"
  make build-linux
  if [ ! -f "./bin/mothership-linux" ]; then
     # shellcheck disable=SC2046
     log::error "Mothership-linux binary was not built."
     exit 1
  fi

  local kyma_deploy_cmd

  kyma_deploy_cmd="./bin/mothership-linux local --kubeconfig ${KUBECONFIG} --value global.ingress.domainName=${CLUSTER_DOMAIN},global.domainName=${CLUSTER_DOMAIN} --version ${KYMA_VERSION} --profile production"

  ory::prepare_components_file
  kyma_deploy_cmd+=" --components-file ${PWD}/ory.yaml"

  log::info "Deploying Kyma components from version ${KYMA_VERSION}"

  $kyma_deploy_cmd
  local kyma_deploy_exit_code=$?

  kubectl get pods -A

  if [ $kyma_deploy_exit_code -ne 0 ]; then
      log::error "Error during deployment"
      exit 1
  else
    log::success "Kyma components were deployed successfully"
  fi

  popd
}

function run_tests() {
  log::info "Running tests"

  pushd "${RECONCILER_DIR}"

  export ORY_RECONCILER_INTEGRATION_TESTS=1
  go test -v -timeout 5m ./pkg/reconciler/instances/ory/test

  #currently disabling
  #make: go: Permission denied on Gardener
  #make test-ory
  popd
}

load_env

readonly RECONCILER_DIR="${RECONCILER_SOURCES_DIR}"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"

prereq_test \
 && install_prereq

deploy_kyma \
  && run_tests
