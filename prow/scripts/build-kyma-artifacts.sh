#!/usr/bin/env bash

# This script is executed during release process and generates kyma artifacts. Artifacts are stored in $(ARTIFACTS) location
# that is automatically uploaded by Prow to GCS bucket in the following location:
# <plank gcs bucket>/pr-logs/pull/<org_repository>/<pull_request_number>/kyma-artifacts/<build_id>/artifacts
# Information about latest build id is stored in:
# <plank gcs bucket>/pr-logs/pull/<org_repository>/<pull_request_number>/kyma-artifacts/latest-build.txt

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

function export_variables() {
   DOCKER_TAG=$(cat "${SCRIPT_DIR}/../RELEASE_VERSION")
   echo "Reading docker tag from RELEASE_VERSION file, got: ${DOCKER_TAG}"

   readonly DOCKER_TAG
   export DOCKER_TAG
}

init
export_variables

make -C /home/prow/go/src/github.com/kyma-project/kyma/tools/kyma-installer ci-create-release-artifacts

gsutil cp "${ARTIFACTS}/kyma-config-cluster.yaml" "${KYMA_ARTIFACTS_BUCKET}/${DOCKER_TAG}/kyma-config-cluster.yaml"
gsutil cp "${ARTIFACTS}/kyma-installer-cluster.yaml" "${KYMA_ARTIFACTS_BUCKET}/${DOCKER_TAG}/kyma-installer-cluster.yaml"

gsutil cp "${ARTIFACTS}/kyma-config-local.yaml" "${KYMA_ARTIFACTS_BUCKET}/${DOCKER_TAG}/kyma-config-local.yaml"
gsutil cp "${ARTIFACTS}/kyma-installer-local.yaml" "${KYMA_ARTIFACTS_BUCKET}/${DOCKER_TAG}/kyma-installer-local.yaml"

"${SCRIPT_DIR}"/changelog-generator.sh