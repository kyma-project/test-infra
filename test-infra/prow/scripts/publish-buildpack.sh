#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} /path/to/image"
    exit 1
}

readonly SOURCES_DIR=$1

if [[ -z "${SOURCES_DIR}" ]]; then
    usage
fi

function export_variables() {
    if [[ "${BUILD_TYPE}" == "pr" ]]; then
        DOCKER_TAG="PR-${PULL_NUMBER}"
    else
        DOCKER_TAG="$(date +v%Y%m%d)-$(git describe --tags --always --dirty)"
    fi
    readonly DOCKER_TAG
    export DOCKER_TAG
}

init
export_variables

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    make -C "${SOURCES_DIR}" ci-pr
elif [[ "${BUILD_TYPE}" == "release" ]]; then
    make -C "${SOURCES_DIR}" ci-release
else
    echo "Not supported build type - ${BUILD_TYPE}"
    exit 1
fi