#!/usr/bin/env bash

set -e

# scanFolder runs single folder through golangci-lint
# parameters:
# $1 - path to a folder to scan
# variables:
# function returns 0 on success or 1 on fail
function scanFolder() { # expects to get the fqdn of folder passed to scan
    if [[ $1 == "" ]]; then
        echo "path cannot be empty"
        exit 1
    fi
    FOLDER=$1
    pushd "${FOLDER}" > /dev/null # change to passed parameter

    golangci-lint  run ./... -v
    scan_result="$?"

    popd > /dev/null
    if [[ "$scan_result" != 0 ]]; then
        return 1
    else
        return 0
    fi
}

# don't stop scans on first failure, but fail the whole job after all scans have finished
export scan_failed=false

echo "Starting Scan"

pwd

# find all go.mod projects and scan them individually
found_components=$(find . -name "go.mod" -not -path "./tests/*" -not -path "./docs/*" )


while read -r component_definition_path; do
    # remove go.mod part
    component_path="${component_definition_path%/*}"
    # keep only the last directory in the tree as a name

    echo "- Linting $component_path"
    set +e
    scanFolder "${component_path}"
    scan_result="$?"
    set -e

    if [[ "$scan_result" -ne 0 ]]; then
        echo "Scan for ${FOLDER} has failed"
        scan_failed=1
    fi
done <<< "$found_components"

if [[ "$scan_failed" -eq 1 ]]; then
    echo "One or more of the scans have failed"
    exit 1
else
    echo "Scanning Finished"
fi
