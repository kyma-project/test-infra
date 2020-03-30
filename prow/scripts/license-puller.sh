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

LANGUAGE=
function init() {
    if [[ -f "package.json" ]]; then
        LANGUAGE=node
    else
        LANGUAGE=golang
    fi
    readonly LANGUAGE

    echo "Will work in '${LANGUAGE}' mode"
}

function pullGoLicenses() {
    if [ "${#DIRS_TO_PULLING[@]}" -ne 0 ]; then
        for d in "${DIRS_TO_PULLING[@]}"; do
            ( cd "${d}" && pullGoLicensesByDir ) || true
        done
    fi
    
    ( cd "${CWD}" && pullGoLicensesByDir ) || true
}

# TODO: This is temporary solution for Golang
function pullGoLicensesByDir() {
    echo "Gathering dependencies for $PWD"
    
    mkdir -p "${TMP_DIR}"
    go list -json ./... > "${TMP_DIR}/golang.json"

    echo "Downloading license files to '${LICENSES_DIR}'"
    # shellcheck disable=SC2016
    jq -sr '[{ data: map(.) } | .data[] | select(has("ImportMap")) | .ImportMap | keys[]] | unique | values[]' "${TMP_DIR}/golang.json" \
        | sed -e 's/sigs\.k8s\.io/github\.com\/kubernetes-sigs/g' \
        | sed -e 's/k8s\.io/github\.com\/kubernetes/g' \
        | grep -oE "^[^\/]+\/[^\/]+\/[^\/]+" \
        | sort -u \
        | while IFS=$'\t' read -r repository; do
            local outputDir="${LICENSES_DIR}/${repository}"
            mkdir -p "${outputDir}"

            downloadGoLicense "${outputDir}" "${repository}"
        done
}

# TODO: This is temporary solution for Golang
function downloadGoLicense() {
    local output=${1}
    local repository=${2}
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
    init
    mkdir -p "${TMP_DIR}"
    mkdir -p "${LICENSES_DIR}"

    if [[ ${LANGUAGE} == golang ]]; then
        echo "Pulling licenses for Golang dependencies"
        pullGoLicenses
    else
        echo "Pulling licenses for Node dependencies"
        installNodeLicensesChecker
        pullNodeLicenses
    fi

    mergeLicenses
    removeTempFolders
}

main
