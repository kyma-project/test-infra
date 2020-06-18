#!/usr/bin/env bash

#Description: Deletes docker image from registry
#
#Expected vars:
# - KYMA_INSTALLER_IMAGE: Image name with tag
# - CLOUDSDK_CORE_PROJECT: GCloud Project name, used for KYMA_INSTALLER_IMAGE validation
#
#Permissions: In order to run this script you need to use a service account with "Storage Admin" role

set -o errexit

for var in KYMA_INSTALLER_IMAGE CLOUDSDK_CORE_PROJECT GOOGLE_APPLICATION_CREDENTIALS; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

function cleanup() {
  activateDefaultSa
}

shout "Authenticate as service account with write access to GCR"
date
trap cleanup EXIT
authenticateSaGcr

gcloud container images delete "${KYMA_INSTALLER_IMAGE}"

