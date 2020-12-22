#!/usr/bin/env bash

set -o errexit

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

requiredVars=(
    TEST_INFRA_SOURCES_DIR
    LOG_COLLECTOR_SLACK_TOKEN
    PROW_JOB_NAME
)

utils::check_required_vars "${requiredVars[@]}"

function installTestLogColletor() {
    TLC_DIR="${TEST_INFRA_SOURCES_DIR}/development/test-log-collector"

    helm install test-log-collector --set slackToken="${LOG_COLLECTOR_SLACK_TOKEN}" \
    --set prowJobName="${PROW_JOB_NAME}" \
    "${TLC_DIR}/chart/test-log-collector" \
    --namespace=kyma-system \
    --wait \
    --timeout=600s
}

installTestLogColletor
