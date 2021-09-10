#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

date

export DOMAIN=${KYMA_DOMAIN:-local.kyma.dev}

export KYMA_SOURCES_DIR="./kyma"

function provision_k3d_and_run_testsuite() {
    pushd $KYMA_SOURCES_DIR
    echo 'git status:'
    git status
    echo 'git branch -v:'
    git branch -v
    echo 'git remote -v:'
    git remote -v
    #echo 'git pull:'
    #git pull
    echo "show filesystem:"
    ls -al $KYMA_SOURCES_DIR
    ls -al $KYMA_SOURCES_DIR/tests
    ls -al $KYMA_SOURCES_DIR/integration
    ls -al $KYMA_SOURCES_DIR/cluster-users
    popd
    pushd $KYMA_SOURCES_DIR/tests/integration/cluster-users
    ls -latr .
    bash k3d-cluster-users.sh
}
echo "running tests suite on k3d"
provision_k3d_and_run_testsuite

