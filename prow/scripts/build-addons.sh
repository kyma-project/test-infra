#!/usr/bin/env bash

# This script is designed for the kyma-addons releasing process which is different from the Kyma releasing process.
set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=prow/scripts/lib/docker.sh
source "${SCRIPT_DIR}/lib/docker.sh"

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} /path/to/component"
    exit 1
}

readonly SOURCES_DIR=$1

if [[ -z "${SOURCES_DIR}" ]]; then
    usage
fi

docker::start

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    make -C "${SOURCES_DIR}" ci-pr
elif [[ "${BUILD_TYPE}" == "master" ]]; then
    make -C "${SOURCES_DIR}" ci-master
elif [[ "${BUILD_TYPE}" == "release" ]]; then
    make -C "${SOURCES_DIR}" ci-release
else
    echo "Not supported build type - ${BUILD_TYPE}"
    exit 1
fi