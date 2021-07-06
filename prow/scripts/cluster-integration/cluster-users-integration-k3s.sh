#!/usr/bin/env bash

set -o errexit
set -o pipefail # Fail a pipe if any sub-command fails.

date

# https://github.com/kyma-project/test-infra/pull/2967 - explanation for that kaniko image
export KANIKO_IMAGE="eu.gcr.io/kyma-project/external/aerfio/kaniko:v1.5.1"
export DOMAIN=${KYMA_DOMAIN:-local.kyma.dev}
if [[ -z $REGISTRY_VALUES ]]; then
  export REGISTRY_VALUES="dockerRegistry.enableInternal=false,dockerRegistry.serverAddress=registry.localhost:5000,dockerRegistry.registryAddress=registry.localhost:5000,containers.manager.envs.functionBuildExecutorImage.value=${KANIKO_IMAGE}"
fi

export USE_ALPHA=${USE_ALPHA:-false}

export KYMA_SOURCES_DIR="./kyma"

host::create_registries_file() {
  cat >registries.yaml <<EOL
mirrors:
  registry.localhost:5000:
    endpoint:
    - http://registry.localhost:5000
  configs: {}
  auths: {}
  
EOL
}

install::kyma_cli() {
  local settings
  local kyma_version
  mkdir -p "/usr/local/bin"
  os=$(host::os)

  pushd "/usr/local/bin" || exit

  echo "Install kyma CLI ${os} locally to /usr/local/bin..."

  curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}?alt=media"
  chmod +x kyma
  kyma_version=$(kyma version --client)
  echo "Kyma CLI version: ${kyma_version}"

  echo "OK"

  popd || exit

  eval "${settings}"
}

host::os() {
  local host_os
  case "$(uname -s)" in
  Darwin)
    host_os=darwin
    ;;
  Linux)
    host_os=linux
    ;;
  *)
    echo >&2 -e "Unsupported host OS. Must be Linux or Mac OS X."
    exit 1
    ;;
  esac
  echo "${host_os}"
}

install::k3s() {
  echo "--> Installing k3s"
  curl -sfL https://get.k3s.io | K3S_KUBECONFIG_MODE=777 INSTALL_K3S_VERSION="v1.19.7+k3s1" INSTALL_K3S_EXEC="server --disable traefik" sh -
  mkdir -p ~/.kube
  cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
  chmod 600 ~/.kube/config
  k3s --version
  date
}

install::k3d() {
  echo "--> Installing k3d"
  curl -sfL https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash
}

function cluster-users::launch_tests() {
  echo "Running Kyma cluster-user tests"

  export ADMIN_EMAIL="admin@kyma.cx"
  export ADMIN_PASSWORD="read-only-user"
  export DEVELOPER_EMAIL="developer@kyma.cx"
  export DEVELOPER_PASSWORD="developer-user"
  export VIEW_EMAIL="read-only-user@kyma.cx"
  export VIEW_PASSWORD="read-only-user"
  export NAMESPACE_ADMIN_EMAIL="namespace.admin@kyma.cx"
  export NAMESPACE_ADMIN_PASSWORD="namespace-admin-user"
  export KYMA_SYSTEM="kyma-system"
  export NAMESPACE="default"

  # port-forward to be able to access dex
  kubectl -n kyma-system port-forward deployment/dex 5556:5556 &
  kubectl -n kyma-system port-forward deployment/iam-kubeconfig-service 8123:8000 &
  export IAM_KUBECONFIG_SVC_FQDN="https://localhost:8123"
  export DEX_SERVICE_SERVICE_HOST="localhost"
  export DEX_SERVICE_SERVICE_PORT_HTTP=5556

  pushd "${KYMA_SOURCES_DIR}/resources/cluster-users/files"
  if ! bash sar-test.sh; then
    echo "Tests failed"
    exit 1
  fi
  popd

  echo "Tests completed"
}

function cluster-users::prepare_components_file() {
  echo "Preparing Kyma installation "

  cat <<EOF >"$PWD/components.yaml"
components:
  - name: "cluster-essentials"
    namespace: "kyma-system"
  - name: "istio"
    namespace: "istio-system"
  - name: "xip-patch"
    namespace: "kyma-installer"
  - name: "dex"
    namespace: "kyma-system"
  - name: "ory"
    namespace: "kyma-system"
  - name: "iam-kubeconfig-service"
    namespace: "kyma-system"
  - name: "cluster-users"
    namespace: "kyma-system"
EOF
}


function install::kvm() {
  sleep 30
  sudo apt-get update
  sudo apt-get install qemu qemu-kvm qemu-system qemu-utils libvirt-clients libvirt-daemon-system virtinst -y
  sudo systemctl start libvirtd
  sudo systemctl status libvirtd
  sudo adduser "$(whoami)" libvirt
}

function install::docker() {
    sleep 30
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
}

function install::minikube() {
  curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
  sudo install minikube-linux-amd64 /usr/local/bin/minikube
}

date
echo "--> Kernel version"
uname -a

echo "--> Installing k3d for kyma-cli"
install::k3d

echo "--> Installing kyma-cli"
install::kyma_cli
echo "--> Installing minikube"
install::minikube
#cluster-users::prepare_components_file
install::kvm

echo "--> Provisioning k3s cluster for Kyma"
kyma provision minikube --kube-version 1.21.2 --vm-driver kvm2 --ci

echo "--> Deploying cluster-users"
kyma install --ci --src-path ${KYMA_SOURCES_DIR} --source=PR-"${PULL_NUMBER}"

echo "##############################################################################"
# shellcheck disable=SC2004
echo "# Kyma cluster-users installed in $(($SECONDS / 60)) min $(($SECONDS % 60)) sec"
echo "##############################################################################"

echo "Starting cluster-users tests"
cluster-users::launch_tests
