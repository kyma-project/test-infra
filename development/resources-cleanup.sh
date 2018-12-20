#!/usr/bin/env bash

################################################################################
# DO NOT INVOKE DIRECTLY, USE FROM OTHER SCRIPTS
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


go run "${DEVELOPMENT_DIR}/tools/cmd/${TOOL_DIR}/main.go" "$@"
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "SUCCESS"
fi
