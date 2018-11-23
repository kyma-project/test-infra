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
        echo "Calculating DOCKER_TAG variable for release..."
        branchPattern='^release-[0-9]+\.[0-9]+$'
        echo "${PULL_BASE_REF}" | grep -E -q "${branchPattern}"
        branchMatchesPattern=$?
        if [ ${branchMatchesPattern} -ne 0 ]
        then
            echo "Branch name does not match pattern: ${branchPattern}"
            exit 1
        fi

        version=${PULL_BASE_REF:8}
        # Getting last tag that matches version
        last=$(git tag --list "${version}.*" --sort "-version:refname" | head -1)

        if [ -z "$last" ]
        then
            newVersion="${version}.0"
        else
            tagPattern='^[0-9]+.[0-9]+.[0-9]+$'
            echo "${last}" | grep -E -q "${tagPattern}"
            lastTagMatches=$?
            if [ ${lastTagMatches} -ne 0 ]
            then
                echo "Last tag does not match pattern: ${tagPattern}"
                exit 1
            fi

            list=$(echo "${last}" | tr '.' ' ')
            vMajor=${list[0]}
            vMinor=${list[1]}
            vPatch=${list[2]}
            vPatch=$((vPatch + 1))
            newVersion="$vMajor.$vMinor.$vPatch"
        fi
        echo "New version is $newVersion"
        DOCKER_TAG=$newVersion

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
else
    echo "Not supported build type - ${BUILD_TYPE}"
    exit 1
fi