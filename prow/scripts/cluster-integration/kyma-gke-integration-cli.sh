#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on GKE. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on real GKE cluster.
#
#
#Expected vars:
#
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - CLOUDSDK_CORE_PROJECT - GCP project for all GCP resources used during execution (Service Account, IP Address, DNS Zone, image registry etc.)
# - CLOUDSDK_COMPUTE_REGION - GCP compute region
# - CLOUDSDK_DNS_ZONE_NAME - GCP zone name (not its DNS name!)
# - GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path
# - GKE_CLUSTER_VERSION - GKE cluster version
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

readonly SUITE_NAME="testsuite-all-$(date '+%Y-%m-%d-%H-%M')"
readonly CONCURRENCY=5
#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/testing-helpers.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcloud.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcloud.sh"

requiredVars=(
    KYMA_PROJECT_DIR
    CLOUDSDK_CORE_PROJECT
    CLOUDSDK_COMPUTE_REGION
    CLOUDSDK_DNS_ZONE_NAME
    GOOGLE_APPLICATION_CREDENTIALS
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

  gcloud::cleanup

  MSG=""
  if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
  log::info "Job is finished ${MSG}"
  set -e

  exit "${EXIT_STATUS}"
}

trap post_hook EXIT INT

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c10)
readonly COMMON_NAME_PREFIX="cli-integration-test-gke"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")

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
DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"


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


log::info "Building Kyma CLI"
cd "${KYMA_PROJECT_DIR}/cli"
make build-linux
mv "${KYMA_PROJECT_DIR}/cli/bin/kyma-linux" "${KYMA_PROJECT_DIR}/cli/bin/kyma"
export PATH="${KYMA_PROJECT_DIR}/cli/bin:${PATH}"


COMPONENT_OVERRIDES_FILE="component-overrides.yaml"
COMPONENT_OVERRIDES=$(cat << EOF
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
  name: "istio-overrides"
  namespace: "kyma-installer"
  labels:
    installer: overrides
    kyma-project.io/installation: ""
    component: istio
data:
  kyma_istio_operator: |
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
)

echo "${COMPONENT_OVERRIDES}" > "${COMPONENT_OVERRIDES_FILE}"

log::info "Installing Kyma"
kyma install \
    --ci \
    --source main \
    -o "${COMPONENT_OVERRIDES_FILE}" \
    --domain "${DOMAIN}" \
    --tls-cert "${TLS_CERT}" \
    --tls-key "${TLS_KEY}" \
    --timeout 90m

log::info "Checking the versions"
kyma version


if [ -n "$(kubectl get  service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
    log::info "Create DNS Record for Apiserver proxy IP"
    APISERVER_IP_ADDRESS=$(kubectl get  service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
    CLEANUP_APISERVER_DNS_RECORD="true"
    gcloud::create_dns_record "${APISERVER_IP_ADDRESS}" "${APISERVER_DNS_FULL_NAME}"
fi

log::info "Create local resources for a sample Function"

kyma init function --name first-function

log::info "Apply local resources for the Function to the Kyma cluster"

kyma apply function

sleep 30

log::info "Check if the Function is running"

attempts=3
for ((i=1; i<=attempts; i++)); do
    set +e
    result=$(kubectl get pods -lserverless.kyma-project.io/function-name=first-function,serverless.kyma-project.io/resource=deployment -o jsonpath='{.items[0].status.phase}')
    set -e
    if [[ "$result" == *"Running"* ]]; then
        echo "The Function is in Running state"
        break
    elif [[ "${i}" == "${attempts}" ]]; then
        echo "ERROR: The Function is in ${result} state"
        exit 1
    fi
    echo "Sleep for 15 seconds"
    sleep 15
done

log::info "Running Kyma tests"

kyma test run \
    --name "${SUITE_NAME}" \
    --concurrency "${CONCURRENCY}" \
    --max-retries 1 \
    --timeout "1h" \
    --watch \
    --non-interactive


echo "Test Summary"
kyma test status "${SUITE_NAME}" -owide

statusSucceeded=$(kubectl get cts "${SUITE_NAME}"  -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
if [[ "${statusSucceeded}" != *"True"* ]]; then
    echo "- Fetching logs due to test suite failure"

    echo "- Fetching logs from testing pods in Failed status..."
    kyma test logs "${SUITE_NAME}" --test-status Failed

    echo "- Fetching logs from testing pods in Unknown status..."
    kyma test logs "${SUITE_NAME}" --test-status Unknown

    echo "- Fetching logs from testing pods in Running status due to running afer test suite timeout..."
    kyma test logs "${SUITE_NAME}" --test-status Running

    echo "ClusterTestSuite details"
    kubectl get cts "${SUITE_NAME}" -oyaml

    exit 1
fi

echo "ClusterTestSuite details"
kubectl get cts "${SUITE_NAME}" -oyaml


log::success "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
