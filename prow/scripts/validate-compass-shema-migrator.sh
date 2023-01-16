#!/usr/bin/env bash

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"
KYMA_PROJECT_DIR=${KYMA_PROJECT_DIR:-"/home/prow/go/src/github.com/kyma-project"}

# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${SCRIPT_DIR}/lib/utils.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${SCRIPT_DIR}/lib/docker.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$SCRIPT_DIR/lib/gcp.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    log::info "Execute Job Guard"
    export JOB_NAME_PATTERN="(pull-.*)"
    export JOBGUARD_TIMEOUT="60m"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

log::info "Authenticate"
gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"

log::info "Start Docker"
docker::start

chmod -R 0777 /home/prow/go/src/github.com/kyma-incubator/compass/.git

log::info "Triggering the validation"

cd /home/prow/go/src/github.com/kyma-incubator/compass/components/schema-migrator/ && make validate

log::info "Validation finished"
