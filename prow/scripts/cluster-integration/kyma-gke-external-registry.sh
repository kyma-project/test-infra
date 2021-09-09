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
export TEST_INFRA_SOURCES_DIR="$KYMA_PROJECT_DIR/test-infra"
export KYMA_SOURCES_DIR="$KYMA_PROJECT_DIR/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="$TEST_INFRA_SOURCES_DIR/prow/scripts/cluster-integration/helpers"

# shellcheck source=prow/scripts/lib/utils.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/docker.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"

KYMA_SCRIPTS_DIR="$KYMA_SOURCES_DIR/installation/scripts"
KYMA_RESOURCES_DIR="$KYMA_SOURCES_DIR/installation/resources"

INSTALLER_YAML="$KYMA_RESOURCES_DIR/installer.yaml"
INSTALLER_CR="$KYMA_RESOURCES_DIR/installer-cr-cluster.yaml.tpl"

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

# docker_cleanup runs at the end of a script or on any error
function docker_cleanup() {
    set +e
    if [ -n "$CLEANUP_DOCKER_IMAGE" ]; then
        log::info "Docker image cleanup"
        if [ -n "$KYMA_INSTALLER_IMAGE" ]; then
            log::info "Delete temporary Kyma-Installer Docker image"
            gcp::delete_docker_image \
                -i "$KYMA_INSTALLER_IMAGE"
        fi
    fi
    set -e
}

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

function create_image() {
#Description: Builds Kyma-Installer image from Kyma sources and pushes it to the repository
#
#Expected vars:
# - KYMA_SOURCES_DIR: directory with Kyma sources to build Kyma-Installer image
# - KYMA_INSTALLER_IMAGE: Full image name (with tag)
#
#Permissions: In order to run this script you need to use a service account with "Storage Admin" role

    set -o errexit

    requiredVars=(
        KYMA_SOURCES_DIR
        KYMA_INSTALLER_IMAGE
    )

    log::info "Build Kyma-Installer Docker image"
    CLEANUP_DOCKER_IMAGE="true"

    utils::check_required_vars "${requiredVars[@]}"

    echo "--------------------------------------------------------------------------------"
    echo "Building Kyma-Installer image: $KYMA_INSTALLER_IMAGE"
    echo "--------------------------------------------------------------------------------"
    echo
    docker build "$KYMA_SOURCES_DIR" -f "$KYMA_SOURCES_DIR"/tools/kyma-installer/kyma.Dockerfile -t "$KYMA_INSTALLER_IMAGE"

    echo "--------------------------------------------------------------------------------"
    echo "pushing Kyma-Installer image"
    echo "--------------------------------------------------------------------------------"
    echo
    docker push "$KYMA_INSTALLER_IMAGE"
    echo "--------------------------------------------------------------------------------"
    echo "Kyma-Installer image pushed: $KYMA_INSTALLER_IMAGE"
    echo "--------------------------------------------------------------------------------"
}

# Using set -f to prevent path globing in post_hook arguments.
# utils::post_hook call set +f at the beginning.
trap 'EXIT_STATUS=$?; docker_cleanup; set -f; utils::post_hook -n "$COMMON_NAME" -p "$CLOUDSDK_CORE_PROJECT" -c "$CLEANUP_CLUSTER" -g "$CLEANUP_GATEWAY_DNS_RECORD" -G "$INGRESS_GATEWAY_HOSTNAME" -a "$CLEANUP_APISERVER_DNS_RECORD" -A "$APISERVER_HOSTNAME" -I "$CLEANUP_GATEWAY_IP_ADDRESS" -l "$ERROR_LOGGING_GUARD" -z "$CLOUDSDK_COMPUTE_ZONE" -R "$CLOUDSDK_COMPUTE_REGION" -r "$PROVISION_REGIONAL_CLUSTER" -d "$DISABLE_ASYNC_DEPROVISION" -s "$COMMON_NAME" -e "$GATEWAY_IP_ADDRESS" -f "$APISERVER_IP_ADDRESS" -N "$COMMON_NAME" -Z "$CLOUDSDK_DNS_ZONE_NAME" -E "$EXIT_STATUS" -j "$JOB_NAME"' EXIT INT

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

if [ "$BUILD_TYPE" = "pr" ]; then
    KYMA_INSTALLER_IMAGE="$DOCKER_PUSH_REPOSITORY$DOCKER_PUSH_DIRECTORY/gke-external/$REPO_OWNER/$REPO_NAME:PR-$PULL_NUMBER"
elif [ "$BUILD_TYPE" != "release" ]; then
    KYMA_INSTALLER_IMAGE="$DOCKER_PUSH_REPOSITORY$DOCKER_PUSH_DIRECTORY/gke-external/$REPO_OWNER/$REPO_NAME:PR-${PULL_BASE_SHA::8}"
fi

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

gcp::authenticate \
    -c "$GOOGLE_APPLICATION_CREDENTIALS"

docker::start

if [[ "$BUILD_TYPE" != "release" ]]; then
    create_image
fi

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
    -e "$GKE_ENABLE_POD_SECURITY_POLICY" \
    -P "$TEST_INFRA_SOURCES_DIR"
export CLEANUP_CLUSTER="true"

utils::generate_self_signed_cert \
    -d "$DNS_DOMAIN" \
    -s "$COMMON_NAME" \
    -v "$SELF_SIGN_CERT_VALID_DAYS"
export TLS_CERT="${utils_generate_self_signed_cert_return_tls_cert:?}"
export TLS_KEY="${utils_generate_self_signed_cert_return_tls_key:?}"

log::info "Apply Kyma config"

kubectl create namespace "kyma-installer"

# TODO: convert create-config-map.sh to function in sourced lib script?
"$TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS/create-config-map.sh" --name "installation-config-overrides" \
    --data "global.domainName=$DNS_SUBDOMAIN.${DNS_DOMAIN%.}" \
    --data "global.loadBalancerIP=$GATEWAY_IP_ADDRESS"

"$TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS/create-config-map.sh" --name "core-test-ui-acceptance-overrides" \
    --data "test.acceptance.ui.logging.enabled=true" \
    --label "component=core"

"$TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS/create-config-map.sh" --name "application-registry-overrides" \
    --data "application-registry.deployment.args.detailedErrorResponse=true" \
    --label "component=application-connector"

"$TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS/create-config-map.sh" --name "cluster-certificate-overrides" \
    --data "global.tlsCrt=$TLS_CERT" \
    --data "global.tlsKey=$TLS_KEY"

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

# TODO: convert create-config-map.sh to function in sourced lib script?
"$TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS/create-config-map-file.sh" --name "istio-overrides" \
    --label "component=istio" \
    --file "$PWD/kyma_istio_operator"

DOCKER_PASSWORD=/tmp/kyma-gke-integration/dockerPassword.json
mkdir -p /tmp/kyma-gke-integration
< "$GOOGLE_APPLICATION_CREDENTIALS" tr -d '\n' > /tmp/kyma-gke-integration/dockerPassword.json
# TODO: convert create-config-map.sh to function in sourced lib script?
"$TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS/create-secret.sh" --name "serverless-external-registry-overrides" \
    --data "dockerRegistry.enableInternal=false" \
    --data "dockerRegistry.username=_json_key" \
    --file "dockerRegistry.password=$DOCKER_PASSWORD" \
    --data "dockerRegistry.serverAddress=$(echo "$DOCKER_PUSH_REPOSITORY" | cut -d'/' -f1)" \
    --data "dockerRegistry.registryAddress=$DOCKER_PUSH_REPOSITORY/functions" \
    --label "component=serverless"

if [[ "$BUILD_TYPE" == "release" ]]; then
    echo "Use released artifacts"
    gsutil cp "$KYMA_ARTIFACTS_BUCKET/$KYMA_SOURCE/kyma-installer-cluster.yaml" /tmp/kyma-gke-integration/downloaded-installer.yaml
    kubectl apply -f /tmp/kyma-gke-integration/downloaded-installer.yaml || true
    wait 2
    kubectl apply -f /tmp/kyma-gke-integration/downloaded-installer.yaml
else
    echo "Manual concatenating yamls"
    local installer_yaml=$("$KYMA_SCRIPTS_DIR"/concat-yamls.sh "$INSTALLER_YAML" "$INSTALLER_CR" \
    | sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: $KYMA_INSTALLER_IMAGE;" \
    | sed -e "s/__VERSION__/0.0.1/g" \
    | sed -e "s/__.*__//g")
    echo "$installer_yaml" | kubectl apply -f- || true
    wait 2
    echo "$installer_yaml" | kubectl apply -f- 
fi

log::info "Installation triggered"
"$KYMA_SCRIPTS_DIR"/is-installed.sh --timeout 30m

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
KYMA_TESTS="serverless serverless-long" "$TEST_INFRA_SOURCES_DIR/prow/scripts/kyma-testing.sh"

log::success "Success"

#!!! Must be at the end of the script !!!
# shellcheck disable=SC2034
ERROR_LOGGING_GUARD="false"
