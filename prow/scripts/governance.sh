#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

readonly ARGS=("$@")
readonly MILV_IMAGE="eu.gcr.io/kyma-project/external/magicmatatjahu/milv:0.0.7-beta"
VOLUME_DIR=""
OUTPUT=0

function read_arguments() {
    # read arguments
    while test $# -gt 0; do
        case "$1" in
            --repository)
                shift
                readonly REPOSITORY_NAME=$1
                shift
                ;;
            --repository-org)
                shift
                readonly REPOSITORY_ORG=$1
                shift
                ;;
            --repository-dir)
                shift
                readonly REPOSITORY_DIR=$1
                shift
                ;;
            --full-validation)
                shift
                readonly FULL_VALIDATION=$1
                shift
                ;;
            *)
                echo "$1 is not a recognized flag!"
                exit 1;
                ;;
        esac
    done

    if [[ -z "${REPOSITORY_NAME}" ]]; then
        echo -e "ERROR: repository name is required"
        exit 1
    fi

    if [[ -z "${REPOSITORY_ORG}" ]]; then
        REPOSITORY_ORG="kyma-project"
    fi

    if [[ -z "${REPOSITORY_DIR}" ]]; then
        REPOSITORY_DIR="${PWD}"
    fi

    VOLUME_DIR="${REPOSITORY_DIR}"
}

function fetch_origin_master() {
    local repository="https://github.com/${REPOSITORY_ORG}/${REPOSITORY_NAME}.git"
    git remote add origin "${repository}"
    git fetch origin master
}

function copy_files() {
    mkdir -p "${VOLUME_DIR}"

    for file in $1; do
        if [[ ! -f "${file}" ]]; then
            echo "Skipping deleted file ${file}..."
            continue
        fi
        
        mkdir -p "${VOLUME_DIR}/$(dirname "${file}")"
        cp -rf "${file}" "${VOLUME_DIR}/${file}"
    done

    cp -rf milv.config.yaml "${VOLUME_DIR}"/milv.config.yaml
}

function run_milv_docker() {
    docker run --rm --dns=8.8.8.8 --dns=8.8.4.4 -v "${VOLUME_DIR}:/${REPOSITORY_NAME}:ro" "${MILV_IMAGE}" --base-path="/${REPOSITORY_NAME}" "${@}"

    local result=$?
    if [ ${result} != 0 ]; then
        OUTPUT=1
    fi
}

function validate_internal() {
    run_milv_docker --ignore-external
}

function validate_external() {
    run_milv_docker --ignore-internal
}

function validate_external_on_pr() {
    echo "Fetching changes between origin/master and your branch"
    if [ -n "${PULL_NUMBER}" ]; then
        fetch_origin_master
    fi

    local files=""
    files=$(git --no-pager diff --name-only origin/master | grep '.md' || echo '')

    if [ -n "${files}" ]; then
        VOLUME_DIR="${REPOSITORY_DIR}/temp"
        copy_files "${files}"

        validate_external
        rm -rf "${VOLUME_DIR}"
    else
        echo "Any markdown files to checking external links"
    fi
}

function main() {
    read_arguments "${ARGS[@]}"
    init

    shout "Validate internal links"
    validate_internal

    if [ "${FULL_VALIDATION}" == true ]; then
        shout "Validate external links"
        validate_external
    else
        shout "Validate external links on changed markdown files"
        validate_external_on_pr
    fi

    exit ${OUTPUT}
}
main
