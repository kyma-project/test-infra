#!/usr/bin/env bash

set -o errexit

VARIABLES=(
    TEST_INFRA_SOURCES_DIR
    LOG_COLLECTOR_SLACK_TOKEN
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
    # same as in install-stability-checker.sh
    curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
    
    TLC_DIR="${TEST_INFRA_SOURCES_DIR}/development/test-log-collector"
    
    helm install test-log-collector --set slackToken="${LOG_COLLECTOR_SLACK_TOKEN}" \
    "${TLC_DIR}/chart/test-log-collector" \
    --namespace=kyma-system \
    --wait \
    --timeout=600s
}

installTestLogColletor
