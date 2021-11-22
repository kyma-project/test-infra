#!/usr/bin/env bash

set -o errexit   # Exit immediately if a command exits with a non-zero status.
set -o pipefail  # Fail a pipe if any sub-command fails.

# INIT ENVIRONMENT VARIABLES
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
export KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

requiredVars=(
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
	SLACK_CLIENT_WEBHOOK_URL
	STABILITY_SLACK_CLIENT_CHANNEL_ID
	SLACK_CLIENT_TOKEN
	TEST_RESULT_WINDOW_TIME
	DOCKER_PUSH_REPOSITORY
	DOCKER_PUSH_DIRECTORY
)

utils::check_required_vars "${requiredVars[@]}"

readonly STANDARIZED_NAME=$(echo "${INPUT_CLUSTER_NAME}" | tr "[:upper:]" "[:lower:]")
readonly DNS_SUBDOMAIN="${STANDARIZED_NAME}"

export CLUSTER_NAME="${STANDARIZED_NAME}"
export CLUSTER_SIZE="Standard_F8s_v2"

export CLUSTER_ADDONS="monitoring,http_application_routing"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/azure.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/azure.sh"
# shellcheck source=prow/scripts/lib/gcloud.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcloud.sh"
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
			gcloud::delete_dns_record "${GATEWAY_IP_ADDRESS}" "${GATEWAY_DNS_FULL_NAME}"
			TMP_STATUS=$?
			if [[ ${TMP_STATUS} -ne 0 ]]; then
			  log::error "Failed delete dns record : ${GATEWAY_DNS_FULL_NAME}"
			  EXIT_STATUS=${TMP_STATUS}
			else
			  log::success "Deleted dns record : ${GATEWAY_DNS_FULL_NAME}"
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
			gcloud::delete_dns_record "${APISERVER_IP_ADDRESS}" "${APISERVER_DNS_FULL_NAME}"
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

function installCluster() {
	log::info "Install Kubernetes on Azure"

	log::info "Find latest cluster version for kubernetes version: ${AKS_CLUSTER_VERSION}"
	AKS_CLUSTER_VERSION_PRECISE=$(az aks get-versions -l "${REGION}" | jq '.orchestrators|.[]|select(.orchestratorVersion | contains("'"${AKS_CLUSTER_VERSION}"'"))' | jq -s '.' | jq -r 'sort_by(.orchestratorVersion | split(".") | map(tonumber)) | .[-1].orchestratorVersion')
	log::info "Latest available version is: ${AKS_CLUSTER_VERSION_PRECISE}"

	az aks create \
	  --resource-group "${RS_GROUP}" \
	  --name "${CLUSTER_NAME}" \
	  --node-count 3 \
	  --node-vm-size "${CLUSTER_SIZE}" \
	  --kubernetes-version "${AKS_CLUSTER_VERSION_PRECISE}" \
	  --enable-addons "${CLUSTER_ADDONS}" \
	  --service-principal "$(jq -r '.app_id' "$AZURE_CREDENTIALS_FILE")" \
	  --client-secret "$(jq -r '.secret' "$AZURE_CREDENTIALS_FILE")" \
	  --generate-ssh-keys \
	  --zones 1 2 3
}

function createPublicIPandDNS() {
	CLUSTER_RS_GROUP=$(az aks show -g "${RS_GROUP}" -n "${CLUSTER_NAME}" --query nodeResourceGroup -o tsv)

	# IP address and DNS for Ingressgateway
	log::info "Reserve IP Address for Ingressgateway"

	GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"
	az network public-ip create -g "${CLUSTER_RS_GROUP}" -n "${GATEWAY_IP_ADDRESS_NAME}" -l "${REGION}" --allocation-method static --sku Standard

	export GATEWAY_IP_ADDRESS
	GATEWAY_IP_ADDRESS=$(az network public-ip show -g "${CLUSTER_RS_GROUP}" -n "${GATEWAY_IP_ADDRESS_NAME}" --query ipAddress -o tsv)
	log::success "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

	log::info "Create DNS Record for Ingressgateway IP"

	GATEWAY_DNS_FULL_NAME="*.${DOMAIN}."
	gcloud::create_dns_record "${GATEWAY_IP_ADDRESS}" "${GATEWAY_DNS_FULL_NAME}"
}

function setupKubeconfig() {
	log::info "Setup kubeconfig and create ClusterRoleBinding"

	az aks get-credentials --resource-group "${RS_GROUP}" --name "${CLUSTER_NAME}"
	kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(az account show | jq -r .user.name)"
}

function installKyma() {

	log::info "Prepare Kyma overrides"

	export DEX_CALLBACK_URL="https://dex.${DOMAIN}/callback"

	KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

	log::info "Apply Azure disable knative-eventing stdout logging"
	kubectl apply -f "${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/azure-knative-eventing-logging.yaml"
	log::info "Apply Azure crb for healthz"
	kubectl apply -f "${KYMA_RESOURCES_DIR}"/azure-crb-for-healthz.yaml

  envsubst < "${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/kyma-installer-overrides.tpl.yaml" > "$PWD/kyma-installer-overrides.yaml"
  envsubst < "${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/overrides-dex-and-monitoring.tpl.yaml" > "$PWD/overrides-dex-and-monitoring.yaml"

	log::info "Trigger installation"

    kyma install \
			--ci \
			--source "1.25.0" \
      -o "$PWD/kyma-installer-overrides.yaml" \
			-o "$PWD/overrides-dex-and-monitoring.yaml" \
			-o "${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/prometheus-cluster-essentials-overrides.tpl.yaml" \
			--domain "${DOMAIN}" \
			--profile production \
			--tls-cert "${TLS_CERT}" \
			--tls-key "${TLS_KEY}" \
			--timeout 60m

	if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
		log::info "Create DNS Record for Apiserver proxy IP"
		APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
		APISERVER_DNS_FULL_NAME="apiserver.${DOMAIN}."
		gcloud::create_dns_record "${APISERVER_IP_ADDRESS}" "${APISERVER_DNS_FULL_NAME}"
	fi
}

function apply_dex_github_kyma_admin_group() {
    kubectl get ClusterRoleBinding kyma-admin-binding -oyaml > kyma-admin-binding.yaml && cat >> kyma-admin-binding.yaml <<EOF 
- apiGroup: rbac.authorization.k8s.io
  kind: Group
  name: kyma-project:cluster-access
EOF

    kubectl replace -f kyma-admin-binding.yaml
}

function test_console_url() {
  CONSOLE_URL="https://console.${DOMAIN}"
  console_response=$(curl -L -s -o /dev/null -w "%{http_code}" "${CONSOLE_URL}")
  if [ "${console_response}" != "200" ]; then
    log::error "Kyma console URL did not returned 200 HTTP response code. Check ingressgateway service."
    exit 1
  fi
}

gcloud::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"
docker::start
az::login "$AZURE_CREDENTIALS_FILE"
az::set_subscription "$AZURE_SUBSCRIPTION_ID"

DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --project="${CLOUDSDK_CORE_PROJECT}" --format="value(dnsName)")"
export DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"

cleanup

az::create_resource_group "${RS_GROUP}" "${REGION}"
installCluster

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
apply_dex_github_kyma_admin_group

log::info "Install stability-checker"
(
export TEST_INFRA_SOURCES_DIR KYMA_SCRIPTS_DIR TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS \
		CLUSTER_NAME SLACK_CLIENT_WEBHOOK_URL STABILITY_SLACK_CLIENT_CHANNEL_ID SLACK_CLIENT_TOKEN TEST_RESULT_WINDOW_TIME
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/install-stability-checker.sh"
)

test_console_url
