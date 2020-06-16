#!/usr/bin/env bash

#Description: Builds Compass-Installer image from Compass sources and pushes it to the repository
#
#Expected vars:
# - COMPASS_SOURCES_DIR: directory with Compass sources to build Compass-Installer image
# - COMPASS_INSTALLER_IMAGE: Full image name (with tag)
# - CLOUDSDK_CORE_PROJECT: GCloud Project name, used for COMPASS_INSTALLER_IMAGE validation
#
#Permissions: In order to run this script you need to use a service account with "Storage Admin" role

set -o errexit

for var in COMPASS_SOURCES_DIR COMPASS_INSTALLER_IMAGE CLOUDSDK_CORE_PROJECT; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

echo "--------------------------------------------------------------------------------"
echo "Building Compass-Installer image: ${COMPASS_INSTALLER_IMAGE}"
echo "--------------------------------------------------------------------------------"
echo
docker build "${COMPASS_SOURCES_DIR}" -f "${COMPASS_SOURCES_DIR}"/tools/compass-installer/compass.Dockerfile -t "${COMPASS_INSTALLER_IMAGE}"

echo "--------------------------------------------------------------------------------"
echo "pushing Compass-Installer image"
echo "--------------------------------------------------------------------------------"
echo
docker push "${COMPASS_INSTALLER_IMAGE}"
echo "--------------------------------------------------------------------------------"
echo "Compass-Installer image pushed: ${COMPASS_INSTALLER_IMAGE}"
echo "--------------------------------------------------------------------------------"
