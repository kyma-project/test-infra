#!/usr/bin/env bash

set -e

function integration_tests::install_kyma() {
  log::info "Installing Kyma from local source using components file"
  kyma deploy --ci --components-file "$PWD/components.yaml" --source=local --workspace "${KYMA_SOURCES_DIR}" --verbose
}

function istio::get_version() {
  pushd "${KYMA_SOURCES_DIR}"
  istio_version=$(git show "${KYMA_VERSION}:resources/istio/Chart.yaml" | grep appVersion | sed -n "s/appVersion: //p")
  popd
}
