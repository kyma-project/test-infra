#!/usr/bin/env bash

#Description: Builds Kyma Control Plane  Installer image from KCP sources and pushes it to the repository
#
#Expected vars:
# - KCP_SOURCES_DIR: directory with KCP sources to build KCP-Installer image
# - KCP_INSTALLER_IMAGE: Full image name (with tag)
# - CLOUDSDK_CORE_PROJECT: GCloud Project name, used for KCP_INSTALLER_IMAGE validation
#
#Permissions: In order to run this script you need to use a service account with "Storage Admin" role

set -o errexit

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

requiredVars=(
    KCP_SOURCES_DIR
    KCP_INSTALLER_IMAGE
    CLOUDSDK_CORE_PROJECT
)

utils::checkRequiredVars ${requiredVars[@]}

echo "--------------------------------------------------------------------------------"
echo "Building KCP-Installer image: ${KCP_INSTALLER_IMAGE}"
echo "--------------------------------------------------------------------------------"
echo
docker build "${KCP_SOURCES_DIR}" -f "${KCP_SOURCES_DIR}"/tools/kcp-installer/kcp.Dockerfile -t "${KCP_INSTALLER_IMAGE}"

echo "--------------------------------------------------------------------------------"
echo "pushing KCP-Installer image"
echo "--------------------------------------------------------------------------------"
echo
docker push "${KCP_INSTALLER_IMAGE}"
echo "--------------------------------------------------------------------------------"
echo "KCP-Installer image pushed: ${KCP_INSTALLER_IMAGE}"
echo "--------------------------------------------------------------------------------"
