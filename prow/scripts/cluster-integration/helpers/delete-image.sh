#!/usr/bin/env bash

#Description: Deletes docker image from registry
#
#Expected vars:
# - KYMA_INSTALLER_IMAGE: Image name with tag
# - CLOUDSDK_CORE_PROJECT: GCloud Project name, used for KYMA_INSTALLER_IMAGE validation
# - GOOGLE_APPLICATION_CREDENTIALS
# - TEST_INFRA_SOURCES_DIR
#
#Permissions: In order to run this script you need to use a service account with "Storage Admin" role

set -o errexit

for var in KYMA_INSTALLER_IMAGE CLOUDSDK_CORE_PROJECT GOOGLE_APPLICATION_CREDENTIALS TEST_INFRA_SOURCES_DIR; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

function cleanup() {
  activateDefaultSa
}

shout "Authenticate as service account with write access to GCR"
date
trap cleanup EXIT
authenticateSaGcr

gcloud container images delete "${KYMA_INSTALLER_IMAGE}"

