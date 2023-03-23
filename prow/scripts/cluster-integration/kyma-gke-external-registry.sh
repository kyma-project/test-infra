#!/usr/bin/env bash

#Description: Kyma Integration plan on GKE with external registry. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma on real GKE cluster.
#
#
#Expected vars:
#
# - REPO_OWNER - Set up by prow, repository owner/organization
# - REPO_NAME - Set up by prow, repository name
# - BUILD_TYPE - Set up by prow, pr/master/release
# - DOCKER_PUSH_REPOSITORY - Docker repository url
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
export TEST_INFRA_SOURCES_DIR="$KYMA_PROJECT_DIR/test-infra"
export KYMA_SOURCES_DIR="$KYMA_PROJECT_DIR/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="$TEST_INFRA_SOURCES_DIR/prow/scripts/cluster-integration/helpers"

# shellcheck source=prow/scripts/lib/utils.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/kyma.sh"

# Enforce lowercase
readonly REPO_OWNER=${REPO_OWNER,,}
export REPO_OWNER
# Enforce lowercase
readonly REPO_NAME=${REPO_NAME,,}
export REPO_NAME
export INGRESS_GATEWAY_HOSTNAME='*'
export APISERVER_HOSTNAME='apiserver'

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
)

utils::check_required_vars "${requiredVars[@]}"

verify_internal_registry() {
    local pods
    pods="$(kubectl get pods --all-namespaces | grep docker-registry || true)"

    if [[ -n "$pods" ]]; then
        echo "Pods with docker registry are running:"
        echo "$pods"
        return 1
    fi

    return 0
}

# Using set -f to prevent path globing in post_hook arguments.
# utils::post_hook call set +f at the beginning.
# shellcheck disable=SC2153
trap 'EXIT_STATUS=$?; set -f; utils::post_hook -k "true" -n "$COMMON_NAME" -p "$CLOUDSDK_CORE_PROJECT" -c "$CLEANUP_CLUSTER" -g "$CLEANUP_GATEWAY_DNS_RECORD" -G "$INGRESS_GATEWAY_HOSTNAME" -a "$CLEANUP_APISERVER_DNS_RECORD" -A "$APISERVER_HOSTNAME" -I "$CLEANUP_GATEWAY_IP_ADDRESS" -l "$ERROR_LOGGING_GUARD" -z "$CLOUDSDK_COMPUTE_ZONE" -R "$CLOUDSDK_COMPUTE_REGION" -r "$PROVISION_REGIONAL_CLUSTER" -d "$DISABLE_ASYNC_DEPROVISION" -s "$COMMON_NAME" -e "$GATEWAY_IP_ADDRESS" -f "$APISERVER_IP_ADDRESS" -N "$COMMON_NAME" -Z "$CLOUDSDK_DNS_ZONE_NAME" -E "$EXIT_STATUS" -j "$JOB_NAME"' EXIT INT

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

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

gcp::authenticate \
    -c "$GOOGLE_APPLICATION_CREDENTIALS"

gcp::set_vars_for_network \
    -n "$JOB_NAME"
export GCLOUD_NETWORK_NAME="${gcp_set_vars_for_network_return_net_name:?}"
export GCLOUD_SUBNET_NAME="${gcp_set_vars_for_network_return_subnet_name:?}"

gcp::create_network \
    -n "$GCLOUD_NETWORK_NAME" \
    -s "$GCLOUD_SUBNET_NAME" \
    -p "$CLOUDSDK_CORE_PROJECT"

log::info "Reserve IP Address for Ingressgateway"
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

# Prepare Docker external registry overrides
export DOCKER_PASSWORD=""
DOCKER_PASSWORD=$(tr -d '\n' < "${GOOGLE_APPLICATION_CREDENTIALS}")

export DOCKER_REPOSITORY_ADDRESS=""
DOCKER_REPOSITORY_ADDRESS=$(echo "$DOCKER_PUSH_REPOSITORY" | cut -d'/' -f1)

export DNS_DOMAIN_TRAILING=${DNS_DOMAIN%.}
envsubst < "${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/kyma-serverless-external-registry-integration-overrides.tpl.yaml" > "$PWD/kyma_overrides.yaml"

log::info "Installation triggered"

# Install unstable cli because released one is not working without following changes:
# https://github.com/kyma-project/cli/pull/1398
kyma::install_unstable_cli

kyma deploy --ci --source=local --workspace "$KYMA_SOURCES_DIR" --verbose \
    --values-file "$PWD/kyma_overrides.yaml"

if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
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

log::info "Verify if internal docker registry is disabled"
verify_internal_registry

log::info "Test Kyma"
# Istio sidecar will keep running and will not terminate when the test pod is completed. This causes the test job to
# be stuck forever. So, we disable the sidecar for the test job pod. Kyma runs with istio mtls STRICT peer authentication. 
# The STRICT mode will block any traffic comming from pods without istio sidecar. SO, we need to make it PERMISSIVE.
kubectl patch peerauthentication -n istio-system default -p '{"spec":{"mtls":{"mode":"PERMISSIVE"}}}' --type=merge

SERVERLESS_CHART_DIR="${KYMA_SOURCES_DIR}/resources/serverless"
job_name="k3s-serverless-test"

helm install serverless-test "${SERVERLESS_CHART_DIR}/charts/k3s-tests" -n default -f "${SERVERLESS_CHART_DIR}/values.yaml" --set jobName="${job_name}"
kubectl patch job -n default k3s-serverless-test -p '{"metadata":{"annotations":{"sidecar.istio.io/inject":"false"}}}'
job_status=1

# helm does not wait for jobs to complete even with --wait
# TODO but helm@v3.5 has a flag that enables that, get rid of this function once we use helm@v3.5
getjobstatus(){
while true; do
    echo "Test job not completed yet..."
    [[ $(kubectl get jobs $job_name -o jsonpath='{.status.conditions[?(@.type=="Failed")].status}') == "True" ]] && job_status=1 && echo "Test job failed" && break
    [[ $(kubectl get jobs $job_name -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}') == "True" ]] && job_status=0 && echo "Test job completed successfully" && break
    sleep 5
done
}

getjobstatus

echo "####################"
echo "kubectl get pods -A"
echo "###################"
kubectl get pods -A

echo "########################"
echo "kubectl get functions -A"
echo "########################"
kubectl get functions -A

echo "########################################################"
echo "kubectl logs -n kyma-system -l app=serverless --tail=-1"
kubectl logs -n kyma-system -l app=serverless --tail=-1

echo "##############################################"
echo "kubectl logs -l job-name=${job_name} --tail=-1"
kubectl logs -l job-name=${job_name} --tail=-1
echo "###############"
echo ""

echo "Exit code ${job_status}"
if [ "${job_status}" != 0 ]; then
    exit 1
fi

#!!! Must be at the end of the script !!!
# shellcheck disable=SC2034
ERROR_LOGGING_GUARD="false"
