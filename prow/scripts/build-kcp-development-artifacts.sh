#!/usr/bin/env bash

# This script generates KCP development artifacts:
# - kcp-installer image
# - kcp-installer.yaml
# - is-installed.sh
# - KCP dependencies artifacts
# Yaml files, as well as is-installed.sh script are stored on GCS.

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"
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

readonly KCP_DEVELOPMENT_ARTIFACTS_BUCKET="${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/kcp"
readonly CURRENT_TIMESTAMP=$(date +%s)


function export_variables() {
    COMMIT_ID=$(echo "${PULL_BASE_SHA}" | cut -c1-8)
    KCP_INSTALLER_PUSH_DIR=""
   if [[ -n "${PULL_NUMBER}" ]]; then
        DOCKER_TAG="PR-${PULL_NUMBER}-${COMMIT_ID}"
        BUCKET_DIR="PR-${PULL_NUMBER}"
    else
        DOCKER_TAG="master-${COMMIT_ID}-${CURRENT_TIMESTAMP}"
        BUCKET_DIR="master-${COMMIT_ID}"
    fi

   readonly DOCKER_TAG
   readonly KCP_INSTALLER_PUSH_DIR
   readonly BUCKET_DIR
   readonly KCP_INSTALLER_VERSION

   export DOCKER_TAG
   export KCP_INSTALLER_PUSH_DIR
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
export KCP_PATH="/home/prow/go/src/github.com/kyma-project/control-plane"
buildTarget="release"

log::info "Build kcp-installer with target ${buildTarget}"
make -C "${KCP_PATH}/tools/kcp-installer" ${buildTarget}

log::info "Switch to a different service account to push to GCS bucket"
export GOOGLE_APPLICATION_CREDENTIALS=/etc/credentials/sa-kyma-artifacts/service-account.json
gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"

log::info "Create development artifacts"
# INPUTS:
# - KCP_INSTALLER_PUSH_DIR
# - KCP_INSTALLER_VERSION
#  These variables are used to calculate installer version: eu.gcr.io/kyma-project/${KCP_INSTALLER_PUSH_DIR}kcp-installer:${KCP_INSTALLER_VERSION}
# - ARTIFACTS_DIR - path to directory where artifacts will be stored
env KCP_INSTALLER_VERSION="${DOCKER_TAG}" ARTIFACTS_DIR="${ARTIFACTS}" "${KCP_PATH}/installation/scripts/generate-installer-artifacts.sh"

log::info "Content of the local artifacts directory"
ls -la "${ARTIFACTS}"

log::info "Copy artifacts to ${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}"
gsutil cp  "${ARTIFACTS}/kcp-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kcp-installer.yaml"
gsutil cp  "${KCP_PATH}/installation/scripts/is-installed.sh" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/is-installed.sh"

gsutil cp  "${ARTIFACTS}/kyma-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kyma-installer.yaml"
gsutil cp  "${ARTIFACTS}/kyma-kcp-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kyma-kcp-installer.yaml"
gsutil cp  "${ARTIFACTS}/is-kyma-installed.sh" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/is-kyma-installed.sh"

gsutil cp  "${ARTIFACTS}/compass-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/compass-installer.yaml"
gsutil cp  "${ARTIFACTS}/is-compass-installed.sh" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/is-compass-installed.sh"

if [[ "${BUILD_TYPE}" == "master" ]]; then
  log::info "Copy artifacts to ${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master"
  gsutil cp "${ARTIFACTS}/kcp-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kcp-installer.yaml"
  gsutil cp  "${KCP_PATH}/installation/scripts/is-installed.sh" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/is-installed.sh"

  gsutil cp  "${ARTIFACTS}/kyma-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kyma-installer.yaml"
  gsutil cp  "${ARTIFACTS}/kyma-kcp-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kyma-kcp-installer.yaml"
  gsutil cp  "${ARTIFACTS}/is-kyma-installed.sh" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/is-kyma-installed.sh"

  gsutil cp "${ARTIFACTS}/compass-installer.yaml" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/compass-installer.yaml"
  gsutil cp  "${ARTIFACTS}/is-compass-installed.sh" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/is-compass-installed.sh"
fi
