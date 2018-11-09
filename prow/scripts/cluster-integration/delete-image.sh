#!/usr/bin/env bash

#Description: Deletes docker image from registry
#
#Expected vars:
# - CLOUDSDK_CORE_PROJECT: name of a GCP project
# - KYMA_INSTALLER_IMAGE: Image name with tag
#
#Permissions: In order to run this script you need to use a service account with "Storage Admin" role

set -o errexit

for var in KYMA_INSTALLER_IMAGE TEST_INFRA_SOURCES_DIR; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/validate-image-name.sh

gcloud container images delete "${KYMA_INSTALLER_IMAGE}"

