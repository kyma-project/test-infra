#!/usr/bin/env bash

#Description: Builds Kyma-Installer image from Kyma sources and pushes it to the repository
#
#Expected vars:
# - KYMA_SOURCES_DIR: directory with Kyma sources to build Kyma-Installer image
# - KYMA_INSTALLER_IMAGE: Full image name (with tag)
#
#Permissions: In order to run this script you need to use a service account with "Storage Admin" role

set -o errexit

for var in KYMA_SOURCES_DIR KYMA_INSTALLER_IMAGE; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

echo "--------------------------------------------------------------------------------"
echo "Building Kyma-Installer image: ${KYMA_INSTALLER_IMAGE}"
echo "--------------------------------------------------------------------------------"
echo
docker build "${KYMA_SOURCES_DIR}" -f "${KYMA_SOURCES_DIR}"/tools/kyma-installer/kyma.Dockerfile -t "${KYMA_INSTALLER_IMAGE}"

echo "--------------------------------------------------------------------------------"
echo "pushing Kyma-Installer image"
echo "--------------------------------------------------------------------------------"
echo
docker push "${KYMA_INSTALLER_IMAGE}"
echo "--------------------------------------------------------------------------------"
echo "Kyma-Installer image pushed: ${KYMA_INSTALLER_IMAGE}"
echo "--------------------------------------------------------------------------------"
