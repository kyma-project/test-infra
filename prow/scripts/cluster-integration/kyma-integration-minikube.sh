#!/usr/bin/env bash

set -e

readonly driver="none"
readonly testsuiteName="testsuite-all"

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
      >&2 echo -e "Unsupported host OS. Must be Linux or Mac OS X."
      exit 1
      ;;
  esac
  echo "${host_os}"
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
kyma provision minikube \
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
echo "  List test definitions"
kyma test definitions --ci
echo "  Run tests"
kyma test run \
          --ci \
          --watch \
          --max-retries=1 \
          --name="${testsuiteName}"

ENDTIME=$(date +%s)
echo "  Test time: $((ENDTIME - STARTTIME)) seconds."

echo "  Test summary"
kyma test status "${testsuiteName}" -owide
statusSucceeded=$(kubectl get cts "${testsuiteName}" -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
if [[ "${statusSucceeded}" != *"True"* ]]; then
  echo "- Fetching logs from testing pods in Failed status..."
  kyma test logs "${testsuiteName}" --test-status Failed

  echo "- Fetching logs from testing pods in Unknown status..."
  kyma test logs "${testsuiteName}" --test-status Unknown

  echo "- Fetching logs from testing pods in Running status due to running afer test suite timeout..."
  kyma test logs "${testsuiteName}" --test-status Running
  exit 1
fi
echo "  Generate junit results"
kyma test status "${testsuiteName}" -ojunit | sed 's/ (executions: [0-9]*)"/"/g' > junit_kyma_octopus-test-suite.xml
echo "--> Success"
