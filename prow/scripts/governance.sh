#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

readonly MILV_IMAGE="magicmatatjahu/milv:0.0.6"
OUTPUT=0

# read arguments
while test $# -gt 0; do
    case "$1" in
        --repository)
            shift
            readonly REPOSITORY_NAME=$1
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

if [[ -z "${REPOSITORY_DIR}" ]]; then
    REPOSITORY_DIR="${PWD}"
fi

function run_milv_docker() {
    docker run --rm --dns=8.8.8.8 --dns=8.8.4.4 -v "${REPOSITORY_DIR}:/${REPOSITORY_NAME}:ro" "${MILV_IMAGE}" --base-path="/${REPOSITORY_NAME}" "${@}"
    local result=$?
    if [ ${result} != 0 ]; then
        OUTPUT=1
    fi
}

function validate_internal_links() {
    echo "Validate internal links"
    run_milv_docker --ignore-external
}

function validate_external_links() {
    echo "Validate external links"
    run_milv_docker --ignore-internal
}

function validate_external_links_on_changed_files() {
    local branch_name=""
    branch_name=$(git branch | cut -d ' ' -f2)
    echo "Fetching changes between origin/master...origin/${branch_name}"
    local files=""
    files=$(git --no-pager diff --name-only origin/master...origin/"${branch_name}" | grep '.md' || echo '')
    local changed_files=""
    for file in $files; do
        changed_files="${changed_files} ./${REPOSITORY_NAME}/${file}"
    done

    if [ -n "${changed_files}" ]; then
        echo "Validate external links in changed markdown files"
        # shellcheck disable=SC2086
        run_milv_docker --ignore-internal ${changed_files}
    else
        echo "Any markdown files to checking external links"
    fi
}

init
validate_internal_links

if [ "${FULL_VALIDATION}" == true ]; then
    validate_external_links
else
    validate_external_links_on_changed_files
fi

exit ${OUTPUT}
