#!/usr/bin/env bash

# Description: This script is responsible for downloading license files for node project dependencies

set -e

readonly ARGS=("$@")
readonly CWD=$PWD
readonly TMP_DIR_NAME=".license-puller"
readonly TMP_DIR="./${TMP_DIR_NAME}"
readonly LICENSES_DIR_NAME="licenses"
readonly LICENSES_DIR="./${LICENSES_DIR_NAME}"

DIRS_TO_PULLING=()
VERBOSE=false
function read_arguments() {
    for arg in "${ARGS[@]}"
    do
        case $arg in
            --dirs-to-pulling=*)
              local dirs_to_pulling=($( echo "${arg#*=}" | tr "," "\n" ))
              shift # remove --dirs-to-pulling=
            ;;
            --verbose)
              VERBOSE=true
              shift # remove this --verbose
            ;;
            *)
              # unknown option
            ;;
        esac
    done

    if [ "${#dirs_to_pulling[@]}" -ne 0 ]; then
        for d in "${dirs_to_pulling[@]}"; do
            DIRS_TO_PULLING+=($( cd "${CWD}/${d}" && pwd ))
        done
    fi
    readonly DIRS_TO_PULLING
}

function installNodeLicensesChecker() {
    echo "Installing license-checker for Node mode"
    npm i -g license-checker
}

function pullNodeLicenses() {
    if [ "${#DIRS_TO_PULLING[@]}" -ne 0 ]; then
        for d in "${DIRS_TO_PULLING[@]}"; do
            ( cd "${d}" && pullNodeLicensesByDir ) || true
        done
    fi

    ( cd "${CWD}" && pullNodeLicensesByDir ) || true
}

function pullNodeLicensesByDir() {
    echo "Gathering dependencies for $PWD"
    npx license-checker --production --json --direct --out "${TMP_DIR}/node.json"

    echo "Copying license files to '${LICENSES_DIR}'"
    # shellcheck disable=SC2016
    jq -r '. | keys[] as $key | [$key, (.[$key] | .licenseFile)] | @tsv' "${TMP_DIR}/node.json" |
        while IFS=$'\t' read -r key licenseFile; do
            if [[ -z "${licenseFile}" ]]; then
                continue
            fi

            local outputDir="${LICENSES_DIR}/${key}"
            if [ "$VERBOSE" = true ] ; then
                echo "Copying '${licenseFile}' to '${outputDir}'"
            fi
           
            mkdir -p "${outputDir}"
            cp "${licenseFile}" "${outputDir}/"
        done
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

    echo "Pulling licenses for Node dependencies"
    installNodeLicensesChecker
    pullNodeLicenses

    mergeLicenses
    removeTempFolders
}

main
