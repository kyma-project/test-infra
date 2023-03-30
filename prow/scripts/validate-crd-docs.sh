#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export KYMA_SOURCES_DIR="/home/prow/go/src/github.com/kyma-project/kyma"

# shellcheck source=prow/scripts/lib/docker.sh
source "${SCRIPT_DIR}/lib/docker.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"

OUTPUT=0

function validate_crd_md() {
    sh ${KYMA_SOURCES_DIR}/hack/verify-md.sh
}

function main() {
    log::info "Validate CRD documentation tables"
    validate_crd_md

    exit ${OUTPUT}
}
main