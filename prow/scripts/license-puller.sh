#!/usr/bin/env bash

#Description: This script is responsible for downloading license files for dependencies

set -e

readonly TMP_DIR="./.license-puller"
readonly LICENSES_DIR="./licenses"

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

# TODO: This is temporary solution for Golang
function pullGoLicenses() {
    echo "Pulling licenses for Golang dependencies"

    echo "Gathering dependencies"
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

            downloadLicense "${outputDir}" "${repository}"
        done
}

# TODO: This is temporary solution for Golang
function downloadLicense() {
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

function pullNodeLicenses() {
    echo "Pulling licenses for Node dependencies"

    echo "Installing license-checker"
    npm i -g license-checker

    echo "Gathering dependencies"
    npx license-checker --production --json --direct --out "${TMP_DIR}/node.json"

    echo "Copying license files to '${LICENSES_DIR}'"
    # shellcheck disable=SC2016
    jq -r '. | keys[] as $key | [$key, (.[$key] | .licenseFile)] | @tsv' "${TMP_DIR}/node.json" |
        while IFS=$'\t' read -r key licenseFile; do
            if [[ -z "${licenseFile}" ]]; then
                continue
            fi

            local outputDir="${LICENSES_DIR}/${key}"
            echo "Copying '${licenseFile}' to '${outputDir}'"
            mkdir -p "${outputDir}"
            cp "${licenseFile}" "${outputDir}/"
        done
}

function main() {
    init
    mkdir -p "${TMP_DIR}"
    mkdir -p "${LICENSES_DIR}"

    if [[ ${LANGUAGE} == golang ]]; then
        pullGoLicenses
    else
        pullNodeLicenses
    fi

    rm -rf "${TMP_DIR}"
}

main
