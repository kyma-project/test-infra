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
export TEST_INFRA_SOURCES_DIR="$KYMA_PROJECT_DIR/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="$TEST_INFRA_SOURCES_DIR/prow/scripts/cluster-integration/helpers"
# shellcheck source=prow/scripts/lib/log.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/log.sh"
# shellcheck disable=SC1090
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/testing-helpers.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"
# Enforce lowercase
readonly REPO_OWNER=${REPO_OWNER,,}
export REPO_OWNER
readonly REPO_NAME=${REPO_NAME,,}
export REPO_NAME
export INGRESS_GATEWAY_HOSTNAME='*'
export APISERVER_HOSTNAME='apiserver'

requiredVars=(
    KYMA_PROJECT_DIR
    CLOUDSDK_CORE_PROJECT
    CLOUDSDK_COMPUTE_REGION
    CLOUDSDK_DNS_ZONE_NAME
    GOOGLE_APPLICATION_CREDENTIALS
    GKE_CLUSTER_VERSION
    JOB_NAME
    INGRESS_GATEWAY_HOSTNAME
    APISERVER_HOSTNAME
)

utils::check_required_vars "${requiredVars[@]}"

# Using set -f to prevent path globing in post_hook arguments.
# utils::post_hook call set +f at the beginning.
trap 'EXIT_STATUS=$?; set -f; utils::post_hook -n "$COMMON_NAME" -p "$CLOUDSDK_CORE_PROJECT" -c "$CLEANUP_CLUSTER" -g "$CLEANUP_GATEWAY_DNS_RECORD" -G "$INGRESS_GATEWAY_HOSTNAME" -a "$CLEANUP_APISERVER_DNS_RECORD" -A "$APISERVER_HOSTNAME" -I "$CLEANUP_GATEWAY_IP_ADDRESS" -l "$ERROR_LOGGING_GUARD" -z "$CLOUDSDK_COMPUTE_ZONE" -R "$CLOUDSDK_COMPUTE_REGION" -r "$PROVISION_REGIONAL_CLUSTER" -d "$DISABLE_ASYNC_DEPROVISION" -s "$COMMON_NAME" -e "$GATEWAY_IP_ADDRESS" -f "$APISERVER_IP_ADDRESS" -N "$COMMON_NAME" -Z "$CLOUDSDK_DNS_ZONE_NAME" -E "$EXIT_STATUS" -j "$JOB_NAME"' EXIT INT

utils::generate_vars_for_build \
    -b "$BUILD_TYPE" \
    -p "$PULL_NUMBER" \
    -s "$PULL_BASE_SHA" \
    -n "$JOB_NAME"
export COMMON_NAME=${utils_generate_vars_for_build_return_commonName:?}
export KYMA_SOURCE=${utils_generate_vars_for_build_return_kymaSource:?}

gcp::set_vars_for_network \
    -n "$JOB_NAME"
export GCLOUD_NETWORK_NAME="${gcp_set_vars_for_network_return_net_name:?}"
export GCLOUD_SUBNET_NAME="${gcp_set_vars_for_network_return_subnet_name:?}"
#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

gcp::authenticate \
    -c "$GOOGLE_APPLICATION_CREDENTIALS"

gcp::create_network \
    -n "$GCLOUD_NETWORK_NAME" \
    -s "$GCLOUD_SUBNET_NAME" \
    -p "$CLOUDSDK_CORE_PROJECT"

gcp::reserve_ip_address \
    -n "$COMMON_NAME" \
    -p "$CLOUDSDK_CORE_PROJECT" \
    -r "$CLOUDSDK_COMPUTE_REGION"
export GATEWAY_IP_ADDRESS="${gcp_reserve_ip_address_return_ip_address:?}"
export CLEANUP_GATEWAY_IP_ADDRESS="true"

log::info "Create DNS Record for Ingressgateway IP"
gcp::create_dns_record \
    -p "$CLOUDSDK_CORE_PROJECT" \
    -z "$CLOUDSDK_DNS_ZONE_NAME" \
    -a "$GATEWAY_IP_ADDRESS" \
    -h "$INGRESS_GATEWAY_HOSTNAME" \
    -s "$COMMON_NAME"
export DNS_DOMAIN=${gcp_create_dns_record_return_dns_domain:?}
export DNS_SUBDOMAIN=${gcp_create_dns_record_return_dns_subdomain:?}
export CLEANUP_GATEWAY_DNS_RECORD="true"



if [ "$PROVISION_REGIONAL_CLUSTER" ]; then NUM_NODES="$NODES_PER_ZONE"; fi

gcp::provision_k8s_cluster \
    -c "$COMMON_NAME" \
    -p "$CLOUDSDK_CORE_PROJECT" \
    -v "$GKE_CLUSTER_VERSION" \
    -j "$JOB_NAME" \
    -J "$PROW_JOB_ID" \
    -l "$ADDITIONAL_LABELS" \
    -t "$TTL_HOURS" \
    -z "$CLOUDSDK_COMPUTE_ZONE" \
    -m "$MACHINE_TYPE" \
    -R "$CLOUDSDK_COMPUTE_REGION" \
    -n "$NUM_NODES" \
    -N "$GCLOUD_NETWORK_NAME" \
    -S "$GCLOUD_SUBNET_NAME" \
    -C "$GKE_RELEASE_CHANNEL" \
    -i "$IMAGE_TYPE" \
    -g "$GCLOUD_SECURITY_GROUP_DOMAIN" \
    -r "$PROVISION_REGIONAL_CLUSTER" \
    -s "$STACKDRIVER_KUBERNETESA" \
    -D "$CLUSTER_USE_SSD" \
    -e "$GKE_ENABLE_POD_SECURITY_POLICY" \
    -P "$TEST_INFRA_SOURCES_DIR"
export CLEANUP_CLUSTER="true"

utils::generate_self_signed_cert \
    -d "$DNS_DOMAIN" \
    -s "$COMMON_NAME" \
    -v "$SELF_SIGN_CERT_VALID_DAYS"
export TLS_CERT="${utils_generate_self_signed_cert_return_tls_cert:?}"
export TLS_KEY="${utils_generate_self_signed_cert_return_tls_key:?}"

log::info "Building Kyma CLI"
cd "$KYMA_PROJECT_DIR/cli"
make build-linux
mv "$KYMA_PROJECT_DIR/cli/bin/kyma-linux" "$KYMA_PROJECT_DIR/cli/bin/kyma"
export PATH="$KYMA_PROJECT_DIR/cli/bin:$PATH"


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


echo "$COMPONENT_OVERRIDES" > "$COMPONENT_OVERRIDES_FILE"

log::info "Kyma installation triggered"
kyma install \
    --ci \
    --source main \
    -o "$COMPONENT_OVERRIDES_FILE" \
    --domain "$DNS_SUBDOMAIN.${DNS_DOMAIN%.}" \
    --tls-cert "$TLS_CERT" \
    --tls-key "$TLS_KEY" \
    --timeout 90m

log::info "Checking the versions"
kyma version

if [ -n "$(kubectl get  service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
    log::info "Create DNS Record for Apiserver proxy IP"
    APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    gcp::create_dns_record \
        -p "$CLOUDSDK_CORE_PROJECT" \
        -z "$CLOUDSDK_DNS_ZONE_NAME" \
        -a "$APISERVER_IP_ADDRESS" \
        -h "$APISERVER_HOSTNAME" \
        -s "$COMMON_NAME"
    export CLEANUP_APISERVER_DNS_RECORD="true"
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
