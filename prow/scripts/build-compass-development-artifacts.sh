#!/usr/bin/env bash

# This script generates Compass development artifacts:
# - compass-installer image
# - compass-installer.yaml
# - is-installed.sh
# Yaml files, as well as is-installed.sh script are stored on GCS.

set -e

readonly SCRIPT_DIR
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=prow/scripts/lib/log.sh
. "${SCRIPT_DIR}/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${SCRIPT_DIR}/lib/utils.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${SCRIPT_DIR}/lib/docker.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$SCRIPT_DIR/lib/gcp.sh"

requiredVars=(
    DOCKER_PUSH_REPOSITORY
    KYMA_DEVELOPMENT_ARTIFACTS_BUCKET
)

utils::check_required_vars "${requiredVars[@]}"

readonly COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET
COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET="${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/compass"
readonly CURRENT_TIMESTAMP
CURRENT_TIMESTAMP=$(date +%s)

function export_variables() {
    COMMIT_ID=$(echo "${PULL_BASE_SHA}" | cut -c1-8)
    COMPASS_INSTALLER_PUSH_DIR=""
   if [[ -n "${PULL_NUMBER}" ]]; then
        DOCKER_TAG="PR-${PULL_NUMBER}-${COMMIT_ID}"
        BUCKET_DIR="PR-${PULL_NUMBER}"
   elif [[ "${PULL_BASE_REF}" == "master" || "${PULL_BASE_REF}" == "main" ]]; then
        DOCKER_TAG="main-${COMMIT_ID}-${CURRENT_TIMESTAMP}"
        BUCKET_DIR="main-${COMMIT_ID}"
   else
        DOCKER_TAG="${PULL_BASE_REF}"
        SKIP_ARTIFACT_UPLOAD=true
   fi

   readonly DOCKER_TAG
   readonly COMPASS_INSTALLER_PUSH_DIR
   readonly BUCKET_DIR
   readonly COMPASS_INSTALLER_VERSION

   export DOCKER_TAG
   export COMPASS_INSTALLER_PUSH_DIR
   export BUCKET_DIR
}

gcp::authenticate \
     -c "${GOOGLE_APPLICATION_CREDENTIALS}"
docker::start
export_variables

# installer ci-pr, ci-main, kyma-installer ci-pr, ci-main
#   DOCKER_TAG - calculated in export_variables
#   DOCKER_PUSH_DIRECTORY, preset-build-main, preset-build-pr
#   DOCKER_PUSH_REPOSITORY - preset-docker-push-repository
export COMPASS_PATH="/home/prow/go/src/github.com/kyma-incubator/compass"
buildTarget="release"

log::info "Build compass-installer with target ${buildTarget}"
make -C "${COMPASS_PATH}/tools/compass-installer" ${buildTarget}

if [[ -n "${SKIP_ARTIFACT_UPLOAD}" ]]; then
    log::info "Skipping development artifacts upload"
    exit
fi


log::info "Create development artifacts"
# INPUTS:
# - COMPASS_INSTALLER_PUSH_DIR
# - COMPASS_INSTALLER_VERSION
#  These variables are used to calculate installer version: eu.gcr.io/kyma-project/${COMPASS_INSTALLER_PUSH_DIR}compass-installer:${COMPASS_INSTALLER_VERSION}
# - ARTIFACTS_DIR - path to directory where artifacts will be stored
env COMPASS_INSTALLER_VERSION="${DOCKER_TAG}" ARTIFACTS_DIR="${ARTIFACTS}" "${COMPASS_PATH}/installation/scripts/generate-compass-installer-artifacts.sh"

log::info "Content of the local artifacts directory"
ls -la "${ARTIFACTS}"

log::info "Switch to a different service account to push to GCS bucket"

export GOOGLE_APPLICATION_CREDENTIALS=/etc/credentials/sa-kyma-artifacts/service-account.json
gcp::authenticate \
     -c "${GOOGLE_APPLICATION_CREDENTIALS}"

log::info "Copy artifacts to ${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}"
gsutil cp  "${ARTIFACTS}/compass-installer.yaml" "${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/compass-installer.yaml"
gsutil cp  "${ARTIFACTS}/kyma-installer.yaml" "${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kyma-installer.yaml"
gsutil cp  "${COMPASS_PATH}/installation/scripts/is-installed.sh" "${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/is-installed.sh"
gsutil cp  "${ARTIFACTS}/is-kyma-installed.sh" "${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/is-kyma-installed.sh"

if [[ "${BUILD_TYPE}" == "master" || "${BUILD_TYPE}" == "main" ]]; then
  log::info "Copy artifacts to ${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/main"
  gsutil cp "${ARTIFACTS}/compass-installer.yaml" "${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/main/compass-installer.yaml"
  gsutil cp  "${ARTIFACTS}/is-kyma-installed.sh" "${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/main/is-kyma-installed.sh"
  gsutil cp  "${COMPASS_PATH}/installation/scripts/is-installed.sh" "${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/main/is-installed.sh"
  gsutil cp  "${ARTIFACTS}/kyma-installer.yaml" "${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/main/kyma-installer.yaml"
fi
