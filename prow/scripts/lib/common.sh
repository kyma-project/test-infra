#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"
# shellcheck source=prow/scripts/lib/gcloud.sh
source ${LIBDIR}/gcloud.sh
# shellcheck source=prow/scripts/lib/github.sh
source ${LIBDIR}/github.sh
# shellcheck source=prow/scripts/lib/docker.sh
source ${LIBDIR}/docker.sh

# common::init initalizes docker in docker, authenticates with gcloud and setsup git correctly
function common::init() {
    echo "Initializing"

    if [[ -n "${GOOGLE_APPLICATION_CREDENTIALS}" ]]; then
        gcloud::authenticate
    fi

    if [[ "${DOCKER_IN_DOCKER_ENABLED}" == true ]]; then
        docker::start
    fi

    if [[ -n "${BOT_GITHUB_SSH_PATH}" ]] || [[ -n "${BOT_GITHUB_EMAIL}" ]] || [[ -n "${BOT_GITHUB_NAME}" ]]; then
        configure_git
    fi
}
