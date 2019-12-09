#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on GKE. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on real GKE cluster.
#
#
#Expected vars:
#
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - GARDENER_REGION - Gardener compute region
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_AZURE_SECRET_NAME Name of the azure secret configured in the gardener project to access the cloud provider
# - MACHINE_TYPE (optional): AKS machine type
# - CLUSTER_VERSION (optional): AKS Kubernetes version TODO
#
#Permissions: In order to run this script you need to use an AKS service account with the contributor role

set -o errexit

discoverUnsetVar=false

for var in KYMA_PROJECT_DIR GARDENER_REGION GARDENER_KYMA_PROW_KUBECONFIG GARDENER_KYMA_PROW_PROJECT_NAME GARDENER_KYMA_PROW_AZURE_SECRET_NAME; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/testing-helpers.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/kyma-cli.sh"

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

    if [ -n "${CLEANUP_CLUSTER}" ]; then
        shout "Deprovision cluster: \"${CLUSTER_NAME}\""
        date
        #Delete cluster
        # Export envvars for the script
        export GARDENER_PROJECT_NAME = ${CLUSTER_NAME}
        export GARDENER_CLUSTER_NAME = ${GARDENER_KYMA_PROW_PROJECT_NAME}
        export GARDENER_CREDENTIALS = ${GARDENER_KYMA_PROW_KUBECONFIG}
        ${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/deprovision-gardener-cluster.sh
    fi

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

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
readonly COMMON_NAME_PREFIX="grdnr"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")

### Cluster name must be less than 20 characters!
export CLUSTER_NAME="${COMMON_NAME}"

#Local variables
DNS_SUBDOMAIN="${COMMON_NAME}"

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"



export INSTALL_DIR=${TMP_DIR}
install::kyma_cli


shout "Provision cluster: \"${CLUSTER_NAME}\""

if [ -z "$MACHINE_TYPE" ]; then
      export MACHINE_TYPE="Standard_D2_v3"
fi

CLEANUP_CLUSTER="true"
kyma provision gardener \
        --target-provider azure --secret ${GARDENER_KYMA_PROW_AZURE_SECRET_NAME} \
        --name "${CLUSTER_NAME}" --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
        --region "${GARDENER_REGION}" -t "${MACHINE_TYPE}" --disk-size 35 --disk-type=Standard_LRS --extra vnetcidr="10.250.0.0/19" \

# shout "Generate self-signed certificate"
# date
# DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
# export DOMAIN
# CERT_KEY=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-self-signed-cert.sh")
# TLS_CERT=$(echo "${CERT_KEY}" | head -1)
# TLS_KEY=$(echo "${CERT_KEY}" | tail -1)



shout "Installing Kyma"
date
# yes | kyma install --non-interactive --source latest --timeout=2h #--domain "${DOMAIN}" --tlsCert "${TLS_CERT}" --tlsKey "${TLS_KEY}"


shout "Checking the versions"
date
kyma version


# if [ -n "$(kubectl get service -n istio-system istio-ingressgateway --ignore-not-found)" ]; then
#     shout "Create DNS Record for Ingressgateway IP"
#     date
#     GATEWAY_IP_ADDRESS=$(kubectl get service -n istio-system istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
#     GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}" # TODO add a proper domain
#     CLEANUP_GATEWAY_DNS_RECORD="true"
#     IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
# fi

# if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
#     shout "Create DNS Record for Apiserver proxy IP"
#     date
#     APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
#     APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}" # TODO add proper domain
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

# readonly CONCURRENCY=5
# kyma test run \
#                 --name "${SUITE_NAME}" \
#                 --concurrency "${CONCURRENCY}" \
#                 --max-retries 1 \
#                 --timeout "1h" \
#                 --watch \
#                 --non-interactive


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

shout "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"