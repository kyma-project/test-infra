#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export COMPASS_SOURCES_DIR="/home/prow/go/src/github.com/kyma-incubator/compass"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

readonly COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET="${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/compass"

# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/kyma.sh"

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
  readonly COMMON_NAME_PREFIX="gkecompint-pr"
  COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}")
  COMPASS_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-compass-benchmark/${REPO_OWNER}/${REPO_NAME}:PR-${PULL_NUMBER}"
  export COMPASS_INSTALLER_IMAGE
else
  # Otherwise (main), operate on triggering commit id
  readonly COMMON_NAME_PREFIX="gkecompint-commit"
  readonly COMMIT_ID=$(cd "$COMPASS_SOURCES_DIR" && git rev-parse --short HEAD)
  COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}")
  COMPASS_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-compass-benchmark/${REPO_OWNER}/${REPO_NAME}:COMMIT-${COMMIT_ID}"
  export COMPASS_INSTALLER_IMAGE
fi

### Cluster name must be less than 40 characters!
COMMON_NAME=$(echo "${COMMON_NAME}" | tr "[:upper:]" "[:lower:]")

gcp::set_vars_for_network \
  -n "$JOB_NAME"
export GCLOUD_NETWORK_NAME="${gcp_set_vars_for_network_return_net_name:?}"
export GCLOUD_SUBNET_NAME="${gcp_set_vars_for_network_return_subnet_name:?}"

#Local variables
COMPASS_SCRIPTS_DIR="${COMPASS_SOURCES_DIR}/installation/scripts"
COMPASS_RESOURCES_DIR="${COMPASS_SOURCES_DIR}/installation/resources"

INSTALLER_YAML="${COMPASS_RESOURCES_DIR}/installer.yaml"
INSTALLER_CR="${COMPASS_RESOURCES_DIR}/installer-cr.yaml.tpl"

# post_hook runs at the end of a script or on any error
function docker_cleanup() {
  #Turn off exit-on-error so that next step is executed even if previous one fails.
  set +e

  if [ -n "${CLEANUP_DOCKER_IMAGE}" ]; then
    log::info "Docker image cleanup"
    if [ -n "${COMPASS_INSTALLER_IMAGE}" ]; then
      log::info "Delete temporary Compass-Installer Docker image"
      gcp::delete_docker_image \
        -i "${COMPASS_INSTALLER_IMAGE}"
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
  DOMAIN="${COMMON_NAME}.${DNS_DOMAIN%?}"
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

  cat << EOF > "$PWD/kyma_istio_operator"
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  namespace: istio-system
spec:
  values:
    global:
      proxy:
        holdApplicationUntilProxyStarts: true
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
  OVERRIDES_FILE="${COMPASS_SOURCES_DIR}/installation/resources/installer-config-gke-benchmark.yaml.tpl"
  if [[ -f "$OVERRIDES_FILE" ]]; then
    # envsubst requires variables to be exported or to be passed to the process execution in order to work
  CLOUDSDK_CORE_PROJECT=${CLOUDSDK_CORE_PROJECT} CLOUDSDK_COMPUTE_ZONE=${CLOUDSDK_COMPUTE_ZONE} COMMON_NAME=${COMMON_NAME} envsubst < "$OVERRIDES_FILE" | kubectl apply -f -
  fi
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
    COMPASS_VERSION="PR-${PULL_NUMBER}"
  else
    COMPASS_VERSION="main-${COMMIT_ID}"
  fi
  COMPASS_ARTIFACTS="${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/${COMPASS_VERSION}"
  
  TMP_DIR="/tmp/compass-gke-integration"

  gsutil cp "${COMPASS_ARTIFACTS}/kyma-installer.yaml" ${TMP_DIR}/kyma-installer.yaml
  gsutil cp "${COMPASS_ARTIFACTS}/is-kyma-installed.sh" ${TMP_DIR}/is-kyma-installed.sh
  chmod +x ${TMP_DIR}/is-kyma-installed.sh
  kubectl apply -f ${TMP_DIR}/kyma-installer.yaml || true
  sleep 2
  kubectl apply -f ${TMP_DIR}/kyma-installer.yaml

  log::info "Installation triggered. Sleeping for 10 seconds before applying check-kyma-installer..."
  sleep 10
  "${TMP_DIR}"/is-kyma-installed.sh --timeout 30m
}

function installCompassOld() {
  kubectl create namespace "compass-installer"
  applyCommonOverrides "compass-installer"
  applyCompassOverrides

  TMP_DIR="/tmp/compass-main-artifacts"

  readonly LATEST_VERSION=main-$(cd "$COMPASS_SOURCES_DIR" && git rev-parse --short main~1)
  echo "Deploying compass version $LATEST_VERSION"

  COMPASS_ARTIFACTS="${COMPASS_DEVELOPMENT_ARTIFACTS_BUCKET}/${LATEST_VERSION}"

  gsutil cp "${COMPASS_ARTIFACTS}/compass-installer.yaml" ${TMP_DIR}/compass-installer.yaml
  kubectl apply -f ${TMP_DIR}/compass-installer.yaml

  log::info "Installation triggered. Sleeping for 10 seconds before applying check-compass-installer..."
  sleep 10
  "${COMPASS_SCRIPTS_DIR}"/is-installed.sh --timeout 30m
}

function installCompassNew() {
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

  log::info "Build Compass-Installer Docker image"
  CLEANUP_DOCKER_IMAGE="true"
  COMPASS_INSTALLER_IMAGE="${COMPASS_INSTALLER_IMAGE}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-compass-image.sh

  echo "Manual concatenating yamls"
  "${COMPASS_SCRIPTS_DIR}"/concat-yamls.sh "${INSTALLER_YAML}" "${INSTALLER_CR}" \
  | sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${COMPASS_INSTALLER_IMAGE};" \
  | sed -e "s/__VERSION__/0.0.1/g" \
  | sed -e "s/__.*__//g" \
  | kubectl apply -f-

  log::info "Installation triggered. Sleeping for 10 seconds before applying check-compass-installer..."
  sleep 10
  "${COMPASS_SCRIPTS_DIR}"/is-installed.sh --timeout 30m

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
    export JOB_NAME_PATTERN="(pre-compass-components-.*)|(pre-compass-tests-.*)"
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

log::info "Choose node for benchmarks execution"
NODE=$(kubectl get nodes | tail -n 1 | cut -d ' ' -f 1)

log::info "Benchmarks will be executed on node: $NODE. Will make it unschedulable."
kubectl label node "$NODE" benchmark=true
kubectl cordon "$NODE"

log::info "Install Kyma"
installKyma

log::info "Install Compass version from main"
installCompassOld

readonly SUITE_NAME="testsuite-all"

log::info "Execute benchmarks on the current main"
kubectl uncordon "$NODE"
CONCURRENCY=1 "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh -l "benchmark=true"
kubectl cordon "$NODE"

PODS=$(kubectl get cts $SUITE_NAME -o=go-template --template='{{range .status.results}}{{range .executions}}{{printf "%s\n" .id}}{{end}}{{end}}')
for POD in $PODS; do
  CONTAINER=$(kubectl -n kyma-system get pod "$POD" -o jsonpath='{.spec.containers[*].name}' | sed s/istio-proxy//g | awk '{$1=$1};1')
  kubectl logs -n kyma-system "$POD" -c "$CONTAINER" > "$CONTAINER"-old
done

kubectl delete cts $SUITE_NAME

# Because of sequential compass installation, the second one fails due to compass-migration job patching failure. K8s job's fields are immutable.
log::info "Deleting the old compass-migration job"
kubectl delete jobs -n compass-system compass-migration

log::info "Install New Compass version"
installCompassNew

log::info "Execute benchmarks on the new release"
kubectl uncordon "$NODE"
CONCURRENCY=1 "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh -l "benchmark=true"
kubectl cordon "$NODE"

PODS=$(kubectl get cts $SUITE_NAME -o=go-template --template='{{range .status.results}}{{range .executions}}{{printf "%s\n" .id}}{{end}}{{end}}')

CHECK_FAILED=false
FAILED_TESTS=''

for POD in $PODS; do
  CONTAINER=$(kubectl -n kyma-system get pod "$POD" -o jsonpath='{.spec.containers[*].name}' | sed s/istio-proxy//g | awk '{$1=$1};1')
  kubectl logs -n kyma-system "$POD" -c "$CONTAINER" > "$CONTAINER"-new

  if [ -f "$CONTAINER"-old ]; then
    log::info "Stats of the main installation"
    benchstat "$CONTAINER"-old

    log::info "Stats of the new installation"
    benchstat "$CONTAINER"-new

    STATS=$(benchstat "$CONTAINER"-old "$CONTAINER"-new)
    log::info "Performance comparison statistics"
    echo "$STATS"

    DELTA=$(echo -n "$STATS" | tail +2 | { grep -v '~' || true; } | awk '{print $(NF-2)}')
    if [[ $DELTA == +* ]]; then # If delta is positive
      log::error "There is significant performance degradation in the new release!"
      CHECK_FAILED=true
      FAILED_TESTS="$CONTAINER\\n$FAILED_TESTS"
    fi
  else
    benchstat "$CONTAINER"-new
  fi
done

if [ $CHECK_FAILED = true ]; then
  log::error "The following benchmark tests failed:\\n $FAILED_TESTS"
  exit 1
fi

log::success "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
