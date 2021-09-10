#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

date

export DOMAIN=${KYMA_DOMAIN:-local.kyma.dev}

export KYMA_SOURCES_DIR="./kyma"

function provision_k3d_and_run_testsuite() {
    pushd $KYMA_SOURCES_DIR/tests/integration/cluster-users
    bash k3d-cluster-users.sh
    popd
}
echo "running tests suite on k3d"
provision_k3d_and_run_testsuite

