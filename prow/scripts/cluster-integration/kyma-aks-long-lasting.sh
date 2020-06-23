#!/usr/bin/env bash

set -o errexit   # Exit immediately if a command exits with a non-zero status.
set -o pipefail  # Fail a pipe if any sub-command fails.

VARIABLES=(
	RS_GROUP
	REGION
	AZURE_SUBSCRIPTION_ID
	AZURE_SUBSCRIPTION_APP_ID
	AZURE_SUBSCRIPTION_SECRET
	AZURE_SUBSCRIPTION_TENANT
	KYMA_PROJECT_DIR
	INPUT_CLUSTER_NAME
	GOOGLE_APPLICATION_CREDENTIALS
	CLOUDSDK_DNS_ZONE_NAME
	CLOUDSDK_CORE_PROJECT
	KYMA_ALERTS_CHANNEL
	KYMA_ALERTS_SLACK_API_URL
	SLACK_CLIENT_WEBHOOK_URL
	STABILITY_SLACK_CLIENT_CHANNEL_ID
	SLACK_CLIENT_TOKEN
	TEST_RESULT_WINDOW_TIME
	DOCKER_PUSH_REPOSITORY
	DOCKER_PUSH_DIRECTORY
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

# INIT ENVIRONMENT VARIABLES
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
export KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

readonly STANDARIZED_NAME=$(echo "${INPUT_CLUSTER_NAME}" | tr "[:upper:]" "[:lower:]")
readonly DNS_SUBDOMAIN="${STANDARIZED_NAME}"

export CLUSTER_NAME="${STANDARIZED_NAME}"
export CLUSTER_SIZE="Standard_F8s_v2"
# set cluster version as MAJOR.MINOR without PATCH part (e.g. 1.10, 1.11)
export DEFAULT_CLUSTER_VERSION="1.15"
if [ -z "${CLUSTER_VERSION}" ]; then
    export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
fi

export CLUSTER_ADDONS="monitoring,http_application_routing"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/kyma-cli.sh"

function check_status() {
  status="${1}"
  error_desc="${2}"
  if [[ ${status} -ne 0 ]]; then
    EXIT_STATUS=${status}
    log::error "${error_desc}"
  fi
}

function cleanup() {
	shout "Cleanup"
	date

	# Turn off exit-on-error so that next step is executed even if previous one fails.
	# Cleanup is best-effort since we don't know in which state the previous cluster is, if there is any.
	set +e
	EXIT_STATUS=$?

	# Exporting for use in subshells.
	export RS_GROUP

	if [[ $(az group exists --name "${RS_GROUP}" -o json) == true ]]; then
		CLUSTER_RS_GROUP=$(az aks show -g "${RS_GROUP}" -n "${CLUSTER_NAME}" --query nodeResourceGroup -o tsv)
		TMP_STATUS=$?
		check_status "${TMP_STATUS}" "Failed to get nodes resource group."

		log::info "\n---\nRemove DNS Record for Ingressgateway\n---"
		GATEWAY_DNS_FULL_NAME="*.${DOMAIN}."
		GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"

		GATEWAY_IP_ADDRESS=$(gcloud dns record-sets list --zone "${CLOUDSDK_DNS_ZONE_NAME}" --name "${GATEWAY_DNS_FULL_NAME}" --format="value(rrdatas[0])")
		TMP_STATUS=$?
		check_status ${TMP_STATUS} "Could not fetch IP for : ${GATEWAY_DNS_FULL_NAME}"
		if [[ -n ${GATEWAY_IP_ADDRESS} ]];then
			log::success "Fetched Azure Gateway IP: ${GATEWAY_IP_ADDRESS}"
			# only try to delete the dns record if the ip address has been found
			"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${GATEWAY_DNS_FULL_NAME}" --address="${GATEWAY_IP_ADDRESS}" --dryRun=false
			TMP_STATUS=$?
			if [[ ${TMP_STATUS} -ne 0 ]]; then
			  log::error "Failed delete dns record : ${GATEWAY_DNS_FULL_NAME}"
			  EXIT_STATUS=${TMP_STATUS}
			else
			  log:success "Deleted dns record : ${GATEWAY_DNS_FULL_NAME}"
			fi
		else
			log::warn "Could not delete DNS record : ${GATEWAY_DNS_FULL_NAME}. Record does not exist."
		fi

		log::info "\n---\nRemove DNS Record for Apiserver Proxy IP\n---"
		APISERVER_DNS_FULL_NAME="apiserver.${DOMAIN}."
		APISERVER_IP_ADDRESS=$(gcloud dns record-sets list --zone "${CLOUDSDK_DNS_ZONE_NAME}" --name "${APISERVER_DNS_FULL_NAME}" --format="value(rrdatas[0])")
		TMP_STATUS=$?
		check_status ${TMP_STATUS} "Could not fetch IP for : ${APISERVER_DNS_FULL_NAME}"
		if [[ -n ${APISERVER_IP_ADDRESS} ]]; then
			"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${APISERVER_DNS_FULL_NAME}" --address="${APISERVER_IP_ADDRESS}" --dryRun=false
			TMP_STATUS=$?
			if [[ ${TMP_STATUS} -ne 0 ]]; then
			  log::error "Failed delete dns record : ${APISERVER_DNS_FULL_NAME}"
			  EXIT_STATUS=${TMP_STATUS}
			else
			  log::success "Deleted dns record : ${APISERVER_DNS_FULL_NAME}"
			fi
		else
		  log::warn "Could not delete DNS record ${APISERVER_DNS_FULL_NAME}. Record does not exist."
		fi

		log::info "\n---\nRemove Cluster, IP Address for Ingressgateway\n---"
		az aks delete -g "${RS_GROUP}" -n "${CLUSTER_NAME}" -y
		TMP_STATUS=$?
		if [[ ${TMP_STATUS} -ne 0 ]]; then
		  log::error "Failed delete cluster : ${CLUSTER_NAME}"
		  EXIT_STATUS=${TMP_STATUS}
		else
		  log::success "Cluster and IP address for Ingressgateway deleted"
		fi

		log::info "Remove group"
		az group delete -n "${RS_GROUP}" -y
		TMP_STATUS=$?
		if [[ ${TMP_STATUS} -ne 0 ]]; then
		  log::error "Failed to delete ResourceGrouop : ${RS_GROUP}"
		  EXIT_STATUS=${TMP_STATUS}
		else
		  log::success "ResourceGroup deleted : ${RS_GROUP}"
		fi
	else
		log::info "Azure group does not exist, skip cleanup process"
	fi

	MSG=""
	if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
	log::info "\n---\nCleanup function is finished ${MSG}\n---"

	# Turn on exit-on-error
	set -e
}

function createGroup() {
	shout "Create Azure group"
	date

	# Export variable for use in subshells.
	export RS_GROUP

	az group create --name "${RS_GROUP}" --location "${REGION}"

	# Wait until resource group will be visible in azure.
	counter=0
	until [[ $(az group exists --name "${RS_GROUP}" -o json) == true ]]; do
		sleep 15
		counter=$(( counter + 1 ))
		if (( counter == 5 )); then
			log::info "\n---\nAzure resource group ${RS_GROUP} still not present after one minute wait.\n---"
			exit 1
		fi
	done
}

function installCluster() {
	shout "Install Kubernetes on Azure"
	date

	log::info "Find latest cluster version for kubernetes version: ${CLUSTER_VERSION}"
	AKS_CLUSTER_VERSION=$(az aks get-versions -l "${REGION}" | jq '.orchestrators|.[]|select(.orchestratorVersion | contains("'"${CLUSTER_VERSION}"'"))' | jq -s '.' | jq -r 'sort_by(.orchestratorVersion | split(".") | map(tonumber)) | .[-1].orchestratorVersion')
	log::info "Latest available version is: ${AKS_CLUSTER_VERSION}"

	az aks create \
	  --resource-group "${RS_GROUP}" \
	  --name "${CLUSTER_NAME}" \
	  --node-count 3 \
	  --node-vm-size "${CLUSTER_SIZE}" \
	  --kubernetes-version "${AKS_CLUSTER_VERSION}" \
	  --enable-addons "${CLUSTER_ADDONS}" \
	  --service-principal "${AZURE_SUBSCRIPTION_APP_ID}" \
	  --client-secret "${AZURE_SUBSCRIPTION_SECRET}" \
	  --generate-ssh-keys \
	  --zones 1 2 3
}

function azureAuthenticating() {
	shout "Authenticating to azure"
	date

	az login --service-principal -u "${AZURE_SUBSCRIPTION_APP_ID}" -p "${AZURE_SUBSCRIPTION_SECRET}" --tenant "${AZURE_SUBSCRIPTION_TENANT}"
	az account set --subscription "${AZURE_SUBSCRIPTION_ID}"
}

function createPublicIPandDNS() {
	CLUSTER_RS_GROUP=$(az aks show -g "${RS_GROUP}" -n "${CLUSTER_NAME}" --query nodeResourceGroup -o tsv)

	# IP address and DNS for Ingressgateway
	shout "Reserve IP Address for Ingressgateway"
	date

	GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"
	az network public-ip create -g "${CLUSTER_RS_GROUP}" -n "${GATEWAY_IP_ADDRESS_NAME}" -l "${REGION}" --allocation-method static --sku Standard

	GATEWAY_IP_ADDRESS=$(az network public-ip show -g "${CLUSTER_RS_GROUP}" -n "${GATEWAY_IP_ADDRESS_NAME}" --query ipAddress -o tsv)
	log::success "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

	shout "Create DNS Record for Ingressgateway IP"
	date

	GATEWAY_DNS_FULL_NAME="*.${DOMAIN}."
	IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-dns-record.sh
}

function setupKubeconfig() {
	shout "Setup kubeconfig and create ClusterRoleBinding"
	date

	az aks get-credentials --resource-group "${RS_GROUP}" --name "${CLUSTER_NAME}"
	kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(az account show | jq -r .user.name)"
}

function installKyma() {

	shout "Prepare Kyma overrides"
	date

	componentOverridesFile="component-overrides.yaml"
	componentOverrides=$(cat << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: "installation-config-overrides"
  namespace: "kyma-installer"
  labels:
    installer: overrides
    kyma-project.io/installation: ""
data:
  global.loadBalancerIP: "${GATEWAY_IP_ADDRESS}"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: "core-test-ui-acceptance-overrides"
  namespace: "kyma-installer"
  labels:
    installer: overrides
    kyma-project.io/installation: ""
    component: core
data:
  console.test.acceptance.ui.logging.enabled: "true"
  console.test.acceptance.enabled: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: "application-registry-overrides"
  namespace: "kyma-installer"
  labels:
    installer: overrides
    kyma-project.io/installation: ""
    component: application-connector
data:
  application-registry.deployment.args.detailedErrorResponse: "true"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: "monitoring-config-overrides"
  namespace: "kyma-installer"
  labels:
    installer: overrides
    kyma-project.io/installation: ""
    component: monitoring
data:
  global.alertTools.credentials.slack.channel: "${KYMA_ALERTS_CHANNEL}"
  global.alertTools.credentials.slack.apiurl: "${KYMA_ALERTS_SLACK_API_URL}"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: "istio-overrides"
  namespace: "kyma-installer"
  labels:
    installer: overrides
    kyma-project.io/installation: ""
    component: istio
data:
  gateways.istio-ingressgateway.loadBalancerIP: "${GATEWAY_IP_ADDRESS}"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: dex-config-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
    component: dex
data:
 connectors: |
  - type: github
    id: github
    name: GitHub
    config:
      clientID: "${GITHUB_INTEGRATION_APP_CLIENT_ID}"
      clientSecret: "${GITHUB_INTEGRATION_APP_CLIENT_SECRET}"
      redirectURI: "https://dex.${DOMAIN}/callback"
      orgs:
      - name: kyma-project
EOF
)
  echo "${componentOverrides}" > "${componentOverridesFile}"

	KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

	log::info "Apply Azure crb for healthz"
	kubectl apply -f "${KYMA_RESOURCES_DIR}"/azure-crb-for-healthz.yaml

	shout "Trigger installation"
	date

  kyma install \
			--ci \
			--source latest-published \
			-o "${KYMA_RESOURCES_DIR}"/installer-config-production.yaml.tpl \
			-o "${componentOverridesFile}" \
			--domain "${DOMAIN}" \
			--tlsCert "${TLS_CERT}" \
			--tlsKey "${TLS_KEY}" \
			--timeout 60m

	if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
		shout "Create DNS Record for Apiserver proxy IP"
		date
		APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
		APISERVER_DNS_FULL_NAME="apiserver.${DOMAIN}."
		IP_ADDRESS=${APISERVER_IP_ADDRESS} DNS_FULL_NAME=${APISERVER_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
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
init
azureAuthenticating

DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --project="${CLOUDSDK_CORE_PROJECT}" --format="value(dnsName)")"
export DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"

cleanup

createGroup
installCluster

createPublicIPandDNS
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-letsencrypt-cert.sh"
TLS_CERT=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/fullchain.pem | tr -d '\n')
export TLS_CERT
TLS_KEY=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/privkey.pem   | tr -d '\n')
export TLS_KEY

setupKubeconfig

export INSTALL_DIR=${TMP_DIR}
install::kyma_cli

installKyma
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-helm-certs.sh"

shout "Override kyma-admin-binding ClusterRoleBinding"
applyDexGithibKymaAdminGroup

shout "Install stability-checker"
date
(
export TEST_INFRA_SOURCES_DIR KYMA_SCRIPTS_DIR TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS \
		CLUSTER_NAME SLACK_CLIENT_WEBHOOK_URL STABILITY_SLACK_CLIENT_CHANNEL_ID SLACK_CLIENT_TOKEN TEST_RESULT_WINDOW_TIME
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/install-stability-checker.sh"
)

test_console_url
