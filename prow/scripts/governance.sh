#!/usr/bin/env bash

readonly MILV_IMAGE="magicmatatjahu/milv:0.0.6"
readonly REPOSITORY_NAME=$1
readonly BRANCH_NAME=$(git branch | grep \* | cut -d ' ' -f2)

if [[ -z "${REPOSITORY_NAME}" ]]; then
    echo "ERROR: repository name is required"
    exit 1
fi

function run_milv_docker() {
    docker run --rm --dns=8.8.8.8 --dns=8.8.4.4 -v $PWD:/${REPOSITORY_NAME}:ro ${MILV_IMAGE} --base-path=/${REPOSITORY_NAME} $1
}

function validate_internal_links() {
    echo "Validate internal links"
    run_milv_docker "--ignore-external"
}

function validate_external_links() {
    echo "Fetching changes between origin/master...origin/${BRANCH_NAME}"
    local files=$(git --no-pager diff --name-only origin/master...origin/${BRANCH_NAME} | grep '.md' || echo '')
    local changed_files=""
    for file in $files; do
        changed_files="${changed_files} ${file}"
    done

    echo "Validate external links in changed markdown files"
    if [ "${changed_files}" = "-lt 1" ]; then
        run_milv_docker "--ignore-internal ${changed_files}"
    else
        echo "Any markdown files to checking external links"
    fi
}

validate_internal_links
validate_external_links
