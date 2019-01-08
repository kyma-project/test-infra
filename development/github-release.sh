#!/usr/bin/env bash

set -e
set -o pipefail

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then
   echo "GOOGLE_APPLICATION_CREDENTIALS environment variable is missing!"
   exit 1
fi

echo "--------------------------------------------------------------------------------"
echo "Creating the Github release for Kyma...  "
echo "--------------------------------------------------------------------------------"

if [ ! -d "${DEVELOPMENT_DIR}/tools/vendor" ]; then
    echo "Vendoring 'tools'"
    pushd "${DEVELOPMENT_DIR}/tools"
    dep ensure -v -vendor-only
    popd
fi

go run "${DEVELOPMENT_DIR}/tools/cmd/githubrelease/main.go" "$@"
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "SUCCESS"
fi
