#!/usr/bin/env bash

# This script is executed during release process and generates kyma artifacts. Artifacts are stored in $(ARTIFACTS) location
# that is automatically uploaded by Prow to GCS bucket in the following location:
# <plank gcs bucket>/pr-logs/pull/<org_repository>/<pull_request_number>/kyma-artifacts/<build_id>/artifacts
# Information about latest build id is stored in:
# <plank gcs bucket>/pr-logs/pull/<org_repository>/<pull_request_number>/kyma-artifacts/latest-build.txt

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_RESOURCES_DIR="/home/prow/go/src/github.com/kyma-project/kyma/installation/resources"

# shellcheck source=prow/scripts/lib/docker.sh
source "${SCRIPT_DIR}/lib/docker.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$SCRIPT_DIR/lib/gcp.sh"

# copy_artifacts copies artifacts to the destined bucket path.
# it accepts one argument BUCKET_PATH which should be formatted as:
# gs://bucket-name/bucket-folder
function copy_artifacts {
  BUCKET_PATH=$1
  log::info "Copying artifacts to $BUCKET_PATH..."

  cp "${KYMA_RESOURCES_DIR}/components.yaml" "${ARTIFACTS_DIR}/kyma-components.yaml"
  gsutil cp "${KYMA_RESOURCES_DIR}/components.yaml" "$BUCKET_PATH/kyma-components.yaml"
}

docker::start

if [ -n "${PULL_NUMBER}" ]; then
  DOCKER_TAG="PR-${PULL_NUMBER}"
elif [[ "${PULL_BASE_REF}" =~ ^release-.* ]]; then
  DOCKER_TAG=$(cat "VERSION")
  if [[ "${NEXT_RELEASE}" == "main" ]]; then
      echo "You can't publish release artifacts with the version set to 'main'"
      exit 1
  fi

  if [[ "${REPO_OWNER}" == "kyma-project" && "${REPO_NAME}" == "kyma" ]]; then
    NEXT_RELEASE="$DOCKER_TAG"
    echo "Checking if ${NEXT_RELEASE} was already published on github..."
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" https://api.github.com/repos/kyma-project/kyma/releases/tags/"${NEXT_RELEASE}")
    if [[ $RESPONSE != 404* ]]; then
        echo "The ${NEXT_RELEASE} is already published on github. Stopping."
        exit 1
    fi
  fi
else
  DOCKER_TAG="${PULL_BASE_SHA::8}"
fi
export DOCKER_TAG
echo "DOCKER_TAG: ${DOCKER_TAG}"

log::info "Content of the local artifacts directory"
ls -la "${ARTIFACTS}"
gcp::authenticate \
  -c "$SA_KYMA_ARTIFACTS_GOOGLE_APPLICATION_CREDENTIALS"

if [ -n "$PULL_NUMBER" ]; then
  copy_artifacts "${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/${DOCKER_TAG}"
elif [[ "$PULL_BASE_REF" =~ ^release-.* ]]; then
  copy_artifacts "${KYMA_ARTIFACTS_BUCKET}/${DOCKER_TAG}"
else
  copy_artifacts "${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/main-${DOCKER_TAG}"
  copy_artifacts "${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/main"
fi
