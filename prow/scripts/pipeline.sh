#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

if [[ -z "${SOURCES_DIR}" ]]; then
    echo "Missing SOURCES_DIR variable"
    exit 1
fi

init

cd "${SOURCES_DIR}"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    make ci-pr
elif [[ "${BUILD_TYPE}" == "master" ]]; then
    # TODO: Add support for release pipeline
    make ci-master
else
    echo "Not supported job type - ${JOB_TYPE}"
    exit 1
fi
