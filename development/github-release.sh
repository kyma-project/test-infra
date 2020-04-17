#!/usr/bin/env bash

set -e
set -o pipefail

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then
   echo "GOOGLE_APPLICATION_CREDENTIALS environment variable is missing!"
   exit 1
fi

NEXT_RELEASE=$(cat "${DEVELOPMENT_DIR}/../prow/RELEASE_VERSION")
echo "Checking if ${NEXT_RELEASE} was already published on github..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" https://api.github.com/repos/kyma-project/kyma/releases/tags/"${NEXT_RELEASE}")
if [[ $RESPONSE != 404* ]]; then
    echo "The ${NEXT_RELEASE} is already published on github. Stopping."
    exit 1
fi

echo "--------------------------------------------------------------------------------"
echo "Creating the Github release for Kyma...  "
echo "--------------------------------------------------------------------------------"

/prow-tools/githubrelease "$@"
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "SUCCESS"
fi
