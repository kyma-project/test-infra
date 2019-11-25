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
# - MACHINE_TYPE (optional): GKE machine type
# - CLUSTER_VERSION (optional): GKE cluster version
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

discoverUnsetVar=false

for var in KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

readonly SUITE_NAME="testsuite-all-$(date '+%Y-%m-%d-%H-%M')"
readonly CONCURRENCY=5
readonly TMP_DIR=$(mktemp -d)

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/testing-helpers.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/kyma-testing.sh"

#!Put cleanup code in this function! Function is executed at exit from the script and on interuption.
cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        shout "AN ERROR OCCURED! Take a look at preceding log entries."
        echo
    fi

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    # if [ -n "${CLEANUP_CLUSTER}" ]; then
    #     shout "Deprovision cluster: \"${CLUSTER_NAME}\""
    #     date

    #     #save disk names while the cluster still exists to remove them later
    #     DISKS=$(kubectl get pvc --all-namespaces -o jsonpath="{.items[*].spec.volumeName}" | xargs -n1 echo)
    #     export DISKS

    #     #Delete cluster
    #     "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/deprovision-gke-cluster.sh"

    #     #Delete orphaned disks
    #     shout "Delete orphaned PVC disks..."
    #     date
    #     "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/delete-disks.sh"
    # fi

    # if [ -n "${CLEANUP_GATEWAY_DNS_RECORD}" ]; then
    #     shout "Delete Gateway DNS Record"
    #     date
    #     "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${GATEWAY_DNS_FULL_NAME}" --address="${GATEWAY_IP_ADDRESS}" --dryRun=false
    # fi

    # if [ -n "${CLEANUP_APISERVER_DNS_RECORD}" ]; then
    #     shout "Delete Apiserver proxy DNS Record"
    #     date
    #     "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${APISERVER_DNS_FULL_NAME}" --address="${APISERVER_IP_ADDRESS}" --dryRun=false
    # fi

    rm -rf "${TMP_DIR}"

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    shout "Job is finished ${MSG}"
    date
    set -e

    exit "${EXIT_STATUS}"
}

trap cleanup EXIT INT

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c10)
readonly COMMON_NAME_PREFIX="cli-integration-test-gke"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")

### Cluster name must be less than 40 characters!
export CLUSTER_NAME="${COMMON_NAME}"

export GCLOUD_NETWORK_NAME="${COMMON_NAME_PREFIX}-net"
export GCLOUD_SUBNET_NAME="${COMMON_NAME_PREFIX}-subnet"

### For provision-gke-cluster.sh
export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"

#Local variables
DNS_SUBDOMAIN="${COMMON_NAME}"

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"



export INSTALL_DIR=${TMP_DIR}
install::kyma_cli


shout "Provision cluster: \"${CLUSTER_NAME}\""
kyma provision gardener --target-provider az --name "${CLUSTER_NAME}"


# shout "Authenticate"
# date
# init

# DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"


# NETWORK_EXISTS=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/network-exists.sh")
# if [ "$NETWORK_EXISTS" -gt 0 ]; then
#     shout "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
#     date
#     "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-network-with-subnet.sh"
# else
#     shout "Network ${GCLOUD_NETWORK_NAME} exists"
# fi


# shout "Provision cluster: \"${CLUSTER_NAME}\""
# date
# export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
# if [ -z "$MACHINE_TYPE" ]; then
#       export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
# fi
# if [ -z "${CLUSTER_VERSION}" ]; then
#       export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
# fi
# CLEANUP_CLUSTER="true"
# "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/provision-gke-cluster.sh"


shout "Generate self-signed certificate"
date
DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
export DOMAIN
CERT_KEY=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-self-signed-cert.sh")
TLS_CERT=$(echo "${CERT_KEY}" | head -1)
TLS_KEY=$(echo "${CERT_KEY}" | tail -1)



shout "Installing Kyma"
date
yes | kyma install --non-interactive --source latest --domain "${DOMAIN}" --tlsCert "${TLS_CERT}" --tlsKey "${TLS_KEY}"


shout "Checking the versions"
date
kyma version


# if [ -n "$(kubectl get service -n istio-system istio-ingressgateway --ignore-not-found)" ]; then
#     shout "Create DNS Record for Ingressgateway IP"
#     date
#     GATEWAY_IP_ADDRESS=$(kubectl get service -n istio-system istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
#     GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
#     CLEANUP_GATEWAY_DNS_RECORD="true"
#     IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
# fi

# if [ -n "$(kubectl get  service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
#     shout "Create DNS Record for Apiserver proxy IP"
#     date
#     APISERVER_IP_ADDRESS=$(kubectl get  service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
#     APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
#     CLEANUP_APISERVER_DNS_RECORD="true"
#     IP_ADDRESS=${APISERVER_IP_ADDRESS} DNS_FULL_NAME=${APISERVER_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
# fi


shout "Running Kyma tests"
date
echo "- Creating ClusterAddonsConfiguration which provides the testing addons"
injectTestingAddons
if [[ $? -eq 1 ]]; then
    exit 1
fi

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


shout "Uninstalling Kyma"
date
kyma uninstall --non-interactive


shout "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
