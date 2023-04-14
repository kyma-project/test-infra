#!/usr/bin/env bash
set -o errexit

readonly TEST_INFRA_DIR="./test-infra"
readonly GO_VERSION=1.19
export KYMA_SOURCES_DIR="./kyma"
export KUBECONFIG="${HOME}/.kube/config"
export CLUSTER_DOMAIN="local.kyma.dev"

function prereq_test() {
  command -v node >/dev/null 2>&1 || { echo >&2 "node not found"; exit 1; }
  command -v npm >/dev/null 2>&1 || { echo >&2 "npm not found"; exit 1; }
  command -v jq >/dev/null 2>&1 || { echo >&2 "jq not found"; exit 1; }
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

  wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && export PATH=$PATH:/usr/local/go/bin && go version

  wget -q "https://github.com/istio/istio/releases/download/${ISTIO_VERSION}/istioctl-${ISTIO_VERSION}-linux-amd64.tar.gz"   && tar -C /usr/local/bin -xzf "istioctl-${ISTIO_VERSION}-linux-amd64.tar.gz" && export PATH=$PATH:/usr/local/bin/istioctl && istioctl version --remote=false && export ISTIOCTL_PATH=/usr/local/bin/istioctl
}

function provision_k3d() {
  log::info "Provisioning k3d cluster"

  k3d version
  kyma provision k3d --ci
  log::success "K3d cluster provisioned"
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

function istio::prepare_components_file() {
  log::info "Preparing Kyma installation with Istio prerequisites"

cat << EOF > "$PWD/istio.yaml"
defaultNamespace: kyma-system
prerequisites:
  - name: "cluster-essentials"
  - name: "istio"
    namespace: "istio-system"
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

  local k3d_cni_overrides="istio.helmValues.cni.cniConfDir=/var/lib/rancher/k3s/agent/etc/cni/net.d,istio.helmValues.cni.cniBinDir=/bin"
  kyma_deploy_cmd="./bin/mothership-linux local --kubeconfig ${KUBECONFIG} --value global.ingress.domainName=${CLUSTER_DOMAIN},global.domainName=${CLUSTER_DOMAIN},${k3d_cni_overrides} --version ${KYMA_VERSION} --profile ${EXECUTION_PROFILE}"

  if [[ $TEST_NAME == ory ]]; then
    ory::prepare_components_file
    kyma_deploy_cmd+=" --components-file ${PWD}/ory.yaml"
  fi

  if [[ $TEST_NAME == istio ]]; then
    istio::prepare_components_file
    kyma_deploy_cmd+=" --components-file ${PWD}/istio.yaml"
  fi

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

  if [[ $TEST_NAME == ory ]]; then
    export ORY_RECONCILER_INTEGRATION_TESTS=1
    go test -v -timeout 5m ./pkg/reconciler/instances/"${TEST_NAME}"/test
  fi

  if [[ $TEST_NAME == istio ]]; then
    export ISTIO_RECONCILER_INTEGRATION_TESTS=1
    export INGRESS_PORT=80
    go test -v -timeout 5m ./pkg/reconciler/instances/"${TEST_NAME}"/tests
  fi
  #currently disabling
  #make: go: Permission denied on Gardener
  #make test-ory
  popd
}

load_env
if [[ "${EXECUTION_PROFILE}" == "evaluation" ]]; then

readonly RECONCILER_DIR="./reconciler"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_DIR}/prow/scripts/lib/log.sh"

prereq_test \
  && install_prereq \
  && provision_k3d 

else

readonly RECONCILER_DIR="${RECONCILER_SOURCES_DIR}"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"

prereq_test \
 && install_prereq
fi

deploy_kyma \
  && run_tests
