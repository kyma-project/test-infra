#!/usr/bin/env bash

# Description: Kyma Upgradeability plan on GKE. The purpose of this script is to install last Kyma release on real GKE cluster, upgrade it with current changes and trigger testing.
#
#
# Expected vars:
#
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
#  - MACHINE_TYPE (optional): GKE machine type
#  - CLUSTER_VERSION (optional): GKE cluster version
#  - KYMA_ARTIFACTS_BUCKET: GCP bucket
#  - BOT_GITHUB_TOKEN: Bot github token used for API queries
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

discoverUnsetVar=false
enableTestLogCollector=false

for var in REPO_OWNER \
  REPO_NAME \
  DOCKER_PUSH_REPOSITORY \
  KYMA_PROJECT_DIR \
  CLOUDSDK_CORE_PROJECT \
  CLOUDSDK_COMPUTE_REGION \
  CLOUDSDK_DNS_ZONE_NAME \
  GOOGLE_APPLICATION_CREDENTIALS \
  KYMA_ARTIFACTS_BUCKET \
  BOT_GITHUB_TOKEN \
  DOCKER_IN_DOCKER_ENABLED \
  GCR_PUSH_GOOGLE_APPLICATION_CREDENTIALS; do
  if [[ -z "${!var}" ]]; then
    echo "ERROR: $var is not set"
    discoverUnsetVar=true
  fi
done
if [[ "${discoverUnsetVar}" = true ]]; then
  exit 1
fi

if [[ "${BUILD_TYPE}" == "master" ]]; then
  if [ -z "${LOG_COLLECTOR_SLACK_TOKEN}" ]; then
    echo "ERROR: LOG_COLLECTOR_SLACK_TOKEN is not set"
    exit 1
  fi
fi

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_INSTALL_TIMEOUT="30m"
export KYMA_UPDATE_TIMEOUT="40m"
export UPGRADE_TEST_PATH="${KYMA_SOURCES_DIR}/tests/end-to-end/upgrade/chart/upgrade"
export UPGRADE_TEST_NAMESPACE="e2e-upgrade-test"
export UPGRADE_TEST_RELEASE_NAME="${UPGRADE_TEST_NAMESPACE}"
export UPGRADE_TEST_RESOURCE_LABEL="kyma-project.io/upgrade-e2e-test"
export EXTERNAL_SOLUTION_TEST_PATH="${KYMA_SOURCES_DIR}/tests/end-to-end/external-solution-integration/chart/external-solution"
export EXTERNAL_SOLUTION_TEST_NAMESPACE="integration-test"
export EXTERNAL_SOLUTION_TEST_RELEASE_NAME="${EXTERNAL_SOLUTION_TEST_NAMESPACE}"
export EXTERNAL_SOLUTION_TEST_RESOURCE_LABEL="kyma-project.io/external-solution-e2e-test"
export TEST_RESOURCE_LABEL_VALUE_PREPARE="prepareData"
export HELM_TIMEOUT_SEC=10000s # timeout in sec for helm install/test operation
export TEST_TIMEOUT_SEC=600    # timeout in sec for test pods until they reach the terminating state
export TEST_CONTAINER_NAME="tests"

TMP_DIR=$(mktemp -d)
KYMA_LABEL_PREFIX="kyma-project.io"
KYMA_TEST_LABEL_PREFIX="${KYMA_LABEL_PREFIX}/test"
BEFORE_UPGRADE_LABEL_QUERY="${KYMA_TEST_LABEL_PREFIX}.before-upgrade=true"
POST_UPGRADE_LABEL_QUERY="${KYMA_TEST_LABEL_PREFIX}.after-upgrade=true"

# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"

# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"

# shellcheck source=prow/scripts/lib/gcloud.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcloud.sh"

# shellcheck source=prow/scripts/lib/testing-helpers.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/testing-helpers.sh"

# shellcheck source=prow/scripts/cluster-integration/helpers/kyma-cli.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/kyma-cli.sh"

cleanup() {
  ## Save status of failed script execution
  EXIT_STATUS=$?

  if [[ "${ERROR_LOGGING_GUARD}" = "true" ]]; then
    log::error "AN ERROR OCCURED! Take a look at preceding log entries."
  fi

  #Turn off exit-on-error so that next step is executed even if previous one fails.
  set +e

  log::banner "CLEANUP"

  # collect logs from failed tests before deprovisioning
  runTestLogCollector

  if [[ -n "${CLEANUP_CLUSTER}" ]]; then
    log::info "Deprovision cluster: \"${CLUSTER_NAME}\""

    #save disk names while the cluster still exists to remove them later
    DISKS=$(kubectl get pvc --all-namespaces -o jsonpath="{.items[*].spec.volumeName}" | xargs -n1 echo)
    export DISKS

    #Delete cluster
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/deprovision-gke-cluster.sh"

    #Delete orphaned disks
    log::info "Delete orphaned PVC disks..."
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/delete-disks.sh"
  fi

  if [[ -n "${CLEANUP_GATEWAY_DNS_RECORD}" ]]; then
    log::info "Delete Gateway DNS Record"
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${GATEWAY_DNS_FULL_NAME}" --address="${GATEWAY_IP_ADDRESS}" --dryRun=false
  fi

  if [[ -n "${CLEANUP_GATEWAY_IP_ADDRESS}" ]]; then
    log::info "Release Gateway IP Address"
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/release-ip-address.sh --project="${CLOUDSDK_CORE_PROJECT}" --ipname="${GATEWAY_IP_ADDRESS_NAME}" --region="${CLOUDSDK_COMPUTE_REGION}" --dryRun=false
  fi

  if [[ -n "${CLEANUP_DOCKER_IMAGE}" ]]; then
    log::info "Delete temporary Kyma-Installer Docker image"
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/delete-image.sh"
  fi

  if [ -n "${CLEANUP_APISERVER_DNS_RECORD}" ]; then
    log::info "Delete Apiserver proxy DNS Record"
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${APISERVER_DNS_FULL_NAME}" --address="${APISERVER_IP_ADDRESS}" --dryRun=false
  fi

  rm -rf "${TMP_DIR}"

  MSG=""
  if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
  log::info "Job is finished ${MSG}"
  set -e

  exit "${EXIT_STATUS}"
}

function installCli() {
  export INSTALL_DIR=${TMP_DIR}
  install::kyma_cli
}

runTestLogCollector() {
  if [ "${enableTestLogCollector}" = true ]; then
    if [[ "$BUILD_TYPE" == "master" ]]; then
      shout "Install test-log-collector"
      date
      export PROW_JOB_NAME="post-master-kyma-gke-upgrade"
      (
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/install-test-log-collector.sh" || true # we want it to work on "best effort" basis, which does not interfere with cluster
      )
    fi
  fi
}

trap cleanup EXIT INT

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
    readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    readonly RELEASE_VERSION=$(cat "${SCRIPT_DIR}/../../RELEASE_VERSION")
    log::info "Reading release version from RELEASE_VERSION file, got: ${RELEASE_VERSION}"
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
  else
    # Otherwise (master), operate on triggering commit id
    readonly COMMON_NAME_PREFIX="gke-upgrade-commit"
    COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
  fi

  ### Cluster name must be less than 40 characters!
  export CLUSTER_NAME="${COMMON_NAME}"

  export GCLOUD_NETWORK_NAME="${COMMON_NAME_PREFIX}-net"
  export GCLOUD_SUBNET_NAME="${COMMON_NAME_PREFIX}-subnet"
}

function reserveIPsAndCreateDNSRecords() {
  DNS_SUBDOMAIN="${COMMON_NAME}"
  log::info "Authenticate with GCP"

  # requires "${GOOGLE_APPLICATION_CREDENTIALS}"
  gcloud::authenticate

  # requires "$DOCKER_IN_DOCKER_ENABLED" (via preset), needed for building the new installer image
  docker::start

  DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"

  log::info "Reserve IP Address for Ingressgateway"
  GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
  GATEWAY_IP_ADDRESS=$(IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/reserve-ip-address.sh")
  CLEANUP_GATEWAY_IP_ADDRESS="true"
  log::info "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

  log::info "Create DNS Record for Ingressgateway IP"
  GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
  CLEANUP_GATEWAY_DNS_RECORD="true"
  IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"

  DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
  export DOMAIN
}

function generateAndExportCerts() {
  log::info "Generate self-signed certificate"
  CERT_KEY=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-self-signed-cert.sh")

  TLS_CERT=$(echo "${CERT_KEY}" | head -1)
  export TLS_CERT
  TLS_KEY=$(echo "${CERT_KEY}" | tail -1)
  export TLS_KEY
}

function createNetwork() {
  export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
  NETWORK_EXISTS=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/network-exists.sh")
  if [ "$NETWORK_EXISTS" -gt 0 ]; then
    log::info "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-network-with-subnet.sh"
  else
    log::info "Network ${GCLOUD_NETWORK_NAME} exists"
  fi
}

function createCluster() {
  log::banner "Provision cluster: \"${CLUSTER_NAME}\""
  ### For provision-gke-cluster.sh
  export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
  export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
  export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"
  if [[ -z "${MACHINE_TYPE}" ]]; then
    export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
  fi
  if [[ -z "${CLUSTER_VERSION}" ]]; then
    export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
  fi

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/provision-gke-cluster.sh"
  CLEANUP_CLUSTER="true"
}

function getLastReleaseVersion() {
  version=$(curl --silent --fail --show-error "https://api.github.com/repos/kyma-project/kyma/releases?access_token=${BOT_GITHUB_TOKEN}" |
    jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-1].tag_name')

  echo "${version}"
}

function installKyma() {
  kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
  mkdir -p /tmp/kyma-gke-upgradeability
  LAST_RELEASE_VERSION=$(getLastReleaseVersion)

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

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "istio-overrides" \
    --data "gateways.istio-ingressgateway.loadBalancerIP=${GATEWAY_IP_ADDRESS}" \
    --label "component=istio"

  log::info "Use released artifacts from version ${LAST_RELEASE_VERSION}"

  if [[ "$LAST_RELEASE_VERSION" == "1.14.0" ]]; then
    curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${LAST_RELEASE_VERSION}/kyma-installer-cluster.yaml" --output /tmp/kyma-gke-upgradeability/last-release-installer.yaml
    kubectl apply -f /tmp/kyma-gke-upgradeability/last-release-installer.yaml
  else
    curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${LAST_RELEASE_VERSION}/kyma-installer.yaml" --output /tmp/kyma-gke-upgradeability/kyma-installer.yaml
    curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${LAST_RELEASE_VERSION}/kyma-installer-cr-cluster.yaml" --output /tmp/kyma-gke-upgradeability/kyma-installer-cr-cluster.yaml

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

  # install external-solution test
  installTestChartOrFail "${EXTERNAL_SOLUTION_TEST_PATH}" "${EXTERNAL_SOLUTION_TEST_RELEASE_NAME}" "${EXTERNAL_SOLUTION_TEST_NAMESPACE}"
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

  if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
    log::info "Create DNS Record for Apiserver proxy IP"
    APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
    CLEANUP_APISERVER_DNS_RECORD="true"
    IP_ADDRESS=${APISERVER_IP_ADDRESS} DNS_FULL_NAME=${APISERVER_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
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

# Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

installCli

generateAndExportClusterName

reserveIPsAndCreateDNSRecords

generateAndExportCerts

createNetwork

createCluster

installKyma

createTestResources

enableTestLogCollector=true # enable test-log-collector before tests; if prowjob fails before test phase we do not have any reason to enable it earlier

testKyma "${BEFORE_UPGRADE_LABEL_QUERY}" testsuite-all-before-upgrade

upgradeKyma

testKyma "${POST_UPGRADE_LABEL_QUERY}" testsuite-all-after-upgrade

log::success "Job finished with success"

# Mark execution as successfully
ERROR_LOGGING_GUARD="false"
