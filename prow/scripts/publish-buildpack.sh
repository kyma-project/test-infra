#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

if [[ -z "${SOURCES_DIR}" ]]; then
    echo "Missing SOURCES_DIR variable"
    exit 1
fi

function export_variables() {
    readonly DOCKER_TAG="$(date +v%Y%m%d)-$(git describe --tags --always --dirty)"
    export DOCKER_TAG
}

init
export_variables

pushd "${SOURCES_DIR}"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    make ci-pr
elif [[ "${BUILD_TYPE}" == "release" ]]; then
    make ci-release
else
    echo "Not supported build type - ${BUILD_TYPE}"
    exit 1
fi