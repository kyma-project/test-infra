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
            exit 1
        fi
    done
    printf '=%.0s' {1..80}; echo

    docker-credential-gcr configure-docker
    echo "Done setting up docker in docker."
}

function authenticate() {
    echo "Authenticating"
    gcloud auth activate-service-account --key-file "${GOOGLE_APPLICATION_CREDENTIALS}" || exit 1
}

function configure_git() {
    echo "Configuring git"
    # configure ssh
    if [[ ! -z "${BOT_GITHUB_SSH_PATH}" ]]; then
        mkdir "${HOME}/.ssh/"
        cp "${BOT_GITHUB_SSH_PATH}" "${HOME}/.ssh/ssh_key.pem"
        local SSH_FILE="${HOME}/.ssh/ssh_key.pem"
        touch "${HOME}/.ssh/known_hosts"
        ssh-keyscan -H github.com >> "${HOME}/.ssh/known_hosts"
        chmod 400 "${SSH_FILE}"
        eval "$(ssh-agent -s)"
        ssh-add "${SSH_FILE}"
        ssh-add -l
        git config --global core.sshCommand "ssh -i ${SSH_FILE}"
    fi

    # configure email
    if [[ ! -z "${BOT_GITHUB_EMAIL}" ]]; then
        git config --global user.email "${BOT_GITHUB_EMAIL}"
    fi
    
    # configure name
    if [[ ! -z "${BOT_GITHUB_NAME}" ]]; then
        git config --global user.name "${BOT_GITHUB_NAME}"
    fi
}

function init() {
    echo "Initializing"

    if [[ ! -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]]; then
        authenticate
    fi

    if [[ "${DOCKER_IN_DOCKER_ENABLED}" == true ]]; then
        start_docker
    fi

    if [[ ! -z "${BOT_GITHUB_SSH_PATH}" ]] || [[ ! -z "${BOT_GITHUB_EMAIL}" ]] || [[ ! -z "${BOT_GITHUB_NAME}" ]]; then
        configure_git
    fi
}

function shout() {
    echo "
#################################################################################################
# $1
#################################################################################################
    "
}
