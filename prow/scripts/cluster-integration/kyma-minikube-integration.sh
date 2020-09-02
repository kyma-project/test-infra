#!/usr/bin/env bash

set -e

driver="none"
testsuite_name="testsuite-all"

installKymaCLI() {
    local settings
    local kyma_version
    mkdir -p "$HOME/bin"
    export PATH="$HOME/bin:${PATH}"
    os=$(host::os)

    pushd "$HOME/bin" || exit

    echo "Install kyma CLI ${os} locally to $HOME/bin..."

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
  installKymaCLI
else
  echo "Kyma CLI is already installed"
  kyma_version=$(kyma version --client)
  echo "Kyma CLI version: ${kyma_version}"
fi
echo "--> Done"

echo "--> Provision Kyma cluster on minikube using VM driver ${driver}"
kyma provision minikube --ci --vm-driver="${driver}" --non-interactive
echo "--> Done"

echo "--> Installing Kyma on minikube cluster"
kyma install --ci --source="local" --src-path=./kyma --non-interactive
echo "--> Done"

echo "--> Run kyma tests"
echo "  List test definitions"
kyma test definitions --ci
echo "  Run tests"
kyma test run \
          --ci \
          --watch \
          --max-retries=1 \
          --non-interactive \
          --name="${testsuite-all}"

kyma test status "${SUITE_NAME}" -owide
echo "--> Success"
