#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

date

export DOMAIN=${KYMA_DOMAIN:-local.kyma.dev}

export KYMA_SOURCES_DIR="./kyma"

function _tmp_get_kyma_sources() {
    git clone https://github.com/piotrkpc/kyma tmp_kyma
    pushd tmp_kyma
    git checkout cluster-users-hydra
    KYMA_SOURCES_DIR="./tmp_kyma"
    popd
}

function provision_k3d_and_run_testsuite() {
    pushd $KYMA_SOURCES_DIR/tests/integration/cluster-users
    bash k3d-cluster-users.sh
}

_tmp_get_kyma_sources # this should be remove after https://github.com/kyma-project/kyma/pull/12052 is merged
echo "running tests suite on k3d"
provision_k3d_and_run_testsuite
