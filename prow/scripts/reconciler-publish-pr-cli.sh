#!/bin/bash

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_PROJECT_DIR=${KYMA_PROJECT_DIR:-"/home/prow/go/src/github.com/kyma-project"}
KYMA_CLI_PROJECT_DIR="${KYMA_PROJECT_DIR}/cli"
RECONCILER_PROJECT_DIR="/home/prow/go/src/github.com/kyma-incubator/reconciler"

# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${SCRIPT_DIR}/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$SCRIPT_DIR/lib/gcp.sh"

log::banner "Reconciler Publish PR-CLI"

cd "${KYMA_CLI_PROJECT_DIR}"

log::banner "Replacing reconciler dependency used by the Kyma CLI to local PR source"
go mod edit -replace "github.com/kyma-incubator/reconciler=${RECONCILER_PROJECT_DIR}"

log::info "Contents of kyma-project/cli/go.mod"
cat go.mod

log::banner "Building Kyma CLI with reconciler bump from PR source"
log::info "Resolve dependencies for kyma cli"
make resolve

log::info "Run unit-tests for kyma cli"
make test

log::info "Building Kyma CLI"
make build-linux

log::info "Renaming Kyma CLI to include PR number"
mv "./bin/kyma-linux" "./bin/kyma-linux-pr-${PULL_NUMBER}"
ls "./bin/"

log::banner "Publishing builds to GCP"
log::info "GCP Authentication"
gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"

log::info "Setting bucket info"
export KYMA_CLI_UNSTABLE_BUCKET="${KYMA_CLI_PR_UNSTABLE_BUCKET}"

log::info "Publishing new builds to $KYMA_CLI_UNSTABLE_BUCKET"
make upload-binaries

log::success "Reconciler Publish PR CLI's all done"
