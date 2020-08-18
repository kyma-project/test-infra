#!/usr/bin/env bash

set -o errexit

VARIABLES=(
    TEST_INFRA_SOURCES_DIR
    LOG_COLLECTOR_SLACK_TOKEN
    PROW_JOB_NAME
)

discoverUnsetVar=false

for var in "${VARIABLES[@]}"; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

function installTestLogColletor() {
    TLC_DIR="${TEST_INFRA_SOURCES_DIR}/development/test-log-collector"
    
    # technically `go run` would suffice here, maybe it would be quicker

    helm install test-log-collector --set slackToken="${LOG_COLLECTOR_SLACK_TOKEN}" \
    --set prowJobName="${PROW_JOB_NAME}" \
    "${TLC_DIR}/chart/test-log-collector" \
    --namespace=kyma-system \
    --wait \
    --timeout=600s
}

installTestLogColletor
