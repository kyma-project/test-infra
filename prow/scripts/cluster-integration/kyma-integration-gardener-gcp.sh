#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on a real Gardener cluster.
#
#
#Expected vars:
#
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - GARDENER_REGION - Gardener compute region
# - GARDENER_ZONES - Gardener compute zones inside the region
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME Name of the GCP secret configured in the gardener project to access the cloud provider
# - MACHINE_TYPE (optional): GCP machine type
#
#Permissions: In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
# - Compute Admin
# - Kubernetes Engine Admin
# - Kubernetes Engine Cluster Admin
# - DNS Administrator
# - Service Account User
# - Storage Admin
# - Compute Network Admin


set -e

discoverUnsetVar=false

VARIABLES=(
    KYMA_PROJECT_DIR
    GARDENER_REGION
    GARDENER_ZONES
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
)

for var in "${VARIABLES[@]}"; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi
readonly GARDENER_CLUSTER_VERSION="1.16.3"

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
        # Export envvars for the script
        export GARDENER_CLUSTER_NAME=${CLUSTER_NAME}
        export GARDENER_PROJECT_NAME=${GARDENER_KYMA_PROW_PROJECT_NAME}
        export GARDENER_CREDENTIALS=${GARDENER_KYMA_PROW_KUBECONFIG}
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/deprovision-gardener-cluster.sh
    fi

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

# Local variables
# DNS_SUBDOMAIN="${COMMON_NAME}"

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

export INSTALL_DIR=${TMP_DIR}
install::kyma_cli

shout "Provision cluster: \"${CLUSTER_NAME}\""

if [ -z "$MACHINE_TYPE" ]; then
      export MACHINE_TYPE="n1-standard-4"
fi

CLEANUP_CLUSTER="true"
(
set -x
kyma provision gardener \
        --target-provider gcp --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" \
        --name "${CLUSTER_NAME}" --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
        --region "${GARDENER_REGION}" -z "${GARDENER_ZONES}" -t "${MACHINE_TYPE}" --nodes 4 --kube-version=${GARDENER_CLUSTER_VERSION} \
)

shout "Installing Kyma"
date

echo "Downlading production profile"
curl -L --silent --fail --show-error "https://raw.githubusercontent.com/kyma-project/kyma/master/installation/resources/installer-config-production.yaml.tpl" \
    --output installer-config-production.yaml.tpl

(
set -x
kyma install \
    --ci \
    --source latest \
    -o installer-config-production.yaml.tpl \
    --timeout 90m
)

shout "Checking the versions"
date
kyma version

shout "Running Kyma tests"
date

readonly SUITE_NAME="testsuite-all-$(date '+%Y-%m-%d-%H-%M')"
readonly CONCURRENCY=5
(
set -x
kyma test run \
    --name "${SUITE_NAME}" \
    --concurrency "${CONCURRENCY}" \
    --max-retries 1 \
    --timeout 90m \
    --watch \
    --non-interactive
)

echo "Test Summary"
kyma test status "${SUITE_NAME}" -owide

statusSucceeded=$(kubectl get cts "${SUITE_NAME}" -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
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
