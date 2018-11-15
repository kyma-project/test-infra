#!/usr/bin/env bash

set -e
trap popd EXIT

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

if [[ -z "${SOURCES_DIR}" ]]; then
    echo "Missing SOURCES_DIR variable"
    exit 1
fi

function export_variables() {
    if [[ "${BUILD_TYPE}" == "pr" ]]; then
        DOCKER_TAG="PR-${PULL_NUMBER}"
    elif [[ "${BUILD_TYPE}" == "master" ]]; then
        DOCKER_TAG="$(git describe --tags --always)"
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

pushd "${SOURCES_DIR}"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    make ci-pr
elif [[ "${BUILD_TYPE}" == "master" ]]; then
    make ci-master
elif [[ "${BUILD_TYPE}" == "release" ]]; then
    make ci-release
else
    echo "Not supported build type - ${BUILD_TYPE}"
    exit 1
fi