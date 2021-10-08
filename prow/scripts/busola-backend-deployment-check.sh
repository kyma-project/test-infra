#!/bin/bash

set -o errexit
set -o pipefail

readonly SCRIPT_DIR
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR
TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"
readonly TMP_DIR
TMP_DIR=$(mktemp -d)


# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    log::info "Check if backend deployment was updated in PR."
    if grep -w "${PULL_NUMBER}" "/home/prow/go/src/github.com/kyma-project/busola/resources/backend/deployment.yaml"; then
        log::success "Pull request number found."
    else
        log::error "Pull request number not found. Please update deployment image in your PR."
        exit 1
    fi
fi