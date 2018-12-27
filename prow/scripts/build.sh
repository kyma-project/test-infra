#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} /path/to/component"
    exit 1
}

readonly SOURCES_DIR=$1

if [[ -z "${SOURCES_DIR}" ]]; then
    usage
fi

function export_variables() {
    if [[ "${BUILD_TYPE}" == "pr" ]]; then
        DOCKER_TAG="PR-${PULL_NUMBER}"
    elif [[ "${BUILD_TYPE}" == "master" ]]; then
        DOCKER_TAG=$(echo "${PULL_BASE_SHA}" | cut -c1-8)
    elif [[ "${BUILD_TYPE}" == "release" ]]; then
        DOCKER_TAG=$(cat "${SCRIPT_DIR}/../RELEASE_VERSION")
        echo "Reading docker tag from RELEASE_VERSION file, got: ${DOCKER_TAG}"
    else
        echo "Not supported build type - ${BUILD_TYPE}"
        exit 1
    fi

    readonly DOCKER_TAG
    export DOCKER_TAG
}

init
export_variables

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    make -C "${SOURCES_DIR}" ci-pr
elif [[ "${BUILD_TYPE}" == "master" ]]; then
    make -C "${SOURCES_DIR}" ci-master
elif [[ "${BUILD_TYPE}" == "release" ]]; then
    make -C "${SOURCES_DIR}" ci-release
elif [[ "${BUILD_TYPE}" == "custom" ]]; then
    if [[ -z "${CUSTOM_MAKEFILE_RULE}" ]]; then
        echo "Missing required environment variable 'CUSTOM_MAKEFILE_RULE'"
        exit 1
    fi
    make -C "${SOURCES_DIR}" "${CUSTOM_MAKEFILE_RULE}"
else
    echo "Not supported build type - ${BUILD_TYPE}"
    exit 1
fi