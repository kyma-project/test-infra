#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

discoverUnsetVar=false

for var in INPUT_CLUSTER_NAME DOCKER_PUSH_REPOSITORY DOCKER_PUSH_DIRECTORY KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_COMPUTE_ZONE CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS SLACK_CLIENT_TOKEN SLACK_CLIENT_WEBHOOK_URL STABILITY_SLACK_CLIENT_CHANNEL_ID STACKDRIVER_COLLECTOR_SIDECAR_IMAGE_TAG CERTIFICATES_BUCKET; do
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

export REPO_OWNER="kyma-project"
export REPO_NAME="kyma"

STANDARIZED_NAME=$(echo "${INPUT_CLUSTER_NAME}" | tr "[:upper:]" "[:lower:]")
export STANDARIZED_NAME
export DNS_SUBDOMAIN="${STANDARIZED_NAME}"

export CLUSTER_NAME="${STANDARIZED_NAME}"
export GCLOUD_NETWORK_NAME="gke-long-lasting-net"
export GCLOUD_SUBNET_NAME="gke-long-lasting-subnet"

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

if [ -z "${SERVICE_CATALOG_CRD}" ]; then
	export SERVICE_CATALOG_CRD="false"
fi

TEST_RESULT_WINDOW_TIME=${TEST_RESULT_WINDOW_TIME:-3h}
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/kyma-cli.sh"

function createCluster() {
	shout "Reserve IP Address for Ingressgateway"
	date
	GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"
	GATEWAY_IP_ADDRESS=$(IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/reserve-ip-address.sh)
	echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

	shout "Create DNS Record for Ingressgateway IP"
	date
	GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
	IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-dns-record.sh

	NETWORK_EXISTS=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/network-exists.sh")
	if [ "$NETWORK_EXISTS" -gt 0 ]; then
		shout "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
		date
		"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-network-with-subnet.sh"
	else
		shout "Network ${GCLOUD_NETWORK_NAME} exists"
	fi

	shout "Provision cluster: \"${CLUSTER_NAME}\""
	date
	
	if [ -z "${CLUSTER_VERSION}" ]; then
		export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
	fi
	env ADDITIONAL_LABELS="created-at=${CURRENT_TIMESTAMP}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/provision-gke-cluster.sh
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

	shout "Prepare Kyma overrides"
	date

	export DEX_CALLBACK_URL="https://dex.${DOMAIN}/callback"

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
  test.acceptance.ui.logging.enabled: "true"
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
#---
#apiVersion: v1
#kind: ConfigMap
#metadata:
#  name: "istio-overrides"
#  namespace: "kyma-installer"
#  labels:
#    installer: overrides
#    kyma-project.io/installation: ""
#    component: istio
#data:
#  gateways.istio-ingressgateway.loadBalancerIP: "${GATEWAY_IP_ADDRESS}"
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
      clientID: ${GITHUB_INTEGRATION_APP_CLIENT_ID}
      clientSecret: ${GITHUB_INTEGRATION_APP_CLIENT_SECRET}
      redirectURI: ${DEX_CALLBACK_URL}
      orgs:
      - name: kyma-project
EOF
)
  echo "${componentOverrides}" > "${componentOverridesFile}"

	if [ "${SERVICE_CATALOG_CRD}" = "true" ]; then
			applyServiceCatalogCRDOverride
	fi

	shout "Trigger installation"
	date

	KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

  # WORKAROUND: add gateway ip address do IstioOperator in installer-config-production.yaml.tpl (see: https://github.com/kyma-project/test-infra/issues/2792)
  echo "#### WORKAROUND: add gateway ip address do IstioOperator in installer-config-production.yaml.tpl (see: https://github.com/kyma-project/test-infra/issues/2792)"
  yq_rel_latest=$(curl --silent "https://api.github.com/repos/mikefarah/yq/releases/latest" | grep -Po '"tag_name": "\K.*?(?=")')
  curl -sSLo /usr/local/bin/yq "https://github.com/mikefarah/yq/releases/download/${yq_rel_latest}/yq_linux_amd64" && chmod +x /usr/local/bin/yq
cat << EOF > istio-ingressgateway-patch-yq.yaml
- command: update
  path: spec.components.ingressGateways[0].k8s.service
  value:
    loadBalancerIP: ${GATEWAY_IP_ADDRESS}
    type: LoadBalancer
EOF
  yq w -i -d1 "${KYMA_RESOURCES_DIR}"/installer-config-production.yaml.tpl 'data.kyma_istio_operator' "$(yq r -d1 "${KYMA_RESOURCES_DIR}"/installer-config-production.yaml.tpl 'data.kyma_istio_operator' | yq w - -s istio-ingressgateway-patch-yq.yaml)"

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
		APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
		IP_ADDRESS=${APISERVER_IP_ADDRESS} DNS_FULL_NAME=${APISERVER_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
	fi
}

function applyServiceCatalogCRDOverride(){
    shout "Apply override for ServiceCatalog to enable CRD implementation"

serviceCatalogOverrides=$(cat << EOF
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: service-catalog-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: service-catalog
    kyma-project.io/installation: ""
data:
  service-catalog-apiserver.enabled: "false"
  service-catalog-crds.enabled: "true"
EOF
)

echo "${serviceCatalogOverrides}" >> "${componentOverridesFile}"
}

function installStackdriverPrometheusCollector(){
  # Patching prometheus resource of prometheus-operator.
  # Injecting stackdriver-collector sidecar to export metrics in to stackdriver monitoring.
  # Adding additional scrape config to get stackdriver sidecar metrics.
  echo "Create additional scrape config secret."
  kubectl -n kyma-system create secret generic prometheus-operator-additional-scrape-config --from-file="${TEST_INFRA_SOURCES_DIR}"/prow/scripts/resources/prometheus-operator-additional-scrape-config.yaml
	echo "Replace tags with current values in patch yaml file."
	sed -i.bak -e 's/__SIDECAR_IMAGE_TAG__/'"${SIDECAR_IMAGE_TAG}"'/g' \
		-e 's/__GCP_PROJECT__/'"${GCLOUD_PROJECT_NAME}"'/g' \
		-e 's/__GCP_REGION__/'"${CLOUDSDK_COMPUTE_REGION}"'/g' \
		-e 's/__CLUSTER_NAME__/'"${CLUSTER_NAME}"'/g' "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/resources/prometheus-operator-stackdriver-patch.yaml
	echo "Patch monitoring prometheus CRD to deploy stackdriver-prometheus collector as sidecar"
	kubectl -n kyma-system patch prometheus monitoring-prometheus --type merge --patch "$(cat "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/resources/prometheus-operator-stackdriver-patch.yaml)"
}

function patchlimitrange(){
  # Patching limitrange on kyma-system namespace to meet prometheus memory requirements.
	echo "Patching kyma-default LimitRange"
	kubectl -n kyma-system patch limitrange kyma-default --type merge --patch "$(cat "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/resources/limitrange-patch.yaml)"

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
export SKIP_IMAGE_REMOVAL=true
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/cleanup-cluster.sh"

shout "Create new cluster"
date
createCluster

export INSTALL_DIR=${TMP_DIR}
install::kyma_cli

shout "Install kyma"
date
installKyma

shout "Override kyma-admin-binding ClusterRoleBinding"
applyDexGithibKymaAdminGroup

# Prometheus container need minimum 6Gi memory limit.
shout "Increase cluster max container memory limit"
date
patchlimitrange

shout "Install stackdriver-prometheus collector"
date
installStackdriverPrometheusCollector

shout "Update stackdriver-metadata-agent memory settings"
updatememorysettings

shout "Collect list of images"
date
if [ -z "$ARTIFACTS" ] ; then
    ARTIFACTS:=/tmp/artifacts
fi

IMAGES_LIST=$(kubectl get pods --all-namespaces -o go-template --template='{{range .items}}{{range .status.containerStatuses}}{{.name}},{{.image}},{{.imageID}}{{printf "\n"}}{{end}}{{range .status.initContainerStatuses}}{{.name}},{{.image}},{{.imageID}}{{printf "\n"}}{{end}}{{end}}' | uniq | sort)
echo "${IMAGES_LIST}" > "${ARTIFACTS}/kyma-images-${CLUSTER_NAME}.csv"

# also generate image list in json
## this is false-positive as we need to use single-quotes for jq
# shellcheck disable=SC2016
IMAGES_LIST=$(kubectl get pods --all-namespaces -o json | jq '{ images: [.items[] | .metadata.ownerReferences[0].name as $owner | (.status.containerStatuses + .status.initContainerStatuses)[] | { name: .imageID, custom_fields: {owner: $owner, image: .image, name: .name }}] | unique | group_by(.name) | map({name: .[0].name, custom_fields: {owner: map(.custom_fields.owner) | unique | join(","), container_name: map(.custom_fields.name) | unique | join(","), image: .[0].custom_fields.image}})}' )
echo "${IMAGES_LIST}" > "${ARTIFACTS}/kyma-images-${CLUSTER_NAME}.json"

shout "Install stability-checker"
date
(
export TEST_INFRA_SOURCES_DIR KYMA_SCRIPTS_DIR TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS \
        CLUSTER_NAME SLACK_CLIENT_WEBHOOK_URL STABILITY_SLACK_CLIENT_CHANNEL_ID SLACK_CLIENT_TOKEN TEST_RESULT_WINDOW_TIME
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/install-stability-checker.sh"
)

shout "Success"
