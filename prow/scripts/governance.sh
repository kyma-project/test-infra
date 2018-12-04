#!/usr/bin/env bash

set -e

readonly MILV_IMAGE="magicmatatjahu/milv:0.0.6"
readonly REPOSITORY_NAME=$1
readonly BRANCH_NAME=$(git branch | grep \* | cut -d ' ' -f2)

if [[ -z "${REPOSITORY_NAME}" ]]; then
    echo "ERROR: repository name is required"
    exit 1
fi

echo "${BRANCH_NAME}"

function run_milv_docker() {
    docker run --rm --dns=8.8.8.8 --dns=8.8.4.4 -v $PWD:/${REPOSITORY_NAME}:ro ${MILV_IMAGE} --base-path=/${REPOSITORY_NAME} $1
}

function changeset() {
    echo "Fetching changes between origin/${BRANCH_NAME}/head and origin/master."
    return $(git --no-pager diff --name-only origin/master...origin/${BRANCH_NAME}/head | grep '.md' || echo '')
}

function validate_internal_links() {
    echo "Validate internal links"
    run_milv_docker --ignore-external
}

changeset

# function validate_external_links() {
#     echo "Validate external links in changed markdown files"
#     readonly CHANGED_FILES=changed_markdown_files
#     run_milv_docker --ignore-internal ${CHANGED_FILES}
# }

# function changed_markdown_files() {

# }

validate_internal_links
# validate_external_links