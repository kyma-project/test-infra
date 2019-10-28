#!/usr/bin/env bash

set -o errexit

VARIABLES=(
   TEST_INFRA_SOURCES_DIR
   KYMA_SCRIPTS_DIR
   TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS
   CLUSTER_NAME
   SLACK_CLIENT_WEBHOOK_URL
   STABILITY_SLACK_CLIENT_CHANNEL_ID
   SLACK_CLIENT_TOKEN
   TEST_RESULT_WINDOW_TIME
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

function installStabilityChecker() {
	SC_DIR=${TEST_INFRA_SOURCES_DIR}/stability-checker

	kubectl create -f "${SC_DIR}/local/provisioning.yaml"
	bash "${SC_DIR}/local/helpers/isready.sh" kyma-system app  stability-test-provisioner
	kubectl exec stability-test-provisioner -n kyma-system --  mkdir -p /home/input
	kubectl cp "${KYMA_SCRIPTS_DIR}/testing.sh" stability-test-provisioner:/home/input/ -n kyma-system
	kubectl cp "${KYMA_SCRIPTS_DIR}/utils.sh" stability-test-provisioner:/home/input/ -n kyma-system
	kubectl cp "${KYMA_SCRIPTS_DIR}/testing-common.sh" stability-test-provisioner:/home/input/ -n kyma-system
    kubectl cp "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-helm-certs.sh" stability-test-provisioner:/home/input/pre-start-scripts.sh -n kyma-system
	kubectl delete pod -n kyma-system stability-test-provisioner

	STATS_FAILING_TEST_REGEXP=${STATS_FAILING_TEST_REGEXP:-"Test status: ([0-9A-Za-z_-]+) - Failed"}
	STATS_SUCCESSFUL_TEST_REGEXP=${STATS_SUCCESSFUL_TEST_REGEXP:-"Test status: ([0-9A-Za-z_-]+) - Succeeded"}

    STATS_ENABLED="true"
    # create a secret with service account used for storing logs
    kubectl create secret generic sa-stability-fluentd-storage-writer --from-file=service-account.json=/etc/credentials/sa-stability-fluentd-storage-writer/service-account.json -n kyma-system

	helm install --set clusterName="${CLUSTER_NAME}" \
	        --set logsPersistence.enabled=true \
	        --set slackClientWebhookUrl="${SLACK_CLIENT_WEBHOOK_URL}" \
	        --set slackClientChannelId="${STABILITY_SLACK_CLIENT_CHANNEL_ID}" \
	        --set slackClientToken="${SLACK_CLIENT_TOKEN}" \
	        --set stats.enabled="${STATS_ENABLED}" \
	        --set stats.failingTestRegexp="${STATS_FAILING_TEST_REGEXP}" \
	        --set stats.successfulTestRegexp="${STATS_SUCCESSFUL_TEST_REGEXP}" \
	        --set testResultWindowTime="${TEST_RESULT_WINDOW_TIME}" \
	        "${SC_DIR}/deploy/chart/stability-checker" \
	        --namespace=kyma-system \
	        --name=stability-checker \
	        --wait \
	        --timeout=600 \
	        --tls
}

installStabilityChecker