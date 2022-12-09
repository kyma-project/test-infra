#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

ENABLE_TEST_LOG_COLLECTOR=false

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KCP_SOURCES_DIR="/home/prow/go/src/github.com/kyma-project/control-plane"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_MAJOR_VERSION="1"

# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"

requiredVars=(
    REPO_OWNER
    REPO_NAME
    DOCKER_PUSH_REPOSITORY
    KYMA_PROJECT_DIR
    CLOUDSDK_CORE_PROJECT
    CLOUDSDK_COMPUTE_REGION
    CLOUDSDK_COMPUTE_ZONE
    CLOUDSDK_DNS_ZONE_NAME
    GOOGLE_APPLICATION_CREDENTIALS
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
)

utils::check_required_vars "${requiredVars[@]}"

export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"

export GARDENER_PROJECT_NAME="${GARDENER_KYMA_PROW_PROJECT_NAME}"
export GARDENER_APPLICATION_CREDENTIALS="${GARDENER_KYMA_PROW_KUBECONFIG}"
export GARDENER_AZURE_SECRET_NAME="${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}"

# Enforce lowercase
readonly REPO_OWNER=$(echo "${REPO_OWNER}" | tr '[:upper:]' '[:lower:]')
export REPO_OWNER
readonly REPO_NAME=$(echo "${REPO_NAME}" | tr '[:upper:]' '[:lower:]')
export REPO_NAME

readonly RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c10)

if [[ "$BUILD_TYPE" == "pr" ]]; then
  # In case of PR, operate on PR number
  readonly COMMON_NAME_PREFIX="gkekcpint-pr"
  COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}")
  KCP_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-control-plane-integration/${REPO_OWNER}/${REPO_NAME}:PR-${PULL_NUMBER}"
  export KCP_INSTALLER_IMAGE
else
  # Otherwise (master), operate on triggering commit id
  readonly COMMON_NAME_PREFIX="gkekcpint-commit"
  readonly COMMIT_ID=$(cd "$KCP_SOURCES_DIR" && git rev-parse --short HEAD)
  COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}")
  KCP_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-control-plane-integration/${REPO_OWNER}/${REPO_NAME}:COMMIT-${COMMIT_ID}"
  export KCP_INSTALLER_IMAGE
fi

readonly KCP_DEVELOPMENT_ARTIFACTS_BUCKET="${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/kcp"
if [[ "$BUILD_TYPE" == "pr" ]]; then
  KCP_VERSION="PR-${PULL_NUMBER}"
else
  KCP_VERSION="master-${COMMIT_ID}"
fi
readonly KCP_ARTIFACTS="${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${KCP_VERSION}"

COMMON_NAME=$(echo "${COMMON_NAME}" | tr "[:upper:]" "[:lower:]")

gcp::set_vars_for_network \
  -n "$JOB_NAME"
export GCLOUD_NETWORK_NAME="${gcp_set_vars_for_network_return_net_name:?}"
export GCLOUD_SUBNET_NAME="${gcp_set_vars_for_network_return_subnet_name:?}"

#Local variables
DNS_SUBDOMAIN="${COMMON_NAME}"
KCP_SCRIPTS_DIR="${KCP_SOURCES_DIR}/installation/scripts"
KCP_RESOURCES_DIR="${KCP_SOURCES_DIR}/installation/resources"

INSTALLER_YAML="${KCP_RESOURCES_DIR}/installer.yaml"
INSTALLER_CR="${KCP_RESOURCES_DIR}/installer-cr.yaml.tpl"

# post_hook runs at the end of a script or on any error
function docker_cleanup() {
  #Turn off exit-on-error so that next step is executed even if previous one fails.
  set +e

  if [ -n "${CLEANUP_DOCKER_IMAGE}" ]; then
    log::info "Docker image cleanup"
    if [ -n "${KCP_INSTALLER_IMAGE}" ]; then
      log::info "Delete temporary KCP-Installer Docker image"
      gcp::delete_docker_image \
        -i "${KCP_INSTALLER_IMAGE}"
    fi
  fi

  set -e
}

function createCluster() {
  #Used to detect errors for logging purposes
  ERROR_LOGGING_GUARD="true"

  log::info "Authenticate"
  gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"
  docker::start
  DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"
  export DNS_DOMAIN
  DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
  export DOMAIN

  log::info "Reserve IP Address for Ingressgateway"
  GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
  gcp::reserve_ip_address \
    -n "${GATEWAY_IP_ADDRESS_NAME}" \
    -p "$CLOUDSDK_CORE_PROJECT" \
    -r "$CLOUDSDK_COMPUTE_REGION"
  GATEWAY_IP_ADDRESS="${gcp_reserve_ip_address_return_ip_address:?}"
  CLEANUP_GATEWAY_IP_ADDRESS="true"
  echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

  log::info "Create DNS Record for Ingressgateway IP"
  gcp::create_dns_record \
      -a "$GATEWAY_IP_ADDRESS" \
      -h "*" \
      -s "$COMMON_NAME" \
      -p "$CLOUDSDK_CORE_PROJECT" \
      -z "$CLOUDSDK_DNS_ZONE_NAME"
  CLEANUP_GATEWAY_DNS_RECORD="true"

  log::info "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
  gcp::create_network \
    -n "${GCLOUD_NETWORK_NAME}" \
    -s "${GCLOUD_SUBNET_NAME}" \
    -p "$CLOUDSDK_CORE_PROJECT"

  log::info "Provision cluster: \"${COMMON_NAME}\""
  export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
  gcp::provision_k8s_cluster \
        -c "$COMMON_NAME" \
        -r "$PROVISION_REGIONAL_CLUSTER" \
        -m "$MACHINE_TYPE" \
        -n "$NODES_PER_ZONE" \
        -p "$CLOUDSDK_CORE_PROJECT" \
        -v "$GKE_CLUSTER_VERSION" \
        -j "$JOB_NAME" \
        -J "$PROW_JOB_ID" \
        -z "$CLOUDSDK_COMPUTE_ZONE" \
        -R "$CLOUDSDK_COMPUTE_REGION" \
        -N "$GCLOUD_NETWORK_NAME" \
        -S "$GCLOUD_SUBNET_NAME" \
        -P "$TEST_INFRA_SOURCES_DIR"
  CLEANUP_CLUSTER="true"
}

function applyKymaOverrides() {
  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "application-resource-tests-overrides" \
    --data "application-operator.tests.enabled=false" \
    --data "tests.application_connector_tests.enabled=false" \
    --data "test.acceptance.external_solution.enabled=false" \
    --data "console.test.acceptance.enabled=false" \
    --data "test.external_solution.event_mesh.enabled=false"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "core-test-ui-acceptance-overrides" \
    --data "test.acceptance.ui.logging.enabled=true" \
    --label "component=core"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "monitoring-config-overrides" \
    --data "global.alertTools.credentials.slack.channel=${KYMA_ALERTS_CHANNEL}" \
    --data "global.alertTools.credentials.slack.apiurl=${KYMA_ALERTS_SLACK_API_URL}" \
    --data "pushgateway.enabled=true" \
    --label "component=monitoring"

  cat << EOF > "$PWD/kyma_istio_operator"
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  namespace: istio-system
spec:
  components:
    ingressGateways:
      - name: istio-ingressgateway
        k8s:
          service:
            loadBalancerIP: ${GATEWAY_IP_ADDRESS}
            type: LoadBalancer
EOF

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map-file.sh" --name "istio-overrides" \
    --label "component=istio" \
    --file "$PWD/kyma_istio_operator"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "ory-overrides" \
    --data "global.istio.gateway.name=kyma-gateway" \
    --data "global.istio.gateway.namespace=kyma-system" \
    --label "component=ory"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "tracing-overrides" \
      --data "global.tracing.enabled=true" \
      --label "component=tracing"
}

function applyCompassOverrides() {
  NAMESPACE="compass-installer"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "compass-auditlog-mock-tests" \
    --data "global.externalServicesMock.enabled=true" \
    --data "gateway.gateway.auditlog.enabled=true" \
    --data "gateway.gateway.auditlog.authMode=oauth" \
    --data "global.externalServicesMock.auditlog=true" \
    --label "component=compass"
}

function applyKebResources() {

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: kcp-system
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kcp-auditlog-script
  namespace: kcp-system
data:
  script: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: kcp-auditlog-secret
  namespace: kcp-system
type: Opaque
data:
  auditlog-user: "dXNyCg=="
  auditlog-password: "dXNyCg=="
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kcp-auditlog-config
  namespace: kcp-system
data:
  auditlog-url: "http://dummy.url"
  auditlog-url-basic: "http://dummy.url"
  auditlog-config-path: "/path"
  auditlog-security-path: "/path"
  auditlog-tenant: "tnt"
EOF
}

function applyControlPlaneOverrides() {
  NAMESPACE="kcp-installer"

  #Create Config map for Provisioner
  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "provisioner-overrides" \
    --data "provisioner.security.skipTLSCertificateVeryfication=true" \
    --data "provisioner.gardener.kubeconfig=$(base64 -w 0 < "${GARDENER_APPLICATION_CREDENTIALS}")" \
    --data "provisioner.gardener.project=$GARDENER_PROJECT_NAME" \
    --label "component=kcp"

   #Create Provisioning/KEB overrides
  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "provisioning-enable-overrides" \
    --data "global.provisioning.enabled=true" \
    --data "global.kyma_metrics_collector.enabled=true" \
    --data "global.database.embedded.enabled=true" \
    --data "global.kyma_environment_broker.enabled=true" \
    --data "kyma-environment-broker.e2e.enabled=false" \
    --data "kyma-environment-broker.disableProcessOperationsInProgress=true" \
    --label "component=kcp"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "compass-auditlog-mock-tests" \
    --data "global.externalServicesMock.enabled=true" \
    --data "gateway.gateway.auditlog.enabled=true" \
    --data "gateway.gateway.auditlog.authMode=oauth" \
    --label "component=compass"

  applyKebResources
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

  TMP_DIR="/tmp/kcp-gke-integration"

  gsutil cp "${KCP_ARTIFACTS}/kyma-installer.yaml" ${TMP_DIR}/kyma-installer.yaml
  gsutil cp "${KCP_ARTIFACTS}/is-kyma-installed.sh" ${TMP_DIR}/is-kyma-installed.sh
  chmod +x ${TMP_DIR}/is-kyma-installed.sh
  kubectl apply -f ${TMP_DIR}/kyma-installer.yaml || true
  sleep 2
  kubectl apply -f ${TMP_DIR}/kyma-installer.yaml

  log::info "Installation triggered"
  "${TMP_DIR}"/is-kyma-installed.sh --timeout 30m
}

function installCompass() {
  kubectl create namespace "compass-installer"
  applyCommonOverrides "compass-installer"
  applyCompassOverrides

  TMP_DIR="/tmp/kcp-gke-integration"

  gsutil cp "${KCP_ARTIFACTS}/compass-installer.yaml" ${TMP_DIR}/compass-installer.yaml
  gsutil cp "${KCP_ARTIFACTS}/is-compass-installed.sh" ${TMP_DIR}/is-compass-installed.sh
  chmod +x ${TMP_DIR}/is-compass-installed.sh
  kubectl apply -f ${TMP_DIR}/compass-installer.yaml || true
  sleep 2
  kubectl apply -f ${TMP_DIR}/compass-installer.yaml

  log::info "Installation triggered"
  "${TMP_DIR}"/is-compass-installed.sh --timeout 30m
}

function installControlPlane() {
  vars=(
    GATEWAY_IP_ADDRESS
  )

  utils::check_required_vars "${vars[@]}"

  log::info "Build Kyma Control Plane Installer Docker image"
  CLEANUP_DOCKER_IMAGE="true"
  KCP_INSTALLER_IMAGE="${KCP_INSTALLER_IMAGE}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-control-plane-image.sh

  log::info "Apply Kyma Control Plane config"
  kubectl create namespace "kcp-installer"
  applyCommonOverrides "kcp-installer"
  applyControlPlaneOverrides

  echo "Manual concatenating yamls"
  "${KCP_SCRIPTS_DIR}"/concat-yamls.sh "${INSTALLER_YAML}" "${INSTALLER_CR}" \
  | sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KCP_INSTALLER_IMAGE};" \
  | sed -e "s/__VERSION__/0.0.1/g" \
  | sed -e "s/__.*__//g" \
  | kubectl apply -f-
  
  log::info "Installation triggered"
  "${KCP_SCRIPTS_DIR}"/is-installed.sh --timeout 30m

  if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
    log::info "Create DNS Record for Apiserver proxy IP"
    APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    gcp::create_dns_record \
        -a "$APISERVER_IP_ADDRESS" \
        -h "apiserver" \
        -s "$COMMON_NAME" \
        -p "$CLOUDSDK_CORE_PROJECT" \
        -z "$CLOUDSDK_DNS_ZONE_NAME"
    CLEANUP_APISERVER_DNS_RECORD="true"
  fi
}

# Using set -f to prevent path globing in post_hook arguments.
# utils::post_hook call set +f at the beginning.
trap 'EXIT_STATUS=$?; docker_cleanup; set -f; utils::post_hook -n "$COMMON_NAME" -p "$CLOUDSDK_CORE_PROJECT" -c "$CLEANUP_CLUSTER" -g "$CLEANUP_GATEWAY_DNS_RECORD" -G "$INGRESS_GATEWAY_HOSTNAME" -a "$CLEANUP_APISERVER_DNS_RECORD" -A "$APISERVER_HOSTNAME" -I "$CLEANUP_GATEWAY_IP_ADDRESS" -l "$ERROR_LOGGING_GUARD" -z "$CLOUDSDK_COMPUTE_ZONE" -R "$CLOUDSDK_COMPUTE_REGION" -r "$PROVISION_REGIONAL_CLUSTER" -d "$DISABLE_ASYNC_DEPROVISION" -s "$COMMON_NAME" -e "$GATEWAY_IP_ADDRESS" -f "$APISERVER_IP_ADDRESS" -N "$COMMON_NAME" -Z "$CLOUDSDK_DNS_ZONE_NAME" -E "$EXIT_STATUS" -j "$JOB_NAME"; ' EXIT INT

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    log::info "Execute Job Guard"
    export JOB_NAME_PATTERN="(pre-control-plane-components-.*)|(pre-control-plane-tests-.*)"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

log::info "Create new cluster"
createCluster

utils::generate_self_signed_cert \
    -d "$DNS_DOMAIN" \
    -s "$COMMON_NAME" \
    -v "$SELF_SIGN_CERT_VALID_DAYS"
export TLS_CERT="${utils_generate_self_signed_cert_return_tls_cert:?}"
export TLS_KEY="${utils_generate_self_signed_cert_return_tls_key:?}"

log::info "Install Kyma"
installKyma

#log::info "Install Compass"
#installCompass

log::info "Install Control Plane"
installControlPlane

# enable test-log-collector before tests; if prowjob fails before test phase we do not have any reason to enable it earlier
if [[ "${BUILD_TYPE}" == "master" && -n "${LOG_COLLECTOR_SLACK_TOKEN}" ]]; then
  ENABLE_TEST_LOG_COLLECTOR=true
fi

log::info "Test Kyma and Control Plane"
"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh -l "release != compass"

log::success "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
