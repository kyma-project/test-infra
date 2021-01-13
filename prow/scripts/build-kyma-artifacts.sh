#!/usr/bin/env bash

# This script is executed during release process and generates kyma artifacts. Artifacts are stored in $(ARTIFACTS) location
# that is automatically uploaded by Prow to GCS bucket in the following location:
# <plank gcs bucket>/pr-logs/pull/<org_repository>/<pull_request_number>/kyma-artifacts/<build_id>/artifacts
# Information about latest build id is stored in:
# <plank gcs bucket>/pr-logs/pull/<org_repository>/<pull_request_number>/kyma-artifacts/latest-build.txt

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=prow/scripts/lib/gcloud.sh
source "${SCRIPT_DIR}/lib/gcloud.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${SCRIPT_DIR}/lib/docker.sh"

# copy_artifacts copies artifacts to the destined bucket path.
# it accepts one argument BUCKET_PATH which should be formatted as:
# gs://bucket-name/bucket-folder
function copy_artifacts {
  BUCKET_PATH=$1
  log::info "Copying artifacts to $BUCKET_PATH..."

  gsutil cp  "installation/scripts/is-installed.sh" "$BUCKET_PATH/is-installed.sh"
  gsutil cp "${ARTIFACTS}/kyma-installer-cluster.yaml" "$BUCKET_PATH/kyma-installer-cluster.yaml"
  gsutil cp "${ARTIFACTS}/kyma-installer-cluster-runtime.yaml" "$BUCKET_PATH/kyma-installer-cluster-runtime.yaml"

  gsutil cp "${ARTIFACTS}/kyma-config-local.yaml" "$BUCKET_PATH/kyma-config-local.yaml"
  gsutil cp "${ARTIFACTS}/kyma-installer-local.yaml" "$BUCKET_PATH/kyma-installer-local.yaml"

  gsutil cp "${ARTIFACTS}/kyma-installer.yaml" "$BUCKET_PATH/kyma-installer.yaml"
  gsutil cp "${ARTIFACTS}/kyma-installer-cr-cluster.yaml" "$BUCKET_PATH/kyma-installer-cr-cluster.yaml"
  gsutil cp "${ARTIFACTS}/kyma-installer-cr-local.yaml" "$BUCKET_PATH/kyma-installer-cr-local.yaml"
  gsutil cp "${ARTIFACTS}/kyma-installer-cr-cluster-runtime.yaml" "$BUCKET_PATH/kyma-installer-cr-cluster-runtime.yaml"
}

gcloud::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"
docker::start

# copied from build-generic.sh
# TODO: Find a way to remove this function and use the build-generic.sh script directly. We should use `source`?
function export_variables() {
if [ -n "${PULL_NUMBER}" ]; then
  DOCKER_TAG="PR-${PULL_NUMBER}"
elif [[ "${PULL_BASE_REF}" =~ ^release-.* ]]; then
  DOCKER_TAG=$(cat "${SCRIPT_DIR}/../RELEASE_VERSION")
else
  DOCKER_TAG="${PULL_BASE_SHA::8}"
fi
export DOCKER_TAG
echo "DOCKER_TAG: ${DOCKER_TAG}"
}
export_variables

log::info "Building kyma-installer"
# Building kyma-installer image using build-generic.sh script.
# Handles basically everything related to building process including determining version, exporting DOCKER_TAG etc.
# In the same level as build-kyma-artifacts.sh
"${SCRIPT_DIR}"/build-generic.sh "tools/kyma-installer"

log::info "Create Kyma artifacts"
if [[ -n "${PULL_NUMBER}" ]] && [[ "${PULL_BASE_REF}" =~ ^release-.* ]]; then
  # work only on presubmit release branch.
  log::info "workaround for release presubmits - rollback release kyma-installer to develop for the PRs"
  cp "installation/resources/installer.yaml" "/tmp/installer.tpl.yaml"
  sed -E ";s;image: eu.gcr.io\/kyma-project\/kyma-installer:.+;image: eu.gcr.io\/kyma-project\/develop\/installer:latest;" < "/tmp/installer.tpl.yaml" > "installation/resources/installer.yaml"
fi
env KYMA_INSTALLER_VERSION="${DOCKER_TAG}" ARTIFACTS_DIR="${ARTIFACTS}" "installation/scripts/release-generate-kyma-installer-artifacts.sh"

log::info "Content of the local artifacts directory"
ls -la "${ARTIFACTS}"
gcloud::authenticate "$SA_KYMA_ARTIFACTS_GOOGLE_APPLICATION_CREDENTIALS"

if [ -n "$PULL_NUMBER" ]; then
  copy_artifacts "${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/${DOCKER_TAG}"
elif [[ "$PULL_BASE_REF" =~ ^release-.* ]]; then
  copy_artifacts "${KYMA_ARTIFACTS_BUCKET}/${DOCKER_TAG}"
  # TODO this script needs to be revisited for future improvements...
  "${SCRIPT_DIR}"changelog-generator.sh
else
  copy_artifacts "${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/master-${DOCKER_TAG}"
  copy_artifacts "${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/master"
fi
