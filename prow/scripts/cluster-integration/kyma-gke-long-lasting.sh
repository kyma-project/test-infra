#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"

# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"

requiredVars=(
	INPUT_CLUSTER_NAME
	DOCKER_PUSH_REPOSITORY
	DOCKER_PUSH_DIRECTORY
	KYMA_PROJECT_DIR
	CLOUDSDK_CORE_PROJECT
	CLOUDSDK_COMPUTE_REGION
	CLOUDSDK_COMPUTE_ZONE
	CLOUDSDK_DNS_ZONE_NAME
	GOOGLE_APPLICATION_CREDENTIALS
	# SLACK_CLIENT_TOKEN
	# SLACK_CLIENT_WEBHOOK_URL
	# STABILITY_SLACK_CLIENT_CHANNEL_ID
	STACKDRIVER_COLLECTOR_SIDECAR_IMAGE_TAG
	CERTIFICATES_BUCKET
	GKE_CLUSTER_VERSION
)

utils::check_required_vars "${requiredVars[@]}"

export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"

export REPO_OWNER="kyma-project"
export REPO_NAME="kyma"

COMMON_NAME=$(echo "${INPUT_CLUSTER_NAME}" | tr "[:upper:]" "[:lower:]")
export COMMON_NAME
export DNS_SUBDOMAIN="${COMMON_NAME}"
export CLUSTER_NAME="${COMMON_NAME}"

gcp::set_vars_for_network \
  -n "$JOB_NAME"
export GCLOUD_NETWORK_NAME="${gcp_set_vars_for_network_return_net_name:?}"
export GCLOUD_SUBNET_NAME="${gcp_set_vars_for_network_return_subnet_name:?}"

# Enable Stackdriver Kubernetes Engine Monitoring support on k8s cluster. Mandatory requirement for stackdriver-prometheus collector.
# https://cloud.google.com/monitoring/kubernetes-engine/prometheus
export STACKDRIVER_KUBERNETES="true"
export SIDECAR_IMAGE_TAG="${STACKDRIVER_COLLECTOR_SIDECAR_IMAGE_TAG}"

#Enable SSD disks for k8s cluster
if [ "${CLUSTER_USE_SSD}" ]; then
	CLUSTER_USE_SSD=$(echo "${CLUSTER_USE_SSD}" | tr '[:upper:]' '[:lower:]')
	if [ "${CLUSTER_USE_SSD}" == "true" ] || [ "${CLUSTER_USE_SSD}" == "yes" ]; then
		export CLUSTER_USE_SSD
	else
		echo "CLUSTER_USE_SSD prowjob env variable allowed values are true or yes. Cluster will be build with standard disks."
		unset CLUSTER_USE_SSD
	fi
fi

# Provision GKE regional cluster.
if [ "${PROVISION_REGIONAL_CLUSTER}" ]; then
	PROVISION_REGIONAL_CLUSTER=$(echo "${PROVISION_REGIONAL_CLUSTER}" | tr '[:upper:]' '[:lower:]')
	if [ "${PROVISION_REGIONAL_CLUSTER}" == "true" ] || [ "${PROVISION_REGIONAL_CLUSTER}" == "yes" ]; then
		export PROVISION_REGIONAL_CLUSTER
		export CLOUDSDK_COMPUTE_REGION
	else
		echo "PROVISION_REGIONAL_CLUSTER prowjob env variable allowed values are true or yes. Provisioning standard cluster."
		unset PROVISION_REGIONAL_CLUSTER
	fi
fi

# TEST_RESULT_WINDOW_TIME=${TEST_RESULT_WINDOW_TIME:-3h}
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"

function createCluster() {
	log::info "Reserve IP Address for Ingressgateway"
	GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
	export GATEWAY_IP_ADDRESS
	gcp::reserve_ip_address \
		-n "${GATEWAY_IP_ADDRESS_NAME}" \
		-p "$CLOUDSDK_CORE_PROJECT" \
		-r "$CLOUDSDK_COMPUTE_REGION"
	GATEWAY_IP_ADDRESS="${gcp_reserve_ip_address_return_ip_address:?}"
	echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

	log::info "Create DNS Record for Ingressgateway IP"
	gcp::create_dns_record \
		-a "$GATEWAY_IP_ADDRESS" \
		-h "*" \
		-s "$COMMON_NAME" \
		-p "$CLOUDSDK_CORE_PROJECT" \
		-z "$CLOUDSDK_DNS_ZONE_NAME"

	log::info "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
	gcp::create_network \
    -n "${GCLOUD_NETWORK_NAME}" \
	-s "${GCLOUD_SUBNET_NAME}" \
	-p "$CLOUDSDK_CORE_PROJECT"

	log::info "Provision cluster: \"${COMMON_NAME}\""
	date
	
	gcp::provision_k8s_cluster \
		-c "$COMMON_NAME" \
		-p "$CLOUDSDK_CORE_PROJECT" \
		-v "$GKE_CLUSTER_VERSION" \
		-j "$JOB_NAME" \
		-J "$PROW_JOB_ID" \
		-z "$CLOUDSDK_COMPUTE_ZONE" \
		-m "$MACHINE_TYPE" \
		-R "$CLOUDSDK_COMPUTE_REGION" \
		-N "$GCLOUD_NETWORK_NAME" \
		-S "$GCLOUD_SUBNET_NAME" \
		-r "$PROVISION_REGIONAL_CLUSTER" \
		-s "$STACKDRIVER_KUBERNETES" \
		-D "$CLUSTER_USE_SSD" \
		-P "$TEST_INFRA_SOURCES_DIR"
}

function installKyma() {

	kymaUnsetVar=false

  # shellcheck disable=SC2043
	for var in GATEWAY_IP_ADDRESS ; do
    	if [ -z "${!var}" ] ; then
        	echo "ERROR: $var is not set"
        	kymaUnsetVar=true
    	fi
	done
	if [ "${kymaUnsetVar}" = true ] ; then
    	exit 1
	fi

	CERTIFICATES_BUCKET="${CERTIFICATES_BUCKET}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-letsencrypt-cert.sh"
	TLS_CERT=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/fullchain.pem | tr -d '\n')
	export TLS_CERT
	TLS_KEY=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/privkey.pem   | tr -d '\n')
	export TLS_KEY

	log::info "Trigger installation"
	set -x

	kyma deploy \
			--ci \
			--source local \
			--workspace "${KYMA_SOURCES_DIR}" \
			--domain "${DOMAIN}" \
			--profile production \
			--tls-crt "./letsencrypt/live/${DOMAIN}/fullchain.pem" \
			--tls-key "./letsencrypt/live/${DOMAIN}/privkey.pem" \
			--value "istio.components.ingressGateways.config.service.loadBalancerIP=${GATEWAY_IP_ADDRESS}" \
			--value "global.domainName=${DOMAIN}"

	set +x

}

function installStackdriverPrometheusCollector(){
  # Patching prometheus resource of prometheus-operator.
  # Injecting stackdriver-collector sidecar to export metrics in to stackdriver monitoring.
  # Adding additional scrape config to get stackdriver sidecar metrics.
  echo "Create additional scrape config secret."
  kubectl -n kyma-system create secret generic prometheus-operator-additional-scrape-config --from-file="${TEST_INFRA_SOURCES_DIR}"/prow/scripts/resources/prometheus-operator-additional-scrape-config.yaml
	echo "Replace tags with current values in patch yaml file."
	sed -i.bak -e 's/__SIDECAR_IMAGE_TAG__/'"${SIDECAR_IMAGE_TAG}"'/g' \
		-e 's/__GCP_PROJECT__/'"${CLOUDSDK_CORE_PROJECT}"'/g' \
		-e 's/__GCP_REGION__/'"${CLOUDSDK_COMPUTE_REGION}"'/g' \
		-e 's/__CLUSTER_NAME__/'"${COMMON_NAME}"'/g' "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/resources/prometheus-operator-stackdriver-patch.yaml
	echo "Patch monitoring prometheus CRD to deploy stackdriver-prometheus collector as sidecar"
	kubectl -n kyma-system patch prometheus monitoring-prometheus --type merge --patch "$(cat "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/resources/prometheus-operator-stackdriver-patch.yaml)"
}

log::info "Authenticate"
gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"
docker::start


DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"
export DNS_DOMAIN
DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
export DOMAIN

log::info "Cleanup"
export SKIP_IMAGE_REMOVAL=true
export ASYNC_DEPROVISION=false
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/cleanup-cluster.sh"

log::info "Create new cluster"
createCluster

kyma::install_cli

log::info "Install kyma"
installKyma

#log::info "Install stackdriver-prometheus collector"
#installStackdriverPrometheusCollector

#log::info "Update stackdriver-metadata-agent memory settings"

#cat <<EOF | kubectl replace -f -
#apiVersion: v1
#data:
#  NannyConfiguration: |-
#    apiVersion: nannyconfig/v1alpha1
#    kind: NannyConfiguration
#    baseMemory: 100Mi
#kind: ConfigMap
#metadata:
#  labels:
#    addonmanager.kubernetes.io/mode: EnsureExists
#    kubernetes.io/cluster-service: "true"
#  name: metadata-agent-config
#  namespace: kube-system
#EOF
#kubectl delete deployment -n kube-system stackdriver-metadata-agent-cluster-level


log::info "Collect list of images"
if [ -z "$ARTIFACTS" ] ; then
    ARTIFACTS=/tmp/artifacts
fi

IMAGES_LIST=$(kubectl get pods --all-namespaces -o go-template --template='{{range .items}}{{range .status.containerStatuses}}{{.name}},{{.image}},{{.imageID}}{{printf "\n"}}{{end}}{{range .status.initContainerStatuses}}{{.name}},{{.image}},{{.imageID}}{{printf "\n"}}{{end}}{{end}}' | uniq | sort)
echo "${IMAGES_LIST}" > "${ARTIFACTS}/kyma-images-${COMMON_NAME}.csv"

# also generate image list in json
## this is false-positive as we need to use single-quotes for jq
# shellcheck disable=SC2016
IMAGES_LIST=$(kubectl get pods --all-namespaces -o json | jq '{ images: [.items[] | .metadata.ownerReferences[0].name as $owner | (.status.containerStatuses + .status.initContainerStatuses)[] | { name: .imageID, custom_fields: {owner: $owner, image: .image, name: .name }}] | unique | group_by(.name) | map({name: .[0].name, custom_fields: {owner: map(.custom_fields.owner) | unique | join(","), container_name: map(.custom_fields.name) | unique | join(","), image: .[0].custom_fields.image}}) | map(select (.name | startswith("sha256") | not))}' )
echo "${IMAGES_LIST}" > "${ARTIFACTS}/kyma-images-${COMMON_NAME}.json"

# generate pod-security-policy list in json
utils::save_psp_list "${ARTIFACTS}/kyma-psp.json"

utils::kubeaudit_create_report "${ARTIFACTS}/kubeaudit.log"

log::success "Success"
