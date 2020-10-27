#!/usr/bin/env bash

# This script generates KCP development artifacts:
# - kcp-installer image
# - kcp-installer.yaml
# - is-installed.sh
# - KCP dependencies artifacts
# Yaml files, as well as is-installed.sh script are stored on GCS.

set -e

discoverUnsetVar=false

for var in DOCKER_PUSH_REPOSITORY KYMA_DEVELOPMENT_ARTIFACTS_BUCKET; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

readonly KCP_DEVELOPMENT_ARTIFACTS_BUCKET="${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/kcp"
readonly CURRENT_TIMESTAMP=$(date +%s)

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

function export_variables() {
    COMMIT_ID=$(echo "${PULL_BASE_SHA}" | cut -c1-8)
    KCP_INSTALLER_PUSH_DIR=""
   if [[ -n "${PULL_NUMBER}" ]]; then
        DOCKER_TAG="PR-${PULL_NUMBER}-${COMMIT_ID}"
        BUCKET_DIR="PR-${PULL_NUMBER}"
        CLI_VERSION="PR-${PULL_NUMBER}-${COMMIT_ID}"
    else
        DOCKER_TAG="master-${COMMIT_ID}-${CURRENT_TIMESTAMP}"
        BUCKET_DIR="master-${COMMIT_ID}"
        CLI_VERSION="master-${COMMIT_ID}"
    fi

   readonly DOCKER_TAG
   readonly KCP_INSTALLER_PUSH_DIR
   readonly BUCKET_DIR
   readonly KCP_INSTALLER_VERSION
   readonly CLI_VERSION

   export DOCKER_TAG
   export KCP_INSTALLER_PUSH_DIR
   export BUCKET_DIR
   export CLI_VERSION
}

init
export_variables

# installer ci-pr, ci-master, kyma-installer ci-pr, ci-master
#   DOCKER_TAG - calculated in export_variables
#   DOCKER_PUSH_DIRECTORY, preset-build-master, preset-build-pr
#   DOCKER_PUSH_REPOSITORY - preset-docker-push-repository
export KCP_PATH="/home/prow/go/src/github.com/kyma-project/control-plane"
buildTarget="release"

shout "Build kcp-installer with target ${buildTarget}"
make -C "${KCP_PATH}/tools/kcp-installer" ${buildTarget}

shout "Build KCP CLI with target ${buildTarget}"
make -C "${KCP_PATH}/components/kyma-environment-broker" -f Makefile.cli ${buildTarget}

shout "Switch to a different service account to push to GCS bucket"
export GOOGLE_APPLICATION_CREDENTIALS=/etc/credentials/sa-kyma-artifacts/service-account.json
authenticate

shout "Create development artifacts"
# INPUTS:
# - KCP_INSTALLER_PUSH_DIR
# - KCP_INSTALLER_VERSION
#  These variables are used to calculate installer version: eu.gcr.io/kyma-project/${KCP_INSTALLER_PUSH_DIR}kcp-installer:${KCP_INSTALLER_VERSION}
# - ARTIFACTS_DIR - path to directory where artifacts will be stored
env KCP_INSTALLER_VERSION="${DOCKER_TAG}" ARTIFACTS_DIR="${ARTIFACTS}" "${KCP_PATH}/installation/scripts/generate-installer-artifacts.sh"

shout "Content of the local artifacts directory"
ls -la "${ARTIFACTS}"

shout "Copy artifacts to ${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}"
gsutil cp  "${ARTIFACTS}/kcp-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kcp-installer.yaml"
gsutil cp  "${KCP_PATH}/installation/scripts/is-installed.sh" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/is-installed.sh"

gsutil cp  "${ARTIFACTS}/kyma-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kyma-installer.yaml"
gsutil cp  "${ARTIFACTS}/is-kyma-installed.sh" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/is-kyma-installed.sh"

gsutil cp  "${ARTIFACTS}/compass-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/compass-installer.yaml"
gsutil cp  "${ARTIFACTS}/is-compass-installed.sh" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/is-compass-installed.sh"

gsutil cp "${ARTIFACTS}/kcp.exe" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kcp.exe"
gsutil cp "${ARTIFACTS}/kcp-linux" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kcp-linux"
gsutil cp "${ARTIFACTS}/kcp-darwin" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kcp-darwin"

if [[ "${BUILD_TYPE}" == "master" ]]; then
  shout "Copy artifacts to ${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master"
  gsutil cp "${ARTIFACTS}/kcp-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kcp-installer.yaml"
  gsutil cp  "${KCP_PATH}/installation/scripts/is-installed.sh" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/is-installed.sh"

  gsutil cp  "${ARTIFACTS}/kyma-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kyma-installer.yaml"
  gsutil cp  "${ARTIFACTS}/is-kyma-installed.sh" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/is-kyma-installed.sh"

  gsutil cp "${ARTIFACTS}/compass-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/compass-installer.yaml"
  gsutil cp  "${ARTIFACTS}/is-compass-installed.sh" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/is-compass-installed.sh"

  gsutil cp "${ARTIFACTS}/kcp.exe" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kcp.exe"
  gsutil cp "${ARTIFACTS}/kcp-linux" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kcp-linux"
  gsutil cp "${ARTIFACTS}/kcp-darwin" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kcp-darwin"
fi
