#!/usr/bin/env bash

#Description: Performs basic check of image name to filter out nonsensical values.
#
#Expected vars:
# - KYMA_INSTALLER_IMAGE: Full image name (with hostname and tag)
# - CLOUDSDK_CORE_PROJECT: Name of a GCP project in context of which the image is created.

set -o errexit

#Basic sanity checks
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

