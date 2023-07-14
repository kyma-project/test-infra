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

log::info "Content of the local artifacts directory"
ls -la "${ARTIFACTS}"

mkdir -p "${ARTIFACTS}"/sync/{ers,kcp}/"${BUCKET_DIR}"/
mv "${ARTIFACTS}"/ers* "${ARTIFACTS}/sync/ers/${BUCKET_DIR}/"
mv "${ARTIFACTS}"/kcp* "${ARTIFACTS}/sync/kcp/${BUCKET_DIR}/"

if [[ "${BUILD_TYPE}" == "master" ]]; then
  mkdir -p "${ARTIFACTS}"/sync/{ers,kcp}/master/
  cp "${ARTIFACTS}/sync/ers/${BUCKET_DIR}"/ers* "${ARTIFACTS}/sync/ers/master/"
  cp "${ARTIFACTS}/sync/kcp/${BUCKET_DIR}"/kcp* "${ARTIFACTS}/sync/kcp/master/"
fi

log::info "Copy artifacts to ${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/{ers,kcp}/${BUCKET_DIR}"
gsutil -m rsync -r "${ARTIFACTS}/sync/" "${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}"
