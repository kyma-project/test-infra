#!/usr/bin/env bash

function docker::start {
    if docker info > /dev/null 2>&1 ; then
        echo "Docker already started"
        return 0
    fi

    echo "Initializing Docker..."
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
            return 1
        fi
    done
    printf '=%.0s' {1..80}; echo

    docker-credential-gcr configure-docker
    echo "Done starting up docker."
}

function docker::print_processes {
    docker ps -a
}