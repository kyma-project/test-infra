#!/usr/bin/env bash

#Description: Kyma Integration plan on GKE with external registry. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma on real GKE cluster.
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
# - KYMA_ARTIFACTS_BUCKET - GCP bucket
# - MACHINE_TYPE - (optional) GKE machine type
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

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/gcloud.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcloud.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"

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

  gcloud::cleanup

  if [ -n "${CLEANUP_DOCKER_IMAGE}" ]; then
    log::info "Docker image cleanup"
    if [ -n "${KYMA_INSTALLER_IMAGE}" ]; then
      log::info "Delete temporary Kyma-Installer Docker image"
      gcloud::authenticate "${GCR_PUSH_GOOGLE_APPLICATION_CREDENTIALS}"
      gcloud::delete_docker_image "${KYMA_INSTALLER_IMAGE}"
      gcloud::set_account "${GOOGLE_APPLICATION_CREDENTIALS}"
    fi
  fi

  MSG=""
  if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
  log::info "Job is finished ${MSG}"
  set -e

  exit "${EXIT_STATUS}"
}

trap post_hook EXIT INT

verify_internal_registry() {
    local pods
    pods="$(kubectl get pods --all-namespaces | grep docker-registry || true)"

    if [[ -n "${pods}" ]]; then
        echo "Pods with docker registry are running:"
        echo "${pods}"
        return 1
    fi

    return 0
}

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    log::info "Execute Job Guard"
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
    readonly COMMON_NAME_PREFIX="gkeext-pr"
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
    KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-external/${REPO_OWNER}/${REPO_NAME}:PR-${PULL_NUMBER}"
    export KYMA_INSTALLER_IMAGE
elif [[ "$BUILD_TYPE" == "release" ]]; then
    readonly COMMON_NAME_PREFIX="gkeext-rel"
    readonly RELEASE_VERSION=$(cat "VERSION")
    log::info "Reading release version from RELEASE_VERSION file, got: ${RELEASE_VERSION}"
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
else
    # Otherwise (master), operate on triggering commit id
    readonly COMMON_NAME_PREFIX="gkeext-commit"
    readonly COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
    KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-external/${REPO_OWNER}/${REPO_NAME}:COMMIT-${COMMIT_ID}"
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
KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

log::info "Authenticate"
gcloud::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"
docker::start
DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"

if [[ "$BUILD_TYPE" != "release" ]]; then
    log::info "Build Kyma-Installer Docker image"
    CLEANUP_DOCKER_IMAGE="true"
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-image.sh"
fi

log::info "Reserve IP Address for Ingressgateway"
GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
GATEWAY_IP_ADDRESS=$(gcloud::reserve_ip_address "${GATEWAY_IP_ADDRESS_NAME}")
CLEANUP_GATEWAY_IP_ADDRESS="true"
echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"


log::info "Create DNS Record for Ingressgateway IP"
GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
CLEANUP_GATEWAY_DNS_RECORD="true"
gcloud::create_dns_record "${GATEWAY_IP_ADDRESS}" "${GATEWAY_DNS_FULL_NAME}"


log::info "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
gcloud::create_network "${GCLOUD_NETWORK_NAME}" "${GCLOUD_SUBNET_NAME}"


log::info "Provision cluster: \"${CLUSTER_NAME}\""
export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
if [ -z "$MACHINE_TYPE" ]; then
      export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
fi
CLEANUP_CLUSTER="true"
gcloud::provision_gke_cluster "$CLUSTER_NAME"

log::info "Generate self-signed certificate"
DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
CERT_KEY=$(utils::generate_self_signed_cert "$DOMAIN")
TLS_CERT=$(echo "${CERT_KEY}" | head -1)
TLS_KEY=$(echo "${CERT_KEY}" | tail -1)

log::info "Apply Kyma config"

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

dockerPassword=/tmp/kyma-gke-integration/dockerPassword.json
mkdir -p /tmp/kyma-gke-integration
< "${GCR_PUSH_GOOGLE_APPLICATION_CREDENTIALS}" tr -d '\n' > /tmp/kyma-gke-integration/dockerPassword.json
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-secret.sh" --name "serverless-external-registry-overrides" \
    --data "dockerRegistry.enableInternal=false" \
    --data "dockerRegistry.username=_json_key" \
    --file "dockerRegistry.password=${dockerPassword}" \
    --data "dockerRegistry.serverAddress=$(echo "${DOCKER_PUSH_REPOSITORY}" | cut -d'/' -f1)" \
    --data "dockerRegistry.registryAddress=${DOCKER_PUSH_REPOSITORY}/functions" \
    --label "component=serverless"

if [[ "$BUILD_TYPE" == "release" ]]; then
    echo "Use released artifacts"
    gsutil cp "${KYMA_ARTIFACTS_BUCKET}/${RELEASE_VERSION}/kyma-installer-cluster.yaml" /tmp/kyma-gke-integration/downloaded-installer.yaml
    kubectl apply -f /tmp/kyma-gke-integration/downloaded-installer.yaml
else
    echo "Manual concatenating yamls"
    "${KYMA_SCRIPTS_DIR}"/concat-yamls.sh "${INSTALLER_YAML}" "${INSTALLER_CR}" \
    | sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" \
    | sed -e "s/__VERSION__/0.0.1/g" \
    | sed -e "s/__.*__//g" \
    | kubectl apply -f-
fi

log::info "Installation triggered"
"${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout 30m

if [ -n "$(kubectl get  service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
    log::info "Create DNS Record for Apiserver proxy IP"
    APISERVER_IP_ADDRESS=$(kubectl get  service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
    CLEANUP_APISERVER_DNS_RECORD="true"
    gcloud::create_dns_record "${APISERVER_IP_ADDRESS}" "${APISERVER_DNS_FULL_NAME}"
fi


log::info "Verify if internal docker registry is disabled"
verify_internal_registry

log::info "Test Kyma"
KYMA_TESTS="serverless serverless-long" "${TEST_INFRA_SOURCES_DIR}/prow/scripts/kyma-testing.sh"

log::success "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
