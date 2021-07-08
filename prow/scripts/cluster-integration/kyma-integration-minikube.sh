#!/usr/bin/env bash

set -e

readonly driver="none"
readonly testSuiteName=${TEST_SUITE:-"testsuite-all"}
KYMA_TEST_TIMEOUT=${KYMA_TEST_TIMEOUT:=1h}

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

function testsuite::testsuite-all() {
  echo "  List test definitions"
  kyma test definitions --ci
  echo "  Run tests"
  kyma test run \
    --ci \
    --watch \
    --max-retries=1 \
    --name="${testSuiteName}" \
    --timeout="${KYMA_TEST_TIMEOUT}" \
    --selector kyma-project.io/test.integration=true

  ENDTIME=$(date +%s)
  echo "  Test time: $((ENDTIME - STARTTIME)) seconds."

  echo "  Test summary"
  kyma test status "${testSuiteName}" -owide
  statusSucceeded=$(kubectl get cts "${testSuiteName}" -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")

  # disable exit on error to get as many logs as possible
  set +e

  echo "  Generate junit results"
  kyma test status "${testSuiteName}" -ojunit | sed 's/ (executions: [0-9]*)"/"/g' >junit_kyma_octopus-test-suite.xml
  if [[ "${statusSucceeded}" != *"True"* ]]; then
    echo "- Fetching logs from testing pods in Failed status..."
    kyma test logs "${testSuiteName}" --test-status Failed

    echo "- Fetching logs from testing pods in Unknown status..."
    kyma test logs "${testSuiteName}" --test-status Unknown

    echo "- Fetching logs from testing pods in Running status due to running afer test suite timeout..."
    kyma test logs "${testSuiteName}" --test-status Running
    exit 1
  fi
}

function cluster-users::setup_env_vars() {
  echo "--> Setting up variables"
  ADMIN_EMAIL=$(kubectl -n kyma-system get secrets admin-user -o=jsonpath="{.data.email}" | base64 -d)
  echo "---> ADMIN_EMAIL: $ADMIN_EMAIL"
  export ADMIN_EMAIL
  ADMIN_PASSWORD=$(kubectl -n kyma-system get secrets admin-user -o=jsonpath="{.data.password}" | base64 -d)
  echo "---> ADMIN_PASSWORD: $ADMIN_PASSWORD"
  export ADMIN_PASSWORD
  DEVELOPER_EMAIL="$(kubectl -n kyma-system get secrets test-developer-user -o=jsonpath="{.data.email}" | base64 -d)"
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
  local namespace_id=$(tr </dev/urandom -dc 'a-zA-Z0-9' | fold -w 5 | head -n 1 | tr '[:upper:]' '[:lower:]')
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

function testsuite::cluster-users() {

  echo "Starting cluster-users tests"
  cluster-users::port-forward
  cluster-users::setup_env_vars
  cluster-users::launch_tests
  cluster-users::port-forward-cleanup

}

echo "--> Installing Kyma CLI"
if ! [[ -x "$(command -v kyma)" ]]; then
  echo "Kyma CLI not found"
  install::kyma_cli
else
  echo "Kyma CLI is already installed"
  kyma_version=$(kyma version --client)
  echo "Kyma CLI version: ${kyma_version}"
  minikube version
fi
echo "--> Done"

echo "--> Provision Kyma cluster on minikube using VM driver ${driver}"
STARTTIME=$(date +%s)
yes | kyma provision minikube \
  --ci \
  --vm-driver="${driver}"
ENDTIME=$(date +%s)
echo "  Execution time: $((ENDTIME - STARTTIME)) seconds."
echo "--> Done"

echo "--> Installing Kyma on minikube cluster"
STARTTIME=$(date +%s)
yes | kyma install \
  --ci \
  --source="local" \
  --src-path=./kyma
ENDTIME=$(date +%s)
echo "  Install time: $((ENDTIME - STARTTIME)) seconds."
echo "--> Done"

echo "--> Run kyma tests"
STARTTIME=$(date +%s)

case ${testSuiteName} in
"testsuite-all")
  testsuite::testsuite-all
  ;;
"cluster-users")
  testsuite::cluster-users
  ;;
*)
  echo "testsuite not selected"
  exit 1
  ;;
esac

echo "--> Success"
