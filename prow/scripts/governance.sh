#!/usr/bin/env bash

readonly MILV_IMAGE="magicmatatjahu/milv:0.0.6"
readonly BRANCH_NAME=$(git branch | cut -d ' ' -f2)
readonly REPOSITORY_NAME=$1
readonly NIGHTLY_VALIDATION=$2
OUTPUT=0

if [[ -z "${REPOSITORY_NAME}" ]]; then
    echo -e "ERROR: repository name is required"
    exit 1
fi

function run_milv_docker() {
    docker run --rm --dns=8.8.8.8 --dns=8.8.4.4 -v "${PWD}":/"${REPOSITORY_NAME}":ro "${MILV_IMAGE}" --base-path=/"${REPOSITORY_NAME}" "$1"
    local result=$?
    if [ ${result} != 0 ]; then
        OUTPUT=1
    fi
}

function validate_internal_links() {
    echo "Validate internal links"
    run_milv_docker "--ignore-external"
}

function validate_external_links() {
    echo "Validate external links"
    run_milv_docker "--ignore-internal"
}

function validate_external_links_on_changed_files() {
    echo "Fetching changes between origin/master...origin/${BRANCH_NAME}"
    local files=""
    files=$(git --no-pager diff --name-only origin/master...origin/"${BRANCH_NAME}" | grep '.md' || echo '')
    local changed_files=""
    for file in $files; do
        changed_files="${changed_files} ${file}"
    done

    if [ -n "${changed_files}" ]; then
        echo "Validate external links in changed markdown files"
        run_milv_docker "--ignore-internal ${changed_files}"
    else
        echo "Any markdown files to checking external links"
    fi
}

if [ -n "${NIGHTLY_VALIDATION}" ]; then
    validate_external_links
else
    validate_internal_links
    validate_external_links_on_changed_files
fi

exit ${OUTPUT}
