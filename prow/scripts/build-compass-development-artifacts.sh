#!/usr/bin/env bash

# This script generates Compass development artifacts:
# - compass-installer image
# - compass-installer.yaml
# - is-installed.sh
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

readonly COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET="${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/compass"
readonly CURRENT_TIMESTAMP=$(date +%s)

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

function export_variables() {
    COMMIT_ID=$(echo "${PULL_BASE_SHA}" | cut -c1-8)
    COMPASS_INSTALLER_PUSH_DIR=""
   if [[ -n "${PULL_NUMBER}" ]]; then
        DOCKER_TAG="PR-${PULL_NUMBER}-${COMMIT_ID}"
        BUCKET_DIR="PR-${PULL_NUMBER}"
    else
        DOCKER_TAG="master-${COMMIT_ID}-${CURRENT_TIMESTAMP}"
        BUCKET_DIR="master-${COMMIT_ID}"
    fi

   readonly DOCKER_TAG
   readonly COMPASS_INSTALLER_PUSH_DIR
   readonly BUCKET_DIR
   readonly COMPASS_INSTALLER_VERSION

   export DOCKER_TAG
   export COMPASS_INSTALLER_PUSH_DIR
   export BUCKET_DIR
}

init
export_variables

# installer ci-pr, ci-master, kyma-installer ci-pr, ci-master
#   DOCKER_TAG - calculated in export_variables
#   DOCKER_PUSH_DIRECTORY, preset-build-master, preset-build-pr
#   DOCKER_PUSH_REPOSITORY - preset-docker-push-repository
export COMPASS_PATH="/home/prow/go/src/github.com/kyma-incubator/compass"
buildTarget="release"

shout "Build compass-installer with target ${buildTarget}"
make -C "${COMPASS_PATH}/tools/compass-installer" ${buildTarget}

shout "Create development artifacts"
# INPUTS:
# - COMPASS_INSTALLER_PUSH_DIR
# - COMPASS_INSTALLER_VERSION
#  These variables are used to calculate installer version: eu.gcr.io/kyma-project/${COMPASS_INSTALLER_PUSH_DIR}compass-installer:${COMPASS_INSTALLER_VERSION}
# - ARTIFACTS_DIR - path to directory where artifacts will be stored
env COMPASS_INSTALLER_VERSION="${DOCKER_TAG}" ARTIFACTS_DIR="${ARTIFACTS}" "${COMPASS_PATH}/installation/scripts/generate-compass-installer-artifacts.sh"

shout "Content of the local artifacts directory"
ls -la "${ARTIFACTS}"

shout "Switch to a different service account to push to GCS bucket"

export GOOGLE_APPLICATION_CREDENTIALS=/etc/credentials/sa-kyma-artifacts/service-account.json
authenticate

shout "Copy artifacts to ${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}"
gsutil cp  "${ARTIFACTS}/compass-installer.yaml" "${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/compass-installer.yaml"
gsutil cp  "${COMPASS_PATH}/installation/scripts/is-installed.sh" "${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/is-installed.sh"

if [[ "${BUILD_TYPE}" == "master" ]]; then
  shout "Copy artifacts to ${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/master"
  gsutil cp "${ARTIFACTS}/compass-installer.yaml" "${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/master/compass-installer.yaml"
  gsutil cp  "${COMPASS_PATH}/installation/scripts/is-installed.sh" "${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/master/is-installed.sh"
fi
