#!/usr/bin/env bash

install::kyma_cli() {
    mkdir -p "${INSTALL_DIR}/bin"
    export PATH="${INSTALL_DIR}/bin:${PATH}"
    os=$(host::os)

    pushd "${INSTALL_DIR}/bin"


    log::info "- Install kyma CLI ${os} locally to a tempdir..."

    curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}?alt=media"
    chmod +x kyma

    log::success "OK"

    popd
}