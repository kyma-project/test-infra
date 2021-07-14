#!/usr/bin/env bash

set -e

date

export KYMA_SOURCES_DIR="./kyma"

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

function cluster-users::setup_env_vars() {
  echo "--> Setting up variables"
  ADMIN_EMAIL=$(kubectl -n kyma-system get secrets admin-user -o=jsonpath="{.data.email}" | base64 -d)
  echo "---> ADMIN_EMAIL: $ADMIN_EMAIL"
  export ADMIN_EMAIL
  ADMIN_PASSWORD=$(kubectl -n kyma-system get secrets admin-user -o=jsonpath="{.data.password}" | base64 -d)
  echo "---> ADMIN_PASSWORD: $ADMIN_PASSWORD"
  export ADMIN_PASSWORD
  DEVELOPER_EMAIL="$(kubectl -n kyma-system get secrets  test-developer-user -o=jsonpath="{.data.email}" | base64 -d)"
  echo "---> DEVELOPER_EMAIL: $DEVELOPER_EMAIL"
  export DEVELOPER_EMAIL
  DEVELOPER_PASSWORD="$(kubectl -n kyma-system get secrets test-developer-user -o=jsonpath="{.data.password}" | base64 -d)"
  echo "---> DEVELOPER_PASSWORD: $DEVELOPER_PASSWORD"
  export DEVELOPER_PASSWORD
  VIEW_EMAIL="$(kubectl -n kyma-system get secrets test-read-only-user -o=jsonpath="{.data.email}" | base64 -d)"
  echo "---> VIEW_EMAIL: $VIEW_EMAIL"
  export VIEW_EMAIL
  VIEW_PASSWORD="$(kubectl -n kyma-system get secrets test-read-only-user -o=jsonpath="{.data.password}" | base64 -d)"
  echo "---> VIEW_PASSWORD: $VIEW_PASSWORD"
  export VIEW_PASSWORD
  NAMESPACE_ADMIN_EMAIL="$(kubectl -n kyma-system get secrets test-namespace-admin-user -o=jsonpath="{.data.email}" | base64 -d)"
  echo "---> NAMESPACE_ADMIN_EMAIL: $NAMESPACE_ADMIN_EMAIL"
  export NAMESPACE_ADMIN_EMAIL
  NAMESPACE_ADMIN_PASSWORD="$(kubectl -n kyma-system get secrets test-namespace-admin-user -o=jsonpath="{.data.password}" | base64 -d)"
  echo "---> NAMESPACE_ADMIN_PASSWORD: $NAMESPACE_ADMIN_PASSWORD"
  export NAMESPACE_ADMIN_PASSWORD
  export KYMA_SYSTEM="kyma-system"
  export SYSTEM_NAMESPACE="kyma-system"
  # shellcheck disable=SC2155
  local namespace_id=$(< /dev/urandom tr -dc 'a-zA-Z0-9' | fold -w 5 | head -n 1 | tr '[:upper:]' '[:lower:]')
  CUSTOM_NAMESPACE="test-namespace-$namespace_id"
  echo "---> CUSTOM_NAMESPACE: $CUSTOM_NAMESPACE"
  export CUSTOM_NAMESPACE
  export NAMESPACE="default"
  export IAM_KUBECONFIG_SVC_FQDN="http://localhost:8123"
  export DEX_SERVICE_SERVICE_HOST="localhost"
  export DEX_SERVICE_SERVICE_PORT_HTTP=5556
}

function cluster-users::port-forward() {
  echo "--> Port forwarding"
  # cluster-user tests need to access dex. Previously tests were run from within cluster.
  kubectl -n kyma-system port-forward deployment/dex 5556:5556 &
  kubectl -n kyma-system port-forward deployment/iam-kubeconfig-service 8123:8000 &
  sleep 3s # waiting for port-forward to actually forward traffic.
}

function cluster-users::port-forward-cleanup() {
  echo "--> Closing port forwarding"
  pkill kubectl
}

function cluster-users::launch_tests() {
  echo "Running Kyma cluster-user tests"

  echo "--> Starting SAR testing"
  pushd "${KYMA_SOURCES_DIR}/resources/cluster-users/files"
  if ! bash sar-test.sh; then
    echo "Tests failed"
    exit 1
  fi
  popd

  echo "Tests completed"
}

function install::minikube() {
  curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
  sudo install minikube-linux-amd64 /usr/local/bin/minikube
}

date
echo "--> Kernel version"
uname -a

echo "--> Installing kyma-cli"
install::kyma_cli

echo "--> Installing minikube"
install::minikube

echo "--> Provisioning minikube cluster for Kyma"
STARTTIME=$(date +%s)
yes | kyma provision minikube \
               --ci \
               --vm-driver=none
ENDTIME=$(date +%s)
echo "  Provisioning time: $((ENDTIME - STARTTIME)) seconds."

echo "--> Installing Kyma"
STARTTIME=$(date +%s)
yes | kyma install \
     --ci \
     --source=PR-"${PULL_NUMBER}" \
     --src-path=${KYMA_SOURCES_DIR}
ENDTIME=$(date +%s)
echo "  Install time: $((ENDTIME - STARTTIME)) seconds."

echo "##############################################################################"
# shellcheck disable=SC2004
echo "# Kyma cluster-users installed in $(($SECONDS / 60)) min $(($SECONDS % 60)) sec"
echo "##############################################################################"

echo "Starting cluster-users tests"
cluster-users::port-forward
cluster-users::setup_env_vars
cluster-users::launch_tests
cluster-users::port-forward-cleanup

echo "--> Success"
