#!/usr/bin/env bash

# Description: This script is responsible for downloading license files for dependencies
set -e

readonly ARGS=("$@")
readonly CWD=$PWD
readonly TMP_DIR_NAME=".license-puller"
readonly TMP_DIR="./${TMP_DIR_NAME}"
readonly LICENSES_DIR_NAME="licenses"
readonly LICENSES_DIR="./${LICENSES_DIR_NAME}"

DIRS_TO_PULLING=()
function read_arguments() {
    for arg in "${ARGS[@]}"
    do
        case $arg in
            --dirs-to-pulling=*)
              local dirs_to_pulling=()
              IFS="," read -r -a dirs_to_pulling <<< "${arg#*=}"
              shift # remove --dirs-to-pulling=
            ;;
            *)
              # unknown option
            ;;
        esac
    done

    if [ "${#dirs_to_pulling[@]}" -ne 0 ]; then
        for d in "${dirs_to_pulling[@]}"; do
            DIRS_TO_PULLING+=( "$( cd "${CWD}/${d}" && pwd )" )
        done
    fi
    readonly DIRS_TO_PULLING
}

function pullLicenses() {
    if [ "${#DIRS_TO_PULLING[@]}" -ne 0 ]; then
        for d in "${DIRS_TO_PULLING[@]}"; do
            ( cd "${d}" && pullLicensesByDir ) || true
        done
    fi
    
    ( cd "${CWD}" && pullLicensesByDir ) || true
}

# TODO: This is temporary solution for Golang
function pullLicensesByDir() {
    echo "Gathering dependencies for $PWD"
    
    mkdir -p "${TMP_DIR}"
    go list -json ./... > "${TMP_DIR}/golang.json"

    echo "Downloading license files to '${LICENSES_DIR}'"
    # shellcheck disable=SC2016
    jq -sr '[{ data: map(.) } | .data[] | select(has("ImportMap")) | .ImportMap | keys[]] | unique | values[]' "${TMP_DIR}/golang.json" \
        | grep -oE "^[^\/]+\/[^\/]+\/[^\/]+" \
        | sort -u \
        | while IFS=$'\t' read -r repository; do
            local outputDir="${LICENSES_DIR}/${repository}"
            mkdir -p "${outputDir}"

            downloadLicense "${outputDir}" "${repository}"
        done
}

# TODO: This is temporary solution for Golang
function downloadLicense() {
    local output=${1}
    local importPath=${2}

    # laymans vanity-import support
    local repository
    repository=$(curl -L "${importPath}?go-get=1" | pup 'meta[name="go-import"] attr{content}' | paste -sd "," - | awk '{print $3}' | sed 's/.git$// ; s%^[^:]\+://%%')
    local url="https://${repository/github.com/raw.githubusercontent.com}/master"

    echo "Downloading license from '${repository}' to '${output}''"
    for file in "LICENSE" "LICENSE.md" "LICENSE.txt" "UNLICENSE"; do
        if [[ "$(ls -A "${output}")" ]]; then
            break
        fi

        echo "  Trying with '${file}' file..."
        set +e
        curl "${url}/${file}" -sL --fail --output "${output}/${file}"
        set -e
    done

    if [[ -z "$(ls -A "${output}")" ]]; then
        echo "  Cannot find license file for '${repository}'"
        rm -rf "${output}"
    else
        echo "  Downloaded"
    fi
}

function mergeLicenses() {
    if [ "${#DIRS_TO_PULLING[@]}" -ne 0 ]; then
        for d in "${DIRS_TO_PULLING[@]}"; do
            echo "Merging licenses from ${d} to ${CWD}"
            cp -R "${d}/${LICENSES_DIR_NAME}/" "${CWD}/${LICENSES_DIR_NAME}/" || true
        done
    fi
}

function removeTempFolders() {
    if [ "${#DIRS_TO_PULLING[@]}" -ne 0 ]; then
        for d in "${DIRS_TO_PULLING[@]}"; do
            rm -rf "${d:?}/${TMP_DIR}" || true
            rm -rf "${d:?}/${LICENSES_DIR}" || true
        done
    fi

    rm -rf "${CWD:?}/${TMP_DIR}" || true
}

function main() {
    read_arguments "${ARGS[@]}"
    mkdir -p "${TMP_DIR}"
    mkdir -p "${LICENSES_DIR}"

    pullLicenses
    mergeLicenses
    removeTempFolders
}

main
