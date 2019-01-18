#!/usr/bin/env bash

set -o errexit

discoverUnsetVar=false

for var in DOCKER_PUSH_REPOSITORY DOCKER_PUSH_DIRECTORY KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS SLACK_CLIENT_TOKEN SLACK_CLIENT_WEBHOOK_URL SLACK_CLIENT_CHANNEL_ID; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

function removeCluster() {
	# Set +e for testing purposes. This should be deleted only we move to daily schedule
	set +e
	
	COMMON_NAME=$1
	TIMESTAMP=$(echo "${COMMON_NAME}" | cut -d '-' -f 3)
	EXIT_STATUS=$?

	shout "Delete cluster $CLUSTER_NAME"
	CLUSTER_NAME=${COMMON_NAME} "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/deprovision-gke-cluster.sh
	TMP_STATUS=$?
	if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

	# ToDo Add deletion of IP/DNS

	KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-nightly/${REPO_OWNER}/${REPO_NAME}:${TIMESTAMP}" "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/delete-image.sh
	TMP_STATUS=$?
	if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

	MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    shout "Job is finished ${MSG}"
    date
}

function createCluster() {
	# Set +e for testing purposes. This should be deleted only we move to daily schedule
	set +e

	DNS_SUBDOMAIN="${COMMON_NAME}"
	shout "Build Kyma-Installer Docker image"
	date
	"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/create-image.sh

	shout "Reserve IP Address for Ingressgateway"
	date
	GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
	GATEWAY_IP_ADDRESS=$(IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/reserve-ip-address.sh)
	echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"
	export CLEANUP_GATEWAY_IP_ADDRESS="true"

	shout "Create DNS Record for Ingressgateway IP"
	date
	GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
	IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/create-dns-record.sh
	export CLEANUP_GATEWAY_DNS_RECORD="true"

	shout "Reserve IP Address for Remote Environments"
	date
	REMOTEENVS_IP_ADDRESS_NAME="remoteenvs-${COMMON_NAME}"
	REMOTEENVS_IP_ADDRESS=$(IP_ADDRESS_NAME=${REMOTEENVS_IP_ADDRESS_NAME} "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/reserve-ip-address.sh)
	echo "Created IP Address for Remote Environments: ${REMOTEENVS_IP_ADDRESS}"
	export CLEANUP_REMOTEENVS_IP_ADDRESS="true"

	shout "Create DNS Record for Remote Environments IP"
	date
	REMOTEENVS_DNS_FULL_NAME="gateway.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
	IP_ADDRESS=${REMOTEENVS_IP_ADDRESS} DNS_FULL_NAME=${REMOTEENVS_DNS_FULL_NAME} "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/create-dns-record.sh
	export CLEANUP_REMOTEENVS_DNS_RECORD="true"

	shout "Provision cluster: \"${CLUSTER_NAME}\""
	date
	export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
	if [ -z "$MACHINE_TYPE" ]; then
		export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
	fi
	if [ -z "${CLUSTER_VERSION}" ]; then
		export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
	fi
	export CLEANUP_CLUSTER="true"
	"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/provision-gke-cluster.sh
}

function installKyma() {
	DNS_SUBDOMAIN="${COMMON_NAME}"
	
	KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"
	INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
	INSTALLER_CONFIG="${KYMA_RESOURCES_DIR}/installer-config-cluster.yaml.tpl"
	INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"
	
	shout "Generate self-signed certificate"
	date
	DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
	CERT_KEY=$("${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/generate-self-signed-cert.sh)
	TLS_CERT=$(echo "${CERT_KEY}" | head -1)
	TLS_KEY=$(echo "${CERT_KEY}" | tail -1)

	shout "Apply Kyma config"
	date
	"${KYMA_SCRIPTS_DIR}"/concat-yamls.sh "${INSTALLER_YAML}" "${INSTALLER_CONFIG}" "${INSTALLER_CR}" \
		| sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" \
		| sed -e "s/__DOMAIN__/${DOMAIN}/g" \
		| sed -e "s/__REMOTE_ENV_IP__/${REMOTEENVS_IP_ADDRESS}/g" \
		| sed -e "s/__TLS_CERT__/${TLS_CERT}/g" \
		| sed -e "s/__TLS_KEY__/${TLS_KEY}/g" \
		| sed -e "s/__EXTERNAL_PUBLIC_IP__/${GATEWAY_IP_ADDRESS}/g" \
		| sed -e "s/__SKIP_SSL_VERIFY__/true/g" \
		| sed -e "s/__VERSION__/0.0.1/g" \
		| sed -e "s/__.*__//g" \
		| kubectl apply -f-

	shout "Trigger installation"
	date
	kubectl label installation/kyma-installation action=install
	"${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout 30m
}

function installStabilityChecker() {
	STATS_FAILING_TEST_REGEXP=${STATS_FAILING_TEST_REGEXP:-"'\"'([0-9A-Za-z_-]+)'\"' (?:has Failed status?|failed due to too long Running status?|failed due to too long Pending status?|failed with Unknown status?)"}
	STATS_SUCCESSFUL_TEST_REGEXP=${STATS_SUCCESSFUL_TEST_REGEXP:-"Test of '\"'([0-9A-Za-z_-]+)'\"' was successful"}
	STATS_ENABLED=true
	
	SC_DIR=${KYMA_SOURCES_DIR}/tools/stability-checker

	kubectl create -f "${SC_DIR}/local/provisioning.yaml"
	bash "${SC_DIR}/local/helpers/isready.sh" kyma-system app  stability-test-provisioner
	kubectl exec stability-test-provisioner -n kyma-system --  mkdir -p /home/input
	kubectl cp "${KYMA_SCRIPTS_DIR}/testing.sh" stability-test-provisioner:/home/input/ -n kyma-system
	kubectl cp "${KYMA_SCRIPTS_DIR}/utils.sh" stability-test-provisioner:/home/input/ -n kyma-system
	kubectl delete pod -n kyma-system stability-test-provisioner

	helm install --set clusterName="Nightly" \
	        --set slackClientWebhookUrl="${SLACK_CLIENT_WEBHOOK_URL}" \
	        --set slackClientChannelId="${SLACK_CLIENT_CHANNEL_ID}" \
	        --set slackClientToken="${SLACK_CLIENT_TOKEN}" \
	        --set stats.enabled="${STATS_ENABLED}" \
	        --set stats.failingTestRegexp="${STATS_FAILING_TEST_REGEXP}" \
	        --set stats.successfulTestRegexp="${STATS_SUCCESSFUL_TEST_REGEXP}" \
	        --set testResultWindowTime="3h" \
	        "${SC_DIR}/deploy/chart/stability-checker" \
	        --namespace=kyma-system \
	        --name=stability-checker
}

readonly REPO_OWNER="kyma-project"
export REPO_OWNER
readonly REPO_NAME="kyma"
export REPO_NAME

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"

readonly CURRENT_TIMESTAMP=$(date +%Y%m%d)

if [[ "$OSTYPE" == "darwin"* ]]; then
	readonly LAST_TIMESTAMP=$(date -v -1d '+%Y%m%d')
else
	readonly LAST_TIMESTAMP=$(date +%Y%m%d --date="yesterday")
fi

readonly NAME_ROOT="gkeint-nightly"
readonly COMMON_NAME=$(echo "${NAME_ROOT}-${CURRENT_TIMESTAMP}" | tr "[:upper:]" "[:lower:]")
readonly KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/${NAME_ROOT}/${REPO_OWNER}/${REPO_NAME}:${CURRENT_TIMESTAMP}"
export KYMA_INSTALLER_IMAGE

### Cluster name must be less than 40 characters!
export CLUSTER_NAME="${COMMON_NAME}"

### For provision-gke-cluster.sh
export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

shout "Authenticate"
date
init
readonly DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"
export DNS_DOMAIN

OLD_CLUSTERS=$(gcloud container clusters list --filter="name~${NAME_ROOT}" --format json | jq '.[].name' | tr -d '"')
CLUSTERS_SIZE=$(echo "$OLD_CLUSTERS" | wc -l)
if [[ "$CLUSTERS_SIZE" -gt 0 ]]; then
	shout "Delete old cluster"
	date
	for CLUSTER in $OLD_CLUSTERS; do
		removeCluster "${CLUSTER}"
	done
fi

shout "Build Kyma-Installer Docker image"
date
"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/create-image.sh

shout "Create new cluster"
date
createCluster

shout "Install tiller"
date
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
"${KYMA_SCRIPTS_DIR}"/install-tiller.sh

shout "Install kyma"
date
installKyma

shout "Test kyma"
date
"${KYMA_SCRIPTS_DIR}"/testing.sh

shout "Install stability-checker"
date
installStabilityChecker
