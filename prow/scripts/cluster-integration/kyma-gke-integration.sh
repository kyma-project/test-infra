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

#Exported variables
export TEST_INFRA_SOURCES_DIR="$KYMA_PROJECT_DIR/test-infra"
export KYMA_SOURCES_DIR="$KYMA_PROJECT_DIR/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="$TEST_INFRA_SOURCES_DIR/prow/scripts/cluster-integration/helpers"

# shellcheck source=prow/scripts/lib/kyma.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"

ENABLE_TEST_LOG_COLLECTOR=false

# Enforce lowercase
readonly REPO_OWNER=${REPO_OWNER,,}
export REPO_OWNER
# Enforce lowercase
readonly REPO_NAME=${REPO_NAME,,}
export REPO_NAME
export INGRESS_GATEWAY_HOSTNAME='*'
export APISERVER_HOSTNAME='apiserver'

# Used by kyma-testing.sh as an argument.
KYMA_LABEL_PREFIX="kyma-project.io"
KYMA_TEST_LABEL_PREFIX="${KYMA_LABEL_PREFIX}/test"
INTEGRATION_TEST_LABEL_QUERY="${KYMA_TEST_LABEL_PREFIX}.integration=true"

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
    GKE_CLUSTER_VERSION
)
# allow GKE_CLUSTER_VERSION to be specified explicitly for this script
GKE_CLUSTER_VERSION=${GKE_CLUSTER_VERSION_OVERRIDE:-$GKE_CLUSTER_VERSION}
utils::check_required_vars "${requiredVars[@]}"

# Using set -f to prevent path globing in post_hook arguments.
# utils::post_hook call set +f at the beginning.
trap 'EXIT_STATUS=$?; set -f; utils::post_hook -n "$COMMON_NAME" -p "$CLOUDSDK_CORE_PROJECT" -c "$CLEANUP_CLUSTER" -g "$CLEANUP_GATEWAY_DNS_RECORD" -G "$INGRESS_GATEWAY_HOSTNAME" -a "$CLEANUP_APISERVER_DNS_RECORD" -A "$APISERVER_HOSTNAME" -I "$CLEANUP_GATEWAY_IP_ADDRESS" -l "$ERROR_LOGGING_GUARD" -z "$CLOUDSDK_COMPUTE_ZONE" -R "$CLOUDSDK_COMPUTE_REGION" -r "$PROVISION_REGIONAL_CLUSTER" -d "$DISABLE_ASYNC_DEPROVISION" -s "$COMMON_NAME" -e "$GATEWAY_IP_ADDRESS" -f "$APISERVER_IP_ADDRESS" -N "$COMMON_NAME" -Z "$CLOUDSDK_DNS_ZONE_NAME" -E "$EXIT_STATUS" -j "$JOB_NAME"' EXIT INT

utils::run_jobguard \
    -b "$BUILD_TYPE" \
    -P "$TEST_INFRA_SOURCES_DIR"

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

kyma::install_cli

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

# if GKE_RELEASE_CHANNEL is set, get latest possible cluster version
if [ "${GKE_RELEASE_CHANNEL}" ]; then
    gcp::set_latest_cluster_version_for_channel \
        -C "$GKE_RELEASE_CHANNEL"
    GKE_CLUSTER_VERSION="${gcp_set_latest_cluster_version_for_channel_return_cluster_version:?}"
fi

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
    -s "$STACKDRIVER_KUBERNETES" \
    -D "$CLUSTER_USE_SSD" \
    -P "$TEST_INFRA_SOURCES_DIR"
export CLEANUP_CLUSTER="true"

utils::generate_self_signed_cert \
    -d "$DNS_DOMAIN" \
    -s "$COMMON_NAME" \
    -v "$SELF_SIGN_CERT_VALID_DAYS"
export TLS_CERT="${utils_generate_self_signed_cert_return_tls_cert:?}"
export TLS_KEY="${utils_generate_self_signed_cert_return_tls_key:?}"

log::info "Create Kyma CLI overrides"
envsubst < "$TEST_INFRA_SOURCES_DIR/prow/scripts/resources/kyma-installer-overrides.tpl.yaml" > "$PWD/kyma-installer-overrides.yaml"

log::info "Kyma installation triggered"

yes | kyma install \
    --ci \
    -s "$KYMA_SOURCE" \
    -o "$PWD/kyma-installer-overrides.yaml" \
    --domain "$DNS_SUBDOMAIN.${DNS_DOMAIN%.}" \
    --tls-cert="$TLS_CERT" \
    --tls-key="$TLS_KEY" \
    --timeout 60m

log::info "Patch serverless to enable containerd"
cp -va "${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/containerd-gke-patch.tpl.yaml" \
       "${KYMA_SOURCES_DIR}/resources/serverless/templates/containerd-gke-patch.yaml"
helm template -s templates/containerd-gke-patch.yaml resources/serverless/ --dry-run > "$PWD/patch-containerd.yaml"
kubectl apply -f "$PWD/patch-containerd.yaml"
# This command expects serverless to be running in the namespace `kyma-system`
# The `initialized` condition is checked because we only care about the init-container.
kubectl -n kyma-system wait pods -l app=serverless-docker-registry-cert-update --for="condition=Initialized"

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

log::info "Collect list of images"
# generate pod-security-policy list in json
utils::save_psp_list "$ARTIFACTS/kyma-psp.json"

utils::kubeaudit_create_report "$ARTIFACTS/kubeaudit.log"
utils::kubeaudit_check_report "$ARTIFACTS/kubeaudit.log"

log::info "Test Kyma"
# shellcheck disable=SC2031
# TODO (@Ressetkk): Kyma test functions as a separate library
"$TEST_INFRA_SOURCES_DIR"/prow/scripts/kyma-testing.sh "$INTEGRATION_TEST_LABEL_QUERY"

log::success "Integration Test successful"

#!!! Must be at the end of the script !!!
export ERROR_LOGGING_GUARD="false"
