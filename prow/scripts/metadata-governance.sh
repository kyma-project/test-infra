#!/usr/bin/env bash


set -e
readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

readonly ARGS=("$@")
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
}

function run_metadata_validation_docker() {
    set +e

    docker run -v "${VOLUME_DIR}:/work" -w /work --rm "miy4/json-schema-validator" --syntax ${1}

    local result=$?
    if [[ ${result} -ne 0 ]]; then
        OUTPUT=1
    fi
    set -e
}

function validate_metadata_schema_on_pr() {
    echo "Fetching changes between origin/master and your branch"
    if [ -n "${PULL_NUMBER}" ]; then
        fetch_origin_master
    fi

    local files=""
    files=$(git --no-pager diff --name-only origin/master | grep 'values.schema.json' || echo '')

    if [ -n "${files}" ]; then
        VOLUME_DIR="${REPOSITORY_DIR}/temp"
        copy_files "${files}"

        local schemas=""
        for file in ${files}; do
            schemas="${schemas} /work/${file}"
        done
        run_metadata_validation_docker "${schemas}"
        rm -rf "${VOLUME_DIR}"
    else
        echo "No metadata files to validate"
    fi
}

function main() {
    read_arguments "${ARGS[@]}"
    init

    shout "Validate changed json schema files"
    validate_metadata_schema_on_pr

    exit ${OUTPUT}
}

main

