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

LABELS=""
if [[ -z "${PULL_NUMBER}" ]]; then
    LABELS=(--labels "branch=$PULL_BASE_REF,job-name=compass-smoke-test")
else
    LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=compass-smoke-test")
fi

log::info "Authenticate"
gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"

log::info "Start Docker"
docker::start

chmod -R 0777 /home/prow/go/src/github.com/kyma-incubator/compass/.git
mkdir -p /home/prow/go/src/github.com/kyma-incubator/compass/components/console/shared/build

log::info "Get ORD commit ID"
ORD_PR_NUMBER=$(yq e .global.images.ord_service.version /home/prow/go/src/github.com/kyma-incubator/compass/chart/compass/values.yaml | cut -d '-' -f 2 | xargs)
log::info "ORD_PR_NUMBER PR is: ${ORD_PR_NUMBER}"

ORD_PR_DATA=$(curl -sS "https://api.github.com/repos/kyma-incubator/ord-service/pulls/${ORD_PR_NUMBER}")
log::info "ORD_PR_DATA is: ${ORD_PR_DATA}"

ORD_PR_STATE=$(jq -r '.state' <<< "${ORD_PR_DATA}")

if [[ "$ORD_PR_STATE" == "open" ]]; then
    ORD_PR_COMMIT_HASH=$(jq -r '.head.sha' <<< "${ORD_PR_DATA}")
else
    ORD_PR_COMMIT_HASH=$(jq -r '.merge_commit_sha' <<< "${ORD_PR_DATA}")
fi

log::info "ORD_PR_COMMIT_HASH is: ${ORD_PR_COMMIT_HASH}"

log::info "Fetch ORD service sources"
cd /home/prow/ && git clone https://github.com/kyma-incubator/ord-service.git && cd ord-service && git checkout "${ORD_PR_COMMIT_HASH}" && cd ..

log::info "Triggering the test"

cd /home/prow/go/src/github.com/kyma-incubator/compass/installation/scripts/prow/
./compass-smoke-test.sh "/home/prow/"

log::info "Test finished"
