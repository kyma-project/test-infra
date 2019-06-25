#!/usr/bin/env bash

#Description: Deletes docker image from registry
#
#Expected vars:
# - KYMA_INSTALLER_IMAGE: Image name with tag
# - CLOUDSDK_CORE_PROJECT: GCloud Project name, used for KYMA_INSTALLER_IMAGE validation
#
#Permissions: In order to run this script you need to use a service account with "Storage Admin" role

set -o errexit

for var in KYMA_INSTALLER_IMAGE CLOUDSDK_CORE_PROJECT; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

gcloud container images delete "${KYMA_INSTALLER_IMAGE}"

