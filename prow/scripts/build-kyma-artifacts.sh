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
   DOCKER_TAG=$(cat "${SCRIPT_DIR}/../RELEASE_VERSION")
   echo "Reading docker tag from RELEASE_VERSION file, got: ${DOCKER_TAG}"
}

init
export_variables

echo "LS -la logs"
ls -la /logs

make -C "${SOURCES_DIR}" ci-create-release-artifacts