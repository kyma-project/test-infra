#!/usr/bin/env bash

install::kyma_cli() {
    mkdir -p "${INSTALL_DIR}/bin"
    export PATH="${INSTALL_DIR}/bin:${PATH}"
    os=$(host::os)

    pushd "${INSTALL_DIR}/bin"

    log::info "- Install kyma CLI ${os} locally to a tempdir..."
#TODO
#    curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}?alt=media"
    curl -sSLo kyma "https://www.dropbox.com/s/iias1qtuhwflewi/kyma-8cbb19a?dl=0"
    chmod +x kyma

    log::success "OK"

    popd
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
      log::error "Unsupported host OS. Must be Linux or Mac OS X."
      exit 1
      ;;
  esac
  echo "${host_os}"
}
