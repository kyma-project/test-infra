#!/usr/bin/env bash

install::kyma_cli() {
    local settings
    local kyma_version
    settings="$(set +o); set -$-"
    set -u
    mkdir -p "${INSTALL_DIR}/bin"
    export PATH="${INSTALL_DIR}/bin:${PATH}"
    os=$(host::os)

    pushd "${INSTALL_DIR}/bin" || exit

    echo "--> Install kyma CLI ${os} locally to ${INSTALL_DIR}/bin"

    curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}?alt=media"
    chmod +x kyma
    kyma_version=$(kyma version --client)
    echo "--> Kyma CLI version: ${kyma_version}"
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
      echo "Unsupported host OS. Must be Linux or Mac OS X."
      exit 1
      ;;
  esac
  echo "${host_os}"
}
