#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

discoverUnsetVar=false
for var in REPO_OWNER REPO_NAME DOCKER_PUSH_REPOSITORY KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_COMPUTE_ZONE CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS GARDENER_KYMA_PROW_PROJECT_NAME GARDENER_KYMA_PROW_KUBECONFIG GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME; do
  if [ -z "${!var}" ] ; then
    echo "ERROR: $var is not set"
    discoverUnsetVar=true
  fi
done
if [ "${discoverUnsetVar}" = true ] ; then
  exit 1
fi

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export COMPASS_SOURCES_DIR="/home/prow/go/src/github.com/kyma-incubator/compass"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

readonly COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET="${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/compass"

# shellcheck source=prow/scripts/library.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"

export GARDENER_PROJECT_NAME="${GARDENER_KYMA_PROW_PROJECT_NAME}"
export GARDENER_APPLICATION_CREDENTIALS="${GARDENER_KYMA_PROW_KUBECONFIG}"
export GARDENER_AZURE_SECRET_NAME="${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}"

# Enforce lowercase
readonly REPO_OWNER=$(echo "${REPO_OWNER}" | tr '[:upper:]' '[:lower:]')
export REPO_OWNER
readonly REPO_NAME=$(echo "${REPO_NAME}" | tr '[:upper:]' '[:lower:]')
export REPO_NAME

readonly CURRENT_TIMESTAMP=$(date +%Y%m%d)
readonly RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c10)

if [[ "$BUILD_TYPE" == "pr" ]]; then
  # In case of PR, operate on PR number
  readonly COMMON_NAME_PREFIX="gkecompint-pr"
  COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}")
  COMPASS_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-compass-integration/${REPO_OWNER}/${REPO_NAME}:PR-${PULL_NUMBER}"
  export COMPASS_INSTALLER_IMAGE
else
  # Otherwise (master), operate on triggering commit id
  readonly COMMON_NAME_PREFIX="gkecompint-commit"
  readonly COMMIT_ID=$(cd "$COMPASS_SOURCES_DIR" && git rev-parse --short HEAD)
  COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}")
  COMPASS_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-compass-integration/${REPO_OWNER}/${REPO_NAME}:COMMIT-${COMMIT_ID}"
  export COMPASS_INSTALLER_IMAGE
fi

### Cluster name must be less than 40 characters!
COMMON_NAME=$(echo "${COMMON_NAME}" | tr "[:upper:]" "[:lower:]")
export CLUSTER_NAME="${COMMON_NAME}"

export GCLOUD_NETWORK_NAME="gke-long-lasting-net"
export GCLOUD_SUBNET_NAME="gke-long-lasting-subnet"

### For provision-gke-cluster.sh
export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"

#Local variables
DNS_SUBDOMAIN="${COMMON_NAME}"
COMPASS_SCRIPTS_DIR="${COMPASS_SOURCES_DIR}/installation/scripts"
COMPASS_RESOURCES_DIR="${COMPASS_SOURCES_DIR}/installation/resources"

INSTALLER_YAML="${COMPASS_RESOURCES_DIR}/installer.yaml"
INSTALLER_CR="${COMPASS_RESOURCES_DIR}/installer-cr.yaml.tpl"

function createCluster() {
  #Used to detect errors for logging purposes
  ERROR_LOGGING_GUARD="true"

  shout "Authenticate"
  date
  init
  DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"
  export DNS_DOMAIN
  DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
  export DOMAIN

  shout "Reserve IP Address for Ingressgateway"
  date
  GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
  GATEWAY_IP_ADDRESS=$(IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/reserve-ip-address.sh)
  CLEANUP_GATEWAY_IP_ADDRESS="true"
  echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

  shout "Create DNS Record for Ingressgateway IP"
  date
  GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
  CLEANUP_GATEWAY_DNS_RECORD="true"
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
  export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
  if [ -z "$MACHINE_TYPE" ]; then
    export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
  fi
  if [ -z "${CLUSTER_VERSION}" ]; then
    export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
  fi
  CLEANUP_CLUSTER="true"
  env ADDITIONAL_LABELS="created-at=${CURRENT_TIMESTAMP}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/provision-gke-cluster.sh
}

function applyKymaOverrides() {
  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "application-resource-tests-overrides" \
    --data "application-operator.tests.enabled=false" \
    --data "tests.application_connector_tests.enabled=false" \
    --data "application-registry.tests.enabled=false" \
    --data "console-backend-service.tests.enabled=false" \
    --data "test.acceptance.service-catalog.enabled=false" \
    --data "test.acceptance.external_solution.enabled=false" \
    --data "console.test.acceptance.enabled=false" \
    --data "test.external_solution.event_mesh.enabled=false"


  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "core-test-ui-acceptance-overrides" \
    --data "test.acceptance.ui.logging.enabled=true" \
    --label "component=core"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "application-registry-overrides" \
    --data "application-registry.deployment.args.detailedErrorResponse=true" \
    --label "component=application-connector"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "monitoring-config-overrides" \
    --data "global.alertTools.credentials.slack.channel=${KYMA_ALERTS_CHANNEL}" \
    --data "global.alertTools.credentials.slack.apiurl=${KYMA_ALERTS_SLACK_API_URL}" \
    --data "pushgateway.enabled=true" \
    --label "component=monitoring"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "istio-overrides" \
    --data "gateways.istio-ingressgateway.loadBalancerIP=${GATEWAY_IP_ADDRESS}" \
    --label "component=istio"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "api-gateway-overrides" \
    --data "tests.env.gatewayName=compass-istio-gateway" \
    --data "tests.env.gatewayNamespace=compass-system" \
    --label "component=api-gateway"

  shout "Testing custom implementation of prow jobs!"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "dex-overrides" \
    --data "global.istio.gateway.name=kyma-gateway" \
    --data "global.istio.gateway.namespace=kyma-system" \
    --label "component=dex"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "ory-overrides" \
    --data "global.istio.gateway.name=kyma-gateway" \
    --data "global.istio.gateway.namespace=kyma-system" \
    --label "component=ory"
}

function applyCompassOverrides() {
  NAMESPACE="compass-installer"

  if [ "${RUN_PROVISIONER_TESTS}" == "true" ]; then
    # Change timeout for kyma test to 3h
    export KYMA_TEST_TIMEOUT=3h

    # Create Config map for Provisioner Tests
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "provisioner-tests-overrides" \
      --data "global.provisioning.enabled=true" \
      --data "provisioner.security.skipTLSCertificateVeryfication=true" \
      --data "provisioner.tests.enabled=true" \
      --data "provisioner.gardener.kubeconfig=$(base64 -w 0 < "${GARDENER_APPLICATION_CREDENTIALS}")" \
      --data "provisioner.gardener.project=$GARDENER_PROJECT_NAME" \
      --data "provisioner.tests.gardener.azureSecret=$GARDENER_AZURE_SECRET_NAME" \
      --label "component=compass"
  fi

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "compass-auditlog-mock-tests" \
    --data "global.externalServicesMock.enabled=true" \
    --data "gateway.gateway.auditlog.enabled=true" \
    --data "gateway.gateway.auditlog.authMode=oauth" \
    --label "component=compass"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "compass-gateway-overrides" \
    --data "global.agentPreconfiguration=true" \
    --data "global.istio.gateway.name=kyma-gateway" \
    --data "global.istio.gateway.namespace=kyma-system" \
    --data "global.connector.secrets.ca.name=connector-service-app-ca" \
    --data "global.connector.secrets.ca.namespace=kyma-integration" \
    --data "gateway.gateway.enabled=false" \
    --label "component=compass"
}

function applyCommonOverrides() {
  NAMESPACE=${1}

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "installation-config-overrides" \
    --data "global.domainName=${DOMAIN}" \
    --data "global.loadBalancerIP=${GATEWAY_IP_ADDRESS}"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "feature-flags-overrides" \
    --data "global.enableAPIPackages=true" \
    --data "global.disableLegacyConnectivity=true"


  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "cluster-certificate-overrides" \
    --data "global.tlsCrt=${TLS_CERT}" \
    --data "global.tlsKey=${TLS_KEY}"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "global-ingress-overrides" \
    --data "global.ingress.domainName=${DOMAIN}" \
    --data "global.ingress.tlsCrt=${TLS_CERT}" \
    --data "global.ingress.tlsKey=${TLS_KEY}" \
    --data "global.environment.gardener=false"
}

function installKyma() {
  kubectl create namespace "kyma-installer"
  applyCommonOverrides "kyma-installer"
  applyKymaOverrides

  if [[ "$BUILD_TYPE" == "pr" ]]; then
#    COMPASS_VERSION="PR-${PULL_NUMBER}"
    COMPASS_VERSION="PR-1596"
  else
    COMPASS_VERSION="master-${COMMIT_ID}"
  fi
  readonly COMPASS_ARTIFACTS="${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/${COMPASS_VERSION}"
  
  readonly TMP_DIR="/tmp/compass-gke-integration"

  gsutil cp "${COMPASS_ARTIFACTS}/kyma-installer.yaml" ${TMP_DIR}/kyma-installer.yaml
  gsutil cp "${COMPASS_ARTIFACTS}/is-kyma-installed.sh" ${TMP_DIR}/is-kyma-installed.sh
  chmod +x ${TMP_DIR}/is-kyma-installed.sh
  kubectl apply -f ${TMP_DIR}/kyma-installer.yaml

  shout "Installation triggered"
  date
  "${TMP_DIR}"/is-kyma-installed.sh --timeout 30m
}

function installCompass() {
  compassUnsetVar=false

  # shellcheck disable=SC2043
  for var in GATEWAY_IP_ADDRESS ; do
    if [ -z "${!var}" ] ; then
      echo "ERROR: $var is not set"
      compassUnsetVar=true
    fi
  done
  if [ "${compassUnsetVar}" = true ] ; then
    exit 1
  fi

  shout "Build Compass-Installer Docker image"
  date
  CLEANUP_DOCKER_IMAGE="true"
  COMPASS_INSTALLER_IMAGE="${COMPASS_INSTALLER_IMAGE}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-compass-image.sh

  shout "Apply Compass config"
  date
  kubectl create namespace "compass-installer"
  applyCommonOverrides "compass-installer"
  applyCompassOverrides

  echo "Manual concatenating yamls"
  "${COMPASS_SCRIPTS_DIR}"/concat-yamls.sh "${INSTALLER_YAML}" "${INSTALLER_CR}" \
  | sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${COMPASS_INSTALLER_IMAGE};" \
  | sed -e "s/__VERSION__/0.0.1/g" \
  | sed -e "s/__.*__//g" \
  | kubectl apply -f-
  
  shout "Installation triggered"
  date
  "${COMPASS_SCRIPTS_DIR}"/is-installed.sh --timeout 30m

  if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
    shout "Create DNS Record for Apiserver proxy IP"
    date
    APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
    CLEANUP_APISERVER_DNS_RECORD="true"
    IP_ADDRESS=${APISERVER_IP_ADDRESS} DNS_FULL_NAME=${APISERVER_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
  fi
}

trap gkeCleanup EXIT INT

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    shout "Execute Job Guard"
    export JOB_NAME_PATTERN="(pre-compass-components-.*)|(pre-compass-tests-.*)"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

shout "Create new cluster"
date
createCluster

shout "Generate self-signed certificate"
date
CERT_KEY=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-self-signed-cert.sh")
TLS_CERT=$(echo "${CERT_KEY}" | head -1)
TLS_KEY=$(echo "${CERT_KEY}" | tail -1)

shout "Install Kyma"
date
installKyma

shout "Install Compass"
date
installCompass

shout "Test Kyma with Compass"
date
"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh

shout "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
