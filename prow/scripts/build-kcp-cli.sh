#!/usr/bin/env bash

# This script builds and publishes KCP CLI development artifacts

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${SCRIPT_DIR}/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$SCRIPT_DIR/lib/gcp.sh"

requiredVars=(
    KYMA_DEVELOPMENT_ARTIFACTS_BUCKET
)

utils::check_required_vars "${requiredVars[@]}"

readonly KCP_DEVELOPMENT_ARTIFACTS_BUCKET="${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/kcp"
readonly CURRENT_TIMESTAMP=$(date +%s)



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

gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"
export_variables

export KCP_PATH="/home/prow/go/src/github.com/kyma-project/control-plane"
buildTarget="release"
log::info "Build KCP CLI with target ${buildTarget}"
make -C "${KCP_PATH}/tools/cli" ${buildTarget}

log::info "Switch to a different service account to push to GCS bucket"
export GOOGLE_APPLICATION_CREDENTIALS=/etc/credentials/sa-kyma-artifacts/service-account.json
gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"

log::info "Content of the local artifacts directory"
ls -la "${ARTIFACTS}"

log::info "Copy artifacts to ${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}"

gsutil cp "${ARTIFACTS}/kcp.exe" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kcp.exe"
gsutil cp "${ARTIFACTS}/kcp-linux" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kcp-linux"
gsutil cp "${ARTIFACTS}/kcp-darwin" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${BUCKET_DIR}/kcp-darwin"

if [[ "${BUILD_TYPE}" == "master" ]]; then
  log::info "Copy artifacts to ${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master"

  gsutil cp "${ARTIFACTS}/kcp.exe" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kcp.exe"
  gsutil cp "${ARTIFACTS}/kcp-linux" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kcp-linux"
  gsutil cp "${ARTIFACTS}/kcp-darwin" "${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/master/kcp-darwin"
fi
