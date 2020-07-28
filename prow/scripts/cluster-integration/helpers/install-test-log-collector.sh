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
	TLC_DIR="${TEST_INFRA_SOURCES_DIR}/development/test-log-collector"

    if [  -f "$(helm home)/ca.pem" ]; then
        local HELM_ARGS="--tls"
    fi

	helm install --name test-log-collector --set slackToken="${LOG_COLLECTOR_SLACK_TOKEN}" \
	        "${TLC_DIR}/chart/test-log-collector" \
	        --namespace=kyma-system \
	        --wait ${HELM_ARGS} \
	        --timeout=600 \
            --tls
}

installTestLogColletor
