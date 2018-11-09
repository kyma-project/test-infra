#!/usr/bin/env bash

set -o errexit

#Basic sanity check for image name, just to filter out nonsensical values
PROPER_HOST_AND_PROJECT="^eu[.]gcr[.]io[/]${CLOUDSDK_CORE_PROJECT}"
AT_LEAST_THREE_DIRS="([/].{2,}){3,}"
TAG="[:].{2,}"

REGEX="${PROPER_HOST_AND_PROJECT}${AT_LEAST_THREE_DIRS}${TAG}"

if [[ ! "${KYMA_INSTALLER_IMAGE}" =~ ${REGEX} ]]; then
    echo "Unexpected Kyma-Installer image name, please double check!"
    echo "Expected (regex): ${REGEX}"
    echo "Actual: ${KYMA_INSTALLER_IMAGE}"
    exit 1
fi

