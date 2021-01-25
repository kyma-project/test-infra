#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"

# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

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
	SLACK_CLIENT_TOKEN
	SLACK_CLIENT_WEBHOOK_URL
	STABILITY_SLACK_CLIENT_CHANNEL_ID
	STACKDRIVER_COLLECTOR_SIDECAR_IMAGE_TAG
	CERTIFICATES_BUCKET
)

utils::check_required_vars "${requiredVars[@]}"

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
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/gcloud.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcloud.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"

function createCluster() {
	log::info "Reserve IP Address for Ingressgateway"
	GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"
	GATEWAY_IP_ADDRESS=$(gcloud::reserve_ip_address "${GATEWAY_IP_ADDRESS_NAME}")
	echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

	log::info "Create DNS Record for Ingressgateway IP"
	GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
	gcloud::create_dns_record "${GATEWAY_IP_ADDRESS}" "${GATEWAY_DNS_FULL_NAME}"

	log::info "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
	gcloud::create_network "${GCLOUD_NETWORK_NAME}" "${GCLOUD_SUBNET_NAME}"

	log::info "Provision cluster: \"${CLUSTER_NAME}\""
	date
	
	if [ -z "${CLUSTER_VERSION}" ]; then
		export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
	fi
	gcloud::provision_gke_cluster "$CLUSTER_NAME"
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

	log::info "Prepare Kyma overrides"

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
  name: "cluster-essentials-overrides"
  namespace: "kyma-installer"
  labels:
    installer: overrides
    kyma-project.io/installation: ""
    component: cluster-essentials
data:
  limitRange.max.memory: "8Gi"
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
  prometheus-istio.envoyStats.sampleLimit: "8000"
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

	log::info "Trigger installation"

	KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

  # WORKAROUND: add gateway ip address do IstioOperator in installer-config-production.yaml.tpl (see: https://github.com/kyma-project/test-infra/issues/2792)
  echo "#### WORKAROUND: add gateway ip address do IstioOperator in installer-config-production.yaml.tpl (see: https://github.com/kyma-project/test-infra/issues/2792)"
  curl -sSLo /usr/local/bin/yq "https://github.com/mikefarah/yq/releases/download/3.4.1/yq_linux_amd64" && chmod +x /usr/local/bin/yq
cat << EOF > istio-ingressgateway-patch-yq.yaml
- command: update
  path: spec.components.ingressGateways[0].k8s.service
  value:
    loadBalancerIP: ${GATEWAY_IP_ADDRESS}
    type: LoadBalancer
EOF
  yq w -i -d1 "${KYMA_RESOURCES_DIR}"/installer-config-production.yaml.tpl 'data.kyma_istio_operator' "$(yq r -d1 "${KYMA_RESOURCES_DIR}"/installer-config-production.yaml.tpl 'data.kyma_istio_operator' | yq w - -s istio-ingressgateway-patch-yq.yaml)"

  # Update the memory override for prometheus-istio."${KYMA_RESOURCES_DIR}"
  sed -i 's/prometheus-istio.server.resources.limits.memory: "4Gi"/prometheus-istio.server.resources.limits.memory: "8Gi"/g' "${KYMA_RESOURCES_DIR}"/installer-config-production.yaml.tpl

	kyma install \
			--ci \
			--source master \
			-o "${KYMA_RESOURCES_DIR}"/installer-config-production.yaml.tpl \
			-o "${componentOverridesFile}" \
			--domain "${DOMAIN}" \
			--tls-cert "${TLS_CERT}" \
			--tls-key "${TLS_KEY}" \
			--timeout 60m

	if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
		log::info "Create DNS Record for Apiserver proxy IP"
		APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
		APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
		gcloud::create_dns_record "${APISERVER_IP_ADDRESS}" "${APISERVER_DNS_FULL_NAME}"
	fi
}

function applyServiceCatalogCRDOverride(){
    log::info "Apply override for ServiceCatalog to enable CRD implementation"

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

function apply_dex_github_kyma_admin_group() {
    kubectl get ClusterRoleBinding kyma-admin-binding -oyaml > kyma-admin-binding.yaml && cat >> kyma-admin-binding.yaml <<EOF 
- apiGroup: rbac.authorization.k8s.io
  kind: Group
  name: kyma-project:cluster-access
EOF

    kubectl replace -f kyma-admin-binding.yaml
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

log::info "Authenticate"
gcloud::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"
docker::start


DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"
export DNS_DOMAIN
DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
export DOMAIN

log::info "Cleanup"
export SKIP_IMAGE_REMOVAL=true
export DISABLE_ASYNC_DEPROVISION=true
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/cleanup-cluster.sh"

log::info "Create new cluster"
createCluster

log::info "install image-guard"
helm install image-guard "$TEST_INFRA_SOURCES_DIR/development/image-guard/image-guard"

kyma::install_cli

log::info "Install kyma"
installKyma

log::info "Override kyma-admin-binding ClusterRoleBinding"
apply_dex_github_kyma_admin_group

log::info "Install stackdriver-prometheus collector"
installStackdriverPrometheusCollector

log::info "Update stackdriver-metadata-agent memory settings"

cat <<EOF | kubectl replace -f -
apiVersion: v1
data:
  NannyConfiguration: |-
    apiVersion: nannyconfig/v1alpha1
    kind: NannyConfiguration
    baseMemory: 100Mi
kind: ConfigMap
metadata:
  labels:
    addonmanager.kubernetes.io/mode: EnsureExists
    kubernetes.io/cluster-service: "true"
  name: metadata-agent-config
  namespace: kube-system
EOF
kubectl delete deployment -n kube-system stackdriver-metadata-agent-cluster-level


log::info "Collect list of images"
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

log::info "Gather Kubeaudit logs"
curl -sL https://github.com/Shopify/kubeaudit/releases/download/v0.11.8/kubeaudit_0.11.8_linux_amd64.tar.gz | tar -xzO kubeaudit > ./kubeaudit
chmod +x ./kubeaudit
# kubeaudit returns non-zero exit code when it finds issues
# In the context of this job we just want to grab the logs
# It should not break the execution of this script
./kubeaudit privileged privesc -p json  > "${ARTIFACTS}/kubeaudit.log" || true

log::info "Install stability-checker"
date
(
export TEST_INFRA_SOURCES_DIR KYMA_SCRIPTS_DIR TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS \
        CLUSTER_NAME SLACK_CLIENT_WEBHOOK_URL STABILITY_SLACK_CLIENT_CHANNEL_ID SLACK_CLIENT_TOKEN TEST_RESULT_WINDOW_TIME
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/install-stability-checker.sh"
)

log::success "Success"
