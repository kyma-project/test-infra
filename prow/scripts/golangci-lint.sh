#!/usr/bin/env bash

set -e

# shellcheck disable=SC2153
# PROJECT_SRC="${REPO_OWNER}/${REPO_NAME}"
COMPONENT_DEFINITION="go.mod"

function install_linter() {
    mkdir -p "/tmp/bin"
    export PATH="/tmp/bin:${PATH}"
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /tmp/bin v1.46.2
    golangci-lint --version
}

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
    pushd "${FOLDER}" # change to passed parameter

    golangci-lint  run ./...
    scan_result="$?"

    popd
    if [[ "$scan_result" != 0 ]]; then
        return 1
    else
        return 0
    fi
}

# TODO include that in base image in the first place
install_linter

# don't stop scans on first failure, but fail the whole job after all scans have finished
export scan_failed

echo "Starting Scan"

if [[ "$CREATE_SUBPROJECTS" == "true" ]]; then
    # treat every found Go project as a separate  project
    #pushd "${PROJECT_SRC}"  > /dev/null # change to passed parameter
    pwd

    # find all go.mod projects and scan them individually
    found_components=$(find . -name "$COMPONENT_DEFINITION" -not -path "./tests/*" -not -path "./docs/*" )


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
    #popd > /dev/null
else
    # scan PROJECT_SRC directory as a single project
    set +e
    scanFolder "." #${PROJECT_SRC}"
    scan_result="$?"
    set -e

    if [[ "$scan_result" -ne 0 ]]; then
        echo "Scan for $(pwd) has failed"
        scan_failed=1
    fi
fi

if [[ "$scan_failed" -eq 1 ]]; then
    echo "One or more of the scans have failed"
    exit 1
else
    echo "Scanning Finished"
fi
