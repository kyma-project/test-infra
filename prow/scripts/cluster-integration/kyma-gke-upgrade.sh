#!/usr/bin/env bash

# Description: Kyma Upgradeability plan on GKE. The purpose of this script is to install last Kyma release on real GKE cluster, upgrade it with current changes and trigger testing.
#
#
# Expected vars:
#
#  - SCENARIO_TYPE - Set up by prow, upgrade test scenario
#  - REPO_OWNER - Set up by prow, repository owner/organization
#  - REPO_NAME - Set up by prow, repository name
#  - BUILD_TYPE - Set up by prow, pr/master/release
#  - DOCKER_PUSH_REPOSITORY - Docker repository hostname
#  - DOCKER_PUSH_DIRECTORY - Docker "top-level" directory (with leading "/")
#  - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
#  - CLOUDSDK_CORE_PROJECT - GCP project for all GCP resources used during execution (Service Account, IP Address, DNS Zone, image registry etc.)
#  - CLOUDSDK_COMPUTE_REGION - GCP compute region
#  - CLOUDSDK_DNS_ZONE_NAME - GCP zone name (not its DNS name!)
#  - GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path
#  - GKE_CLUSTER_VERSION - GKE cluster version
#  - KYMA_ARTIFACTS_BUCKET: GCP bucket
#  - BOT_GITHUB_TOKEN: Bot github token used for API queries
#  - MACHINE_TYPE - (optional) GKE machine type
#
# Permissions: In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
#  - Compute Admin
#  - Kubernetes Engine Admin
#  - Kubernetes Engine Cluster Admin
#  - DNS Administrator
#  - Service Account User
#  - Storage Admin
#  - Compute Network Admin

set -o errexit

ENABLE_TEST_LOG_COLLECTOR=false

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_INSTALL_TIMEOUT="30m"
export KYMA_UPDATE_TIMEOUT="60m"
export UPGRADE_TEST_PATH="${KYMA_SOURCES_DIR}/tests/end-to-end/upgrade/chart/upgrade"
export UPGRADE_TEST_NAMESPACE="e2e-upgrade-test"
export UPGRADE_TEST_RELEASE_NAME="${UPGRADE_TEST_NAMESPACE}"
export UPGRADE_TEST_RESOURCE_LABEL="kyma-project.io/upgrade-e2e-test"
export TEST_RESOURCE_LABEL_VALUE_PREPARE="prepareData"
export HELM_TIMEOUT_SEC=10000s # timeout in sec for helm install/test operation
export TEST_TIMEOUT_SEC=600    # timeout in sec for test pods until they reach the terminating state
export TEST_CONTAINER_NAME="tests"

KYMA_LABEL_PREFIX="kyma-project.io"
KYMA_TEST_LABEL_PREFIX="${KYMA_LABEL_PREFIX}/test"
BEFORE_UPGRADE_LABEL_QUERY="${KYMA_TEST_LABEL_PREFIX}.before-upgrade=true"
POST_UPGRADE_LABEL_QUERY="${KYMA_TEST_LABEL_PREFIX}.after-upgrade=true"

# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"

# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"

# shellcheck source=prow/scripts/lib/testing-helpers.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/testing-helpers.sh"

# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"

# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"

requiredVars=(
    REPO_OWNER
    REPO_NAME
    DOCKER_PUSH_REPOSITORY
    KYMA_PROJECT_DIR
    CLOUDSDK_CORE_PROJECT
    CLOUDSDK_COMPUTE_REGION
    CLOUDSDK_DNS_ZONE_NAME
    GOOGLE_APPLICATION_CREDENTIALS
    KYMA_ARTIFACTS_BUCKET
    BOT_GITHUB_TOKEN
    DOCKER_IN_DOCKER_ENABLED
    GKE_CLUSTER_VERSION
)

utils::check_required_vars "${requiredVars[@]}"

function installCli() {
  kyma::install_cli
}

# Using set -f to prevent path globing in post_hook arguments.
# utils::post_hook call set +f at the beginning.
trap 'EXIT_STATUS=$?; set -f; utils::post_hook -n "$COMMON_NAME" -p "$CLOUDSDK_CORE_PROJECT" -c "$CLEANUP_CLUSTER" -g "$CLEANUP_GATEWAY_DNS_RECORD" -G "$INGRESS_GATEWAY_HOSTNAME" -a "$CLEANUP_APISERVER_DNS_RECORD" -A "$APISERVER_HOSTNAME" -I "$CLEANUP_GATEWAY_IP_ADDRESS" -l "$ERROR_LOGGING_GUARD" -z "$CLOUDSDK_COMPUTE_ZONE" -R "$CLOUDSDK_COMPUTE_REGION" -r "$PROVISION_REGIONAL_CLUSTER" -d "$DISABLE_ASYNC_DEPROVISION" -s "$COMMON_NAME" -e "$GATEWAY_IP_ADDRESS" -f "$APISERVER_IP_ADDRESS" -N "$COMMON_NAME" -Z "$CLOUDSDK_DNS_ZONE_NAME" -E "$EXIT_STATUS" -j "$JOB_NAME"' EXIT INT

if [[ "${BUILD_TYPE}" == "pr" ]]; then
  log::info "Execute Job Guard"
  "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

function generateAndExportClusterName() {
  readonly REPO_OWNER=$(echo "${REPO_OWNER}" | tr '[:upper:]' '[:lower:]')
  readonly REPO_NAME=$(echo "${REPO_NAME}" | tr '[:upper:]' '[:lower:]')
  readonly RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' </dev/urandom | head -c10)

  if [[ "$BUILD_TYPE" == "pr" ]]; then
    readonly COMMON_NAME_PREFIX="gke-upgrade-pr"
    # In case of PR, operate on PR number
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
  elif [[ "$BUILD_TYPE" == "release" ]]; then
    readonly COMMON_NAME_PREFIX="gke-upgrade-rel"
    readonly RELEASE_VERSION=$(cat "VERSION")
    log::info "Reading release version from RELEASE_VERSION file, got: ${RELEASE_VERSION}"
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
  else
    # Otherwise (master), operate on triggering commit id
    readonly COMMON_NAME_PREFIX="gke-upgrade-commit"
    COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
  fi

  gcp::set_vars_for_network \
    -n "$JOB_NAME"
  export GCLOUD_NETWORK_NAME="${gcp_set_vars_for_network_return_net_name:?}"
  export GCLOUD_SUBNET_NAME="${gcp_set_vars_for_network_return_subnet_name:?}"
}

function reserveIPsAndCreateDNSRecords() {
  DNS_SUBDOMAIN="${COMMON_NAME}"
  log::info "Authenticate with GCP"

  # requires "${GOOGLE_APPLICATION_CREDENTIALS}"
  gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"

  # requires "$DOCKER_IN_DOCKER_ENABLED" (via preset), needed for building the new installer image
  docker::start

  DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"

  log::info "Reserve IP Address for Ingressgateway"
  GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
  gcp::reserve_ip_address \
    -n "${GATEWAY_IP_ADDRESS_NAME}" \
    -p "$CLOUDSDK_CORE_PROJECT" \
    -r "$CLOUDSDK_COMPUTE_REGION"
  GATEWAY_IP_ADDRESS="${gcp_reserve_ip_address_return_ip_address:?}"
  CLEANUP_GATEWAY_IP_ADDRESS="true"
  log::info "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

  log::info "Create DNS Record for Ingressgateway IP"
  gcp::create_dns_record \
      -a "$GATEWAY_IP_ADDRESS" \
      -h "*" \
      -s "$COMMON_NAME" \
      -p "$CLOUDSDK_CORE_PROJECT" \
      -z "$CLOUDSDK_DNS_ZONE_NAME"
  CLEANUP_GATEWAY_DNS_RECORD="true"

  DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
  export DOMAIN
}

function getLastRCVersion() {
  version=$(curl --silent --fail --show-error -H "Authorization: token ${BOT_GITHUB_TOKEN}" "https://api.github.com/repos/kyma-project/kyma/releases" |
    jq -r 'del( .[] | select( (.prerelease == false) or (.draft == true) )) | .[0].tag_name ')
  
  echo "${version}"
}

function installKyma() {
  kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
  mkdir -p /tmp/kyma-gke-upgradeability
  kyma::get_last_release_version \
    -t "${BOT_GITHUB_TOKEN}"
  LAST_RELEASE_VERSION="${kyma_get_last_release_version_return_version:?}"
  if [ -z "$LAST_RELEASE_VERSION" ]; then
    log::error "Couldn't grab latest version from GitHub API, stopping."
    exit 1
  fi

  log::banner "Apply Kyma config from version ${LAST_RELEASE_VERSION}"
  kubectl create namespace "kyma-installer"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "installation-config-overrides" \
    --data "global.domainName=${DOMAIN}" \
    --data "global.loadBalancerIP=${GATEWAY_IP_ADDRESS}"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "core-test-ui-acceptance-overrides" \
    --data "test.acceptance.ui.logging.enabled=true" \
    --label "component=core"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "application-registry-overrides" \
    --data "application-registry.deployment.args.detailedErrorResponse=true" \
    --label "component=application-connector"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "cluster-certificate-overrides" \
    --data "global.tlsCrt=${TLS_CERT}" \
    --data "global.tlsKey=${TLS_KEY}"

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

  log::info "Use released artifacts from version ${LAST_RELEASE_VERSION}"

  if [[ "$LAST_RELEASE_VERSION" == "1.14.0" ]]; then
    curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${LAST_RELEASE_VERSION}/kyma-installer-cluster.yaml" --output /tmp/kyma-gke-upgradeability/last-release-installer.yaml
    kubectl apply -f /tmp/kyma-gke-upgradeability/last-release-installer.yaml
  else
    curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${LAST_RELEASE_VERSION}/kyma-installer.yaml" --output /tmp/kyma-gke-upgradeability/kyma-installer.yaml
    curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${LAST_RELEASE_VERSION}/kyma-installer-cr-cluster.yaml" --output /tmp/kyma-gke-upgradeability/kyma-installer-cr-cluster.yaml

    kubectl apply -f /tmp/kyma-gke-upgradeability/kyma-installer.yaml || true
    sleep 2
    kubectl apply -f /tmp/kyma-gke-upgradeability/kyma-installer.yaml
    kubectl apply -f /tmp/kyma-gke-upgradeability/kyma-installer-cr-cluster.yaml
  fi

  log::banner "Installation triggered with timeout ${KYMA_INSTALL_TIMEOUT}"
  "${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout ${KYMA_INSTALL_TIMEOUT}
}

function installTestChartOrFail() {
  local path=$1
  local name=$2
  local namespace=$3

  log::info "Create ${name} resources"

  helm install "${name}" \
    --namespace "${namespace}" \
    --create-namespace \
    "${path}" \
    --timeout "${HELM_TIMEOUT_SEC}" \
    --set domain="${DOMAIN}" \
    --wait

  prepareResult=$?
  if [[ "${prepareResult}" != 0 ]]; then
    echo "Helm install ${name} operation failed: ${prepareResult}"
    exit "${prepareResult}"
  fi
}

function createTestResources() {
  log::banner "Install additional charts"
  # install upgrade test
  installTestChartOrFail "${UPGRADE_TEST_PATH}" "${UPGRADE_TEST_RELEASE_NAME}" "${UPGRADE_TEST_NAMESPACE}"
}

function upgradeKymaToRelease() {
  log::banner "Updating Kyma ${KYMA_UPDATE_TIMEOUT} to Version ${RELEASE_VERSION}"
  log::info "Delete the kyma-installation CR and kyma-installer deployment"
  # Remove the finalizer form kyma-installation the merge type is used because strategic is not supported on CRD.
  # More info about merge strategy can be found here: https://tools.ietf.org/html/rfc7386
  kubectl patch Installation kyma-installation -n default --patch '{"metadata":{"finalizers":null}}' --type=merge
  kubectl delete Installation -n default kyma-installation

  # Remove the current installer to prevent it performing any action.
  kubectl delete deployment -n kyma-installer kyma-installer

  echo "Use released artifacts"
  gsutil cp "${KYMA_ARTIFACTS_BUCKET}/${RELEASE_VERSION}/kyma-installer-cluster.yaml" /tmp/kyma-gke-upgradeability/new-release-kyma-installer.yaml

  log::info "Update kyma installer"
  kubectl apply -f /tmp/kyma-gke-upgradeability/new-release-kyma-installer.yaml

  log::info "Update triggered with timeout ${KYMA_UPDATE_TIMEOUT}"
  "${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout ${KYMA_UPDATE_TIMEOUT}
}

function upgradeKymaToCheckedoutVersion() {
  log::banner "Updating Kyma with timeout ${KYMA_UPDATE_TIMEOUT} from local sources"

  local KYMA_COMMIT_ID
  KYMA_COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)

  local TEST_INFRA_COMMIT_ID
  TEST_INFRA_COMMIT_ID=$(cd "$TEST_INFRA_SOURCES_DIR" && git rev-parse --short HEAD)

  local IMAGE_REPO_NAME="$REPO_NAME"
  if [[ "$IMAGE_REPO_NAME" == "test-infra" ]]; then
    IMAGE_REPO_NAME="kyma"
  fi

  KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-upgradeability/${REPO_OWNER}/${IMAGE_REPO_NAME}:KYMA-${KYMA_COMMIT_ID:-unknown}-TI-${TEST_INFRA_COMMIT_ID:-unknown}"

  (
    set -x
    kyma upgrade \
      --ci \
      --source local \
      --src-path "${KYMA_SOURCES_DIR}" \
      --custom-image "${KYMA_INSTALLER_IMAGE}" \
      --components "${KYMA_SOURCES_DIR}/installation/resources/installer-cr-cluster.yaml.tpl" \
      --timeout "${KYMA_UPDATE_TIMEOUT}"
  )
}

function upgradeKymaToCommit() {
  COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse HEAD)
  log::banner "Updating Kyma with timeout ${KYMA_UPDATE_TIMEOUT} to Kyma commit: ${COMMIT_ID}"
  (
    set -x
    kyma upgrade \
      --ci \
      --source "${COMMIT_ID}" \
      --timeout "${KYMA_UPDATE_TIMEOUT}"
  )
}

function upgradeKyma() {
  if [[ "$BUILD_TYPE" == "release" ]]; then
    upgradeKymaToRelease
  elif [[ "$BUILD_TYPE" == "pr" ]]; then
    upgradeKymaToCheckedoutVersion
  else
    upgradeKymaToCommit
  fi
}

function createDNSRecord() {
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

# testKyma executes the kyma-testing.sh. labelqueries can be passed as arguments and will be passed to the kyma cli
function testKyma() {
  testing::inject_addons_if_necessary

  local labelquery=${1}
  local suitename=${2}
  local test_args=()

  if [[ -n ${labelquery} ]]; then
    test_args+=("-l")
    test_args+=("${labelquery}")
  fi

  if [[ -n ${suitename} ]]; then
    test_args+=("-n")
    test_args+=("${suitename}")
  fi

  log::banner "Test Kyma " "${test_args[@]}"
  "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh "${test_args[@]}"

  testing::remove_addons_if_necessary
}

function applyScenario() {
  if [ "$SCENARIO_TYPE" == "pre" ]; then
    testKyma "${BEFORE_UPGRADE_LABEL_QUERY}" testsuite-all-before-upgrade
    upgradeKyma
    createDNSRecord
    testKyma "${POST_UPGRADE_LABEL_QUERY}" testsuite-all-after-upgrade
  elif [ "$SCENARIO_TYPE" == "post" ] || [ "$SCENARIO_TYPE" == "release" ]; then
    testKyma "${BEFORE_UPGRADE_LABEL_QUERY}" testsuite-all-before-upgrade
    upgradeKyma
    createDNSRecord
    upgradeKyma
    testKyma "${POST_UPGRADE_LABEL_QUERY}" testsuite-all-after-upgrade
  fi
}

# Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

installCli

generateAndExportClusterName

reserveIPsAndCreateDNSRecords

utils::generate_self_signed_cert \
    -d "$DNS_DOMAIN" \
    -s "$COMMON_NAME" \
    -v "$SELF_SIGN_CERT_VALID_DAYS"
export TLS_CERT="${utils_generate_self_signed_cert_return_tls_cert:?}"
export TLS_KEY="${utils_generate_self_signed_cert_return_tls_key:?}"

gcp::create_network \
    -n "${GCLOUD_NETWORK_NAME}" \
    -s "${GCLOUD_SUBNET_NAME}" \
    -p "$CLOUDSDK_CORE_PROJECT"

export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"

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
    -D "$CLUSTER_USE_SSD" \
    -P "$TEST_INFRA_SOURCES_DIR" \
    -r "$PROVISION_REGIONAL_CLUSTER"

CLEANUP_CLUSTER="true"

installKyma

createTestResources

# enable test-log-collector before tests; if prowjob fails before test phase we do not have any reason to enable it earlier
if [[ "${BUILD_TYPE}" == "master" && -n "${LOG_COLLECTOR_SLACK_TOKEN}" ]]; then
  ENABLE_TEST_LOG_COLLECTOR=true
fi

applyScenario

log::success "Job finished with success"

# Mark execution as successfully
ERROR_LOGGING_GUARD="false"
