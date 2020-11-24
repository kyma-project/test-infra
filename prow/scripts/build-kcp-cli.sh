#!/usr/bin/env bash

# This script builds and publishes KCP CLI development artifacts

set -e

discoverUnsetVar=false

function check_missing_var() {
    local var=$1
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
}

check_missing_var KYMA_DEVELOPMENT_ARTIFACTS_BUCKET
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
   if [[ -n "${PULL_NUMBER}" ]]; then
        BUCKET_DIR="PR-${PULL_NUMBER}"
        CLI_VERSION="PR-${PULL_NUMBER}-${COMMIT_ID}"
    else
        BUCKET_DIR="master-${COMMIT_ID}"
        CLI_VERSION="master-${COMMIT_ID}"
    fi

   readonly BUCKET_DIR
   readonly CLI_VERSION

   export BUCKET_DIR
   export CLI_VERSION
}

init
export_variables

export KCP_PATH="/home/prow/go/src/github.com/kyma-project/control-plane"
buildTarget="release"

shout "Build KCP CLI with target ${buildTarget}"
make -C "${KCP_PATH}/tools/cli" ${buildTarget}

shout "Switch to a different service account to push to GCS bucket"
export GOOGLE_APPLICATION_CREDENTIALS=/etc/credentials/sa-kyma-artifacts/service-account.json
authenticate

shout "Content of the local artifacts directory"
ls -la "${ARTIFACTS}"

shout "Copy artifacts to ${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}"

gsutil cp "${ARTIFACTS}/kcp.exe" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kcp.exe"
gsutil cp "${ARTIFACTS}/kcp-linux" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kcp-linux"
gsutil cp "${ARTIFACTS}/kcp-darwin" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kcp-darwin"

if [[ "${BUILD_TYPE}" == "master" ]]; then
  shout "Copy artifacts to ${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master"

  gsutil cp "${ARTIFACTS}/kcp.exe" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kcp.exe"
  gsutil cp "${ARTIFACTS}/kcp-linux" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kcp-linux"
  gsutil cp "${ARTIFACTS}/kcp-darwin" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kcp-darwin"
fi
