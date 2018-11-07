#!/usr/bin/env bash

function start_docker() {
    echo "Docker in Docker enabled, initializing..."
    printf '=%.0s' {1..80}; echo
    # If we have opted in to docker in docker, start the docker daemon,
    service docker start
    # the service can be started but the docker socket not ready, wait for ready
    local WAIT_N=0
    local MAX_WAIT=20
    while true; do
        # docker ps -q should only work if the daemon is ready
        docker ps -q > /dev/null 2>&1 && break
        if [[ ${WAIT_N} -lt ${MAX_WAIT} ]]; then
            WAIT_N=$((WAIT_N+1))
            echo "Waiting for docker to be ready, sleeping for ${WAIT_N} seconds."
            sleep ${WAIT_N}
        else
            echo "Reached maximum attempts, not waiting any longer..."
            break
        fi
    done
    printf '=%.0s' {1..80}; echo
    
    docker-credential-gcr configure-docker
    echo "Done setting up docker in docker."
}

function authenticate() {
    echo "Authenticating"
    gcloud auth activate-service-account --key-file "${GOOGLE_APPLICATION_CREDENTIALS}"
}

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

function init() {
    echo "Initializing"

    if [[ ! -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]]; then
        authenticate
    fi

    if [[ "${DOCKER_IN_DOCKER_ENABLED}" == true ]]; then
        start_docker
    fi

    export_variables
}