#!/usr/bin/env bash

################################################################################
# DO NOT INVOKE DIRECTLY, USE FROM OTHER SCRIPTS
################################################################################

################################################################################
# Generic resource cleanup tool launcher
# This script invokes a cleanup tool specified by input arguments.
# It passes all input parameters, unchanged, to the target tool.
# The script also checks for existence of the mandatory
# GOOGLE_APPLICATION_CREDENTIALS environment variable.
#
# Input parameters (positional):
# $1 - tool subdirectory in the existing hierarchy: ../development/tools/cmd/<tool-directory>/
# $2 - name of the object subject to cleanup, for nice log messages.
################################################################################

set -e
set -o pipefail

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then
   echo "GOOGLE_APPLICATION_CREDENTIALS environment variable is missing!"
   exit 1
fi

readonly TOOL_DIR="$1"
if [ -z "${TOOL_DIR}" ]; then
    echo "TOOL_DIR variable is missing!"
		exit 1
fi

readonly OBJECT_NAME="$2"
if [ -z "${OBJECT_NAME}" ]; then
    echo "OBJECT_NAME variable is missing!"
		exit 1
fi

shift #pass $1
shift #pass $2

echo "--------------------------------------------------------------------------------"
echo "Removing GCP ${OBJECT_NAME} allocated by failed/terminated integration jobs...  "
echo "--------------------------------------------------------------------------------"

if [ ! -d "${DEVELOPMENT_DIR}/tools/vendor" ]; then
    echo "Vendoring 'tools'"
    pushd "${DEVELOPMENT_DIR}/tools"
    dep ensure -v -vendor-only
    popd
fi

echo "running ${DEVELOPMENT_DIR}/tools/cmd/${TOOL_DIR}/main.go"
go run "${DEVELOPMENT_DIR}/tools/cmd/${TOOL_DIR}/main.go" "$@"
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "SUCCESS"
fi
