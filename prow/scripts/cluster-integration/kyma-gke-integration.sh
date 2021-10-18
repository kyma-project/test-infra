#!/usr/bin/env bash

#Description: Kyma Integration plan on GKE. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma on real GKE cluster.
#
#
#Expected vars:
#
# - REPO_OWNER - Set up by prow, repository owner/organization
# - REPO_NAME - Set up by prow, repository name
# - BUILD_TYPE - Set up by prow, pr/master/release
# - DOCKER_PUSH_REPOSITORY - Docker repository hostname
# - DOCKER_PUSH_DIRECTORY - Docker "top-level" directory (with leading "/")
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - CLOUDSDK_CORE_PROJECT - GCP project for all GCP resources used during execution (Service Account, IP Address, DNS Zone, image registry etc.)
# - CLOUDSDK_COMPUTE_REGION - GCP compute region
# - CLOUDSDK_DNS_ZONE_NAME - GCP zone name (not its DNS name!)
# - GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path
# - GKE_CLUSTER_VERSION - GKE cluster version
# - KYMA_ARTIFACTS_BUCKET: GCP bucket
# - LOG_COLLECTOR_SLACK_TOKEN: Slack token for test-log-collector
# - MACHINE_TYPE - (optional) GKE machine type
# - GKE_RELEASE_CHANNEL - (optional) GKE release channel
#
#Permissions: In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
# - Compute Admin
# - Kubernetes Engine Admin
# - Kubernetes Engine Cluster Admin
# - DNS Administrator
# - Service Account User
# - Storage Admin
# - Compute Network Admin

set -o errexit

ENABLE_TEST_LOG_COLLECTOR=false

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

KYMA_LABEL_PREFIX="kyma-project.io"
KYMA_TEST_LABEL_PREFIX="${KYMA_LABEL_PREFIX}/test"
INTEGRATION_TEST_LABEL_QUERY="${KYMA_TEST_LABEL_PREFIX}.integration=true"

# shellcheck source=prow/scripts/lib/gcloud.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcloud.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

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
    GCR_PUSH_GOOGLE_APPLICATION_CREDENTIALS
    GKE_CLUSTER_VERSION
)

utils::check_required_vars "${requiredVars[@]}"

# post_hook runs at the end of a script or on any error
function post_hook() {
  #!!! Must be at the beginning of this function !!!
  EXIT_STATUS=$?

  log::info "Cleanup"

  if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
    log::info "AN ERROR OCCURED! Take a look at preceding log entries."
  fi

  #Turn off exit-on-error so that next step is executed even if previous one fails.
  set +e

  # collect logs from failed tests before deprovisioning
  kyma::run_test_log_collector "post-main-kyma-gke-integration"

  gcloud::cleanup

  MSG=""
  if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
  log::info "Job is finished ${MSG}"
  set -e

  exit "${EXIT_STATUS}"
}

trap post_hook EXIT INT

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    log::info "Execute Job Guard"
    # shellcheck disable=SC2031
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

# Enforce lowercase
readonly REPO_OWNER=$(echo "${REPO_OWNER}" | tr '[:upper:]' '[:lower:]')
export REPO_OWNER
readonly REPO_NAME=$(echo "${REPO_NAME}" | tr '[:upper:]' '[:lower:]')
export REPO_NAME

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c10)

if [[ "$BUILD_TYPE" == "pr" ]]; then
    # In case of PR, operate on PR number
    readonly COMMON_NAME_PREFIX="gkeint-pr"
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
    KYMA_SOURCE="PR-${PULL_NUMBER}"
elif [[ "$BUILD_TYPE" == "release" ]]; then
    readonly COMMON_NAME_PREFIX="gkeint-rel"
    readonly RELEASE_VERSION=$(cat "VERSION")
    log::info "Reading release version from RELEASE_VERSION file, got: ${RELEASE_VERSION}"
    KYMA_SOURCE="${RELEASE_VERSION}"
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
else
    # Otherwise (master), operate on triggering commit id
    readonly COMMON_NAME_PREFIX="gkeint-commit"
    readonly COMMIT_ID="${PULL_BASE_SHA::8}"
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
    KYMA_SOURCE="${COMMIT_ID}"
    export KYMA_INSTALLER_IMAGE
fi


### Cluster name must be less than 40 characters!
export CLUSTER_NAME="${COMMON_NAME}"

export GCLOUD_NETWORK_NAME="${COMMON_NAME_PREFIX}-net"
export GCLOUD_SUBNET_NAME="${COMMON_NAME_PREFIX}-subnet"

### For gcloud::provision_gke_cluster
export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"

#Local variables
DNS_SUBDOMAIN="${COMMON_NAME}"

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

log::info "Authenticate"
gcloud::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"
kyma::install_cli

DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"

log::info "Reserve IP Address for Ingressgateway"
GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
export GATEWAY_IP_ADDRESS
GATEWAY_IP_ADDRESS=$(gcloud::reserve_ip_address "$GATEWAY_IP_ADDRESS_NAME")
export CLEANUP_GATEWAY_IP_ADDRESS="true"
log::info "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"


log::info "Create DNS Record for Ingressgateway IP"
GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
gcloud::create_dns_record "$GATEWAY_IP_ADDRESS" "$GATEWAY_DNS_FULL_NAME"
export CLEANUP_GATEWAY_DNS_RECORD="true"

log::info "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
gcloud::create_network "$GCLOUD_NETWORK_NAME" "$GCLOUD_SUBNET_NAME"

log::banner "Provision cluster: \"${CLUSTER_NAME}\""
if [ -z "$MACHINE_TYPE" ]; then
      export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
fi

# if GKE_RELEASE_CHANNEL is set, get latest possible cluster version
gcloud::set_latest_cluster_version_for_channel

# serverless tests are failing when are running on a cluster with contianerD
#if [[ "${GKE_RELEASE_CHANNEL}" == "rapid" ]]; then
  # set image type to the image that uses docker instead of containerD
export IMAGE_TYPE="cos"
#fi

gcloud::provision_gke_cluster "$CLUSTER_NAME"
export CLEANUP_CLUSTER="true"

log::info "Generate self-signed certificate"
DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
CERT_KEY=$(utils::generate_self_signed_cert "$DOMAIN")
TLS_CERT=$(echo "${CERT_KEY}" | head -1)
TLS_KEY=$(echo "${CERT_KEY}" | tail -1)

log::info "Create Kyma CLI overrides"
envsubst < "${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/kyma-installer-overrides.tpl.yaml" > "$PWD/kyma-installer-overrides.yaml"

log::info "Installation triggered"

yes | kyma install \
  --ci \
  -s "${KYMA_SOURCE}" \
  -o "$PWD/kyma-installer-overrides.yaml" \
  --domain "${DOMAIN}" \
  --tls-cert="${TLS_CERT}" \
  --tls-key="${TLS_KEY}" \
  --timeout 60m

if [ -n "$(kubectl get  service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
    log::info "Create DNS Record for Apiserver proxy IP"
    APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"

    gcloud::create_dns_record "$APISERVER_IP_ADDRESS" "$APISERVER_DNS_FULL_NAME"
    export CLEANUP_APISERVER_DNS_RECORD="true"
fi

log::info "Collect list of images"
if [ -z "$ARTIFACTS" ] ; then
    ARTIFACTS=/tmp/artifacts
fi

# generate pod-security-policy list in json
utils::save_psp_list "${ARTIFACTS}/kyma-psp.json"

utils::kubeaudit_create_report "${ARTIFACTS}/kubeaudit.log"
utils::kubeaudit_check_report "${ARTIFACTS}/kubeaudit.log"

# enable test-log-collector before tests; if prowjob fails before test phase we do not have any reason to enable it earlier
if [[ "${BUILD_TYPE}" == "master" && -n "${LOG_COLLECTOR_SLACK_TOKEN}" ]]; then
  export ENABLE_TEST_LOG_COLLECTOR=true
fi

log::info "Test Kyma"
# shellcheck disable=SC2031
# TODO (@Ressetkk): Kyma test functions as a separate library
"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh "${INTEGRATION_TEST_LABEL_QUERY}"

log::success "Integration Test successful"

#!!! Must be at the end of the script !!!
export ERROR_LOGGING_GUARD="false"
