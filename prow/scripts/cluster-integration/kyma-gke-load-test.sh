#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.
discoverUnsetVar=false

for var in INPUT_CLUSTER_NAME DOCKER_PUSH_REPOSITORY DOCKER_PUSH_DIRECTORY KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_COMPUTE_ZONE CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS SAP_SLACK_BOT_TOKEN LT_TIMEOUT LT_REQS_PER_ROUTINE LOAD_TEST_SLACK_CLIENT_CHANNEL_ID; do
	if [ -z "${!var}" ] ; then
		echo "ERROR: $var is not set"
		discoverUnsetVar=true
	fi
done
if [ "${discoverUnsetVar}" = true ] ; then
	exit 1
fi

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"

export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"
export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"

readonly REPO_OWNER="kyma-project"
readonly REPO_NAME="kyma"
readonly CURRENT_TIMESTAMP=$(date +%Y%m%d)

readonly STANDARIZED_NAME=$(echo "${INPUT_CLUSTER_NAME}" | tr "[:upper:]" "[:lower:]")
readonly DNS_SUBDOMAIN="${STANDARIZED_NAME}"

export CLUSTER_NAME="${STANDARIZED_NAME}"
export GCLOUD_NETWORK_NAME="load-test-net"
export GCLOUD_SUBNET_NAME="load-test-subnet"
export STANDARIZED_NAME
export REPO_OWNER
export REPO_NAME
export CURRENT_TIMESTAMP

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

function waitUntilHPATestIsDone {
	local succeeded='false'
	while true; do
		hpa_test_status=$(kubectl get po -n kyma-system load-test -ojsonpath="{.status.phase}")
		if [ "${hpa_test_status}" != "Running" ]; then
			if [ "${hpa_test_status}" == "Succeeded" ]; then
				succeeded='true'
			fi
			break
		fi
		sleep 5
		echo "HPA tests are in progress..."
	done

	if [ $succeeded == 'true' ]; then
		shout "Load test successfully completed!"
	else
		kubectl describe pod load-test -n kyma-system
		shout "Load test failed!"
		shout "Please check logs by executing: kubectl logs -n kyma-system load-test"
		exit 1
	fi
}

function installLoadTest() {
	LT_FOLDER=${KYMA_SOURCES_DIR}/tools/load-test
	LT_FOLDER_CHART=${LT_FOLDER}'/deploy/chart/load-test'

	shout "Executing load tests..."

	shout "Installing helm chart..."
	helm install --set slackClientToken="${SAP_SLACK_BOT_TOKEN}" \
				--set slackClientChannelId="${LOAD_TEST_SLACK_CLIENT_CHANNEL_ID}" \
				--set loadTestExecutionTimeout="${LT_TIMEOUT}" \
				--set reqsPerRoutine="${LT_REQS_PER_ROUTINE}" \
				"${LT_FOLDER_CHART}" \
				--namespace=kyma-system \
				--name=load-test \
				--timeout=700 \
				--wait \
				--tls

	waitUntilHPATestIsDone
	
}


shout "Authenticate"
date
init

DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"
export DNS_DOMAIN
DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
export DOMAIN

shout "Cleanup"
date
# shellcheck disable=SC1090
source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/cleanup-cluster.sh

shout "Create new cluster"
date
# shellcheck disable=SC1090
source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-cluster.sh

shout "Install tiller"
date

shout "Account is:"
gcloud config get-value account

kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
"${KYMA_SCRIPTS_DIR}"/install-tiller.sh

shout "Install kyma"
date
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/install-kyma.sh
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-helm-certs.sh"

shout "Install load-test"
date
installLoadTest

shout "Cleanup after load test"
date
# shellcheck disable=SC1090
source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/cleanup-cluster.sh
