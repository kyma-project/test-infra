#!/usr/bin/env bash

set -o errexit   # Exit immediately if a command exits with a non-zero status.
set -o pipefail  # Fail a pipe if any sub-command fails.

# INIT ENVIRONMENT VARIABLES
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

requiredVars=(
	AKS_CLUSTER_VERSION
	RS_GROUP
	REGION
	AZURE_SUBSCRIPTION_ID
	AZURE_CREDENTIALS_FILE
	KYMA_PROJECT_DIR
	INPUT_CLUSTER_NAME
	GOOGLE_APPLICATION_CREDENTIALS
	CLOUDSDK_DNS_ZONE_NAME
	CLOUDSDK_CORE_PROJECT
	KYMA_ALERTS_CHANNEL
	KYMA_ALERTS_SLACK_API_URL
	# SLACK_CLIENT_WEBHOOK_URL
	# STABILITY_SLACK_CLIENT_CHANNEL_ID
	# SLACK_CLIENT_TOKEN
	# TEST_RESULT_WINDOW_TIME
	DOCKER_PUSH_REPOSITORY
	DOCKER_PUSH_DIRECTORY
)

utils::check_required_vars "${requiredVars[@]}"

readonly COMMON_NAME=$(echo "${INPUT_CLUSTER_NAME}" | tr "[:upper:]" "[:lower:]")
readonly DNS_SUBDOMAIN="${COMMON_NAME}"

export CLUSTER_NAME="${COMMON_NAME}"
export CLUSTER_SIZE="Standard_F8s_v2"

export CLUSTER_ADDONS="monitoring,http_application_routing"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/azure.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/azure.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcp.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"

function check_status() {
  status="${1}"
  error_desc="${2}"
  if [[ ${status} -ne 0 ]]; then
    EXIT_STATUS=${status}
    log::error "${error_desc}"
  fi
}

function cleanup() {
	log::info "Cleanup"

	# Turn off exit-on-error so that next step is executed even if previous one fails.
	# Cleanup is best-effort since we don't know in which state the previous cluster is, if there is any.
	set +e
	EXIT_STATUS=$?

	# Exporting for use in subshells.
	export RS_GROUP

	az::get_cluster_resource_group \
		-r "$RS_GROUP" \
		-c "$CLUSTER_NAME"

	log::info '\n---\nRemove DNS Record for Ingressgateway\n---'
	GATEWAY_DNS_FULL_NAME="*.${DOMAIN}."

	GATEWAY_IP_ADDRESS=$(gcloud dns record-sets list --zone "${CLOUDSDK_DNS_ZONE_NAME}" --name "${GATEWAY_DNS_FULL_NAME}" --format="value(rrdatas[0])")
	TMP_STATUS=$?
	check_status ${TMP_STATUS} "Could not fetch IP for : ${GATEWAY_DNS_FULL_NAME}"
	if [[ -n ${GATEWAY_IP_ADDRESS} ]];then
		# only try to delete the dns record if the ip address has been found
		gcp::delete_dns_record \
			-a "$GATEWAY_IP_ADDRESS" \
			-p "$CLOUDSDK_CORE_PROJECT" \
			-h "*" \
			-s "$DNS_SUBDOMAIN" \
			-z "$CLOUDSDK_DNS_ZONE_NAME"
	fi

	log::info '\n---\nRemove DNS Record for Apiserver Proxy IP\n---'
	APISERVER_DNS_FULL_NAME="apiserver.${DOMAIN}."
	APISERVER_IP_ADDRESS=$(gcloud dns record-sets list --zone "${CLOUDSDK_DNS_ZONE_NAME}" --name "${APISERVER_DNS_FULL_NAME}" --format="value(rrdatas[0])")
	TMP_STATUS=$?
	check_status ${TMP_STATUS} "Could not fetch IP for : ${APISERVER_DNS_FULL_NAME}"
	if [[ -n ${APISERVER_IP_ADDRESS} ]]; then
		gcp::delete_dns_record \
			-a "$APISERVER_IP_ADDRESS" \
			-p "$CLOUDSDK_CORE_PROJECT" \
			-h "apiserver" \
			-s "$DNS_SUBDOMAIN" \
			-z "$CLOUDSDK_DNS_ZONE_NAME"
	fi

	# Exporting for use in subshells.
	export RS_GROUP
	if [[ $(az group exists --name "${RS_GROUP}" -o json) == true ]]; then
		az::deprovision_k8s_cluster \
			-c "$CLUSTER_NAME"\
			-g "$RS_GROUP"

		az::delete_resource_group \
			-g "$RS_GROUP"
	else
		log::info "Azure group does not exist, skip cleanup process"
	fi

	MSG=""
	if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
	log::info "\\n---\\nCleanup function is finished ${MSG}\\n---"

	# Turn on exit-on-error
	set -e
}

function createPublicIPandDNS() {
	az::get_cluster_resource_group \
		-r "$RS_GROUP" \
		-c "$CLUSTER_NAME"
	CLUSTER_RS_GROUP="${az_get_cluster_resource_group_return_resource_group:?}"

	# IP address and DNS for Ingressgateway
	az::reserve_ip_address \
		-g "$CLUSTER_RS_GROUP" \
		-n "$COMMON_NAME" \
		-r "$REGION"
	GATEWAY_IP_ADDRESS="${az_reserve_ip_address_return_ip_address:?}"
	export GATEWAY_IP_ADDRESS

	log::info "Create DNS Record for Ingressgateway IP"

	gcp::create_dns_record \
				-a "$GATEWAY_IP_ADDRESS" \
				-p "$CLOUDSDK_CORE_PROJECT" \
				-h "*" \
				-s "$DNS_SUBDOMAIN" \
				-z "$CLOUDSDK_DNS_ZONE_NAME"
}

function setupKubeconfig() {
	log::info "Setup kubeconfig and create ClusterRoleBinding"

	az aks get-credentials --resource-group "${RS_GROUP}" --name "${CLUSTER_NAME}"
	kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(az account show | jq -r .user.name)"
}

function installKyma() {

	log::info "Prepare Kyma overrides"

	KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

	log::info "Apply Azure disable knative-eventing stdout logging"
	kubectl apply -f "${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/azure-knative-eventing-logging.yaml"
	log::info "Apply Azure crb for healthz"
	kubectl apply -f "${KYMA_RESOURCES_DIR}"/azure-crb-for-healthz.yaml

  envsubst < "${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/kyma-installer-overrides.tpl.yaml" > "$PWD/kyma-installer-overrides.yaml"

	log::info "Trigger installation"

	kyma install \
			--ci \
			--source main \
			-o "$PWD/kyma-installer-overrides.yaml" \
			-o "${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/prometheus-cluster-essentials-overrides.tpl.yaml" \
			--domain "${DOMAIN}" \
			--profile production \
			--tls-cert "${TLS_CERT}" \
			--tls-key "${TLS_KEY}"

	if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
		log::info "Create DNS Record for Apiserver proxy IP"
		APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
		gcp::create_dns_record \
				-a "$APISERVER_IP_ADDRESS" \
				-p "$CLOUDSDK_CORE_PROJECT" \
				-h "apiserver" \
				-s "$DNS_SUBDOMAIN" \
				-z "$CLOUDSDK_DNS_ZONE_NAME"
	fi
}

function test_console_url() {
  CONSOLE_URL="https://console.${DOMAIN}"
  console_response=$(curl -L -s -o /dev/null -w "%{http_code}" "${CONSOLE_URL}")
  if [ "${console_response}" != "200" ]; then
	log::error "Kyma console URL did not returned 200 HTTP response code. Check ingressgateway service."
	exit 1
  fi
}

gcp::authenticate \
	-c "${GOOGLE_APPLICATION_CREDENTIALS}"
docker::start
az::authenticate \
	-f "$AZURE_CREDENTIALS_FILE"
az::set_subscription \
	-s "$AZURE_SUBSCRIPTION_ID"

DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --project="${CLOUDSDK_CORE_PROJECT}" --format="value(dnsName)")"
export DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"

cleanup

az::create_resource_group \
	-g "${RS_GROUP}" \
	-r "${REGION}"

log::info "Install Kubernetes on Azure"

# shellcheck disable=SC2153
az::provision_k8s_cluster \
	-c "$CLUSTER_NAME" \
	-g "$RS_GROUP" \
	-r "$REGION" \
	-s "$CLUSTER_SIZE" \
	-v "$AKS_CLUSTER_VERSION" \
	-a "$CLUSTER_ADDONS" \
	-f "$AZURE_CREDENTIALS_FILE"

createPublicIPandDNS
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-letsencrypt-cert.sh"
TLS_CERT=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/fullchain.pem | tr -d '\n')
export TLS_CERT
TLS_KEY=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/privkey.pem   | tr -d '\n')
export TLS_KEY

setupKubeconfig

kyma::install_cli

installKyma


log::info "Override kyma-admin-binding ClusterRoleBinding"

test_console_url
