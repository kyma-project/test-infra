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
        shout "Execute Job Guard for Release jobs"
        export TIMEOUT="75m"
        export JOB_NAME_PATTERN="(^pre-rel\\d\\d\\d-kyma-integration$)"
        "${SCRIPT_DIR}/../../development/tools/cmd/jobguard/run.sh"
        # TODO: Improve this part
        if [[ ( "${REPO_OWNER}" == "kyma-project" && ("${REPO_NAME}" == "kyma" || "${REPO_NAME}" == "test-infra") ) || "${REPO_OWNER}" == "kyma-incubator" && "${REPO_NAME}" == "compass" ]]; then
            DOCKER_TAG=$(cat "${SCRIPT_DIR}/../RELEASE_VERSION")
            echo "Reading docker tag from RELEASE_VERSION file, got: ${DOCKER_TAG}"
        else 
            DOCKER_TAG="${PULL_BASE_REF}"
        fi
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
    if [[ "${REPO_OWNER}" == "kyma-project" && "${REPO_NAME}" == "kyma" ]]; then
        NEXT_RELEASE=$(cat "${SCRIPT_DIR}/../RELEASE_VERSION")
        echo "Checking if ${NEXT_RELEASE} was already published on github..."
        RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" https://api.github.com/repos/kyma-project/kyma/releases/tags/"${NEXT_RELEASE}")
        if [[ $RESPONSE != 404* ]]; then
            echo "The ${NEXT_RELEASE} is already published on github. Stopping."
            exit 1
        fi
    fi
    make -C "${SOURCES_DIR}" ci-release
else
    echo "Not supported build type - ${BUILD_TYPE}"
    exit 1
fi
