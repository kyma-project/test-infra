#!/usr/bin/env bash

set -e
set -o pipefail

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} gcpProjectName"
    exit 1
}

if [ -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then
   echo "GOOGLE_APPLICATION_CREDENTIALS environment variable is missing!"
   exit 1
fi

readonly PROJECT="$1"

if [ -z "${PROJECT}" ]; then
    usage
fi

echo "Removing GCP LoadBalancer resources allocated by failed/terminated integration jobs..."

if [ ! -d "${DEVELOPMENT_DIR}/tools/vendor" ]; then
    echo "Vendoring 'tools'"
    pushd "${DEVELOPMENT_DIR}/tools"
    dep ensure -v -vendor-only
    popd
fi

go run "${DEVELOPMENT_DIR}"/tools/cmd/orphanremover/main.go  --project="${PROJECT}" --dryRun=false
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "SUCCESS"
fi
