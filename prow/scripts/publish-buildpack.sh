#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=prow/scripts/lib/docker.sh
source "${SCRIPT_DIR}/lib/docker.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$SCRIPT_DIR/lib/gcp.sh"

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} /path/to/image"
    exit 1
}

readonly SOURCES_DIR=$1

if [[ -z "${SOURCES_DIR}" ]]; then
    usage
fi

function export_variables() {
  log::info "Export variables for docker image tag"
    if [[ "${BUILD_TYPE}" == "pr" ]]; then
        DOCKER_TAG="PR-${PULL_NUMBER}"
    else
        DOCKER_TAG="$(date +v%Y%m%d)-$(git describe --tags --always --dirty)"
        DOCKER_POST_PR_TAG="$(/prow-tools/prtagbuilder)"
        readonly DOCKER_POST_PR_TAG
        export DOCKER_POST_PR_TAG
    fi
    readonly DOCKER_TAG
    export DOCKER_TAG
}

gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"
docker::start
export_variables

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    make -C "${SOURCES_DIR}" ci-pr
elif [[ "${BUILD_TYPE}" == "release" ]]; then
    make -C "${SOURCES_DIR}" ci-release
else
    echo "Not supported build type - ${BUILD_TYPE}"
    exit 1
fi
log::info "Publish build pack's done"