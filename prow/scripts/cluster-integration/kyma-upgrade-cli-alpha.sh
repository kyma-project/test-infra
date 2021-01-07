#!/usr/bin/env bash

#Description: TEMPORARY PIPELINE FOR ALPHA FEATURES TESTING. WORK IN PROGRESS. Related issue: https://github.com/kyma-project/test-infra/issues/3057
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
# - Service Account User
# - Service Account Admin
# - Service Account Token Creator
# - Make sure the service account is enabled for the Google Identity and Access Management API.

set -e

ENABLE_TEST_LOG_COLLECTOR=false

readonly GARDENER_CLUSTER_VERSION="1.16"

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/testing-helpers.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

requiredVars=(
    KYMA_PROJECT_DIR
    GARDENER_REGION
    GARDENER_ZONES
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
)

utils::check_required_vars "${requiredVars[@]}"

#!Put cleanup code in this function! Function is executed at exit from the script and on interuption.
cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?
    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    if [[ -n "${SUITE_NAME}" ]]; then
        testing::test_summary
    fi 

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        shout "AN ERROR OCCURED! Take a look at preceding log entries."
        echo
    fi

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

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c4)
readonly COMMON_NAME_PREFIX="grdnr"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")

### Cluster name must be less than 10 characters!
export CLUSTER_NAME="${COMMON_NAME}"

# Local variables

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

shout "Building Kyma CLI"
date
cd "${KYMA_PROJECT_DIR}/cli"
make build-linux
mv "${KYMA_PROJECT_DIR}/cli/bin/kyma-linux" "${KYMA_PROJECT_DIR}/cli/bin/kyma"
export PATH="${KYMA_PROJECT_DIR}/cli/bin:${PATH}"

shout "Provision cluster: \"${CLUSTER_NAME}\""

if [ -z "$MACHINE_TYPE" ]; then
      export MACHINE_TYPE="n1-standard-4"
fi

CLEANUP_CLUSTER="true"
(
set -x
kyma provision gardener gcp \
        --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" --name "${CLUSTER_NAME}" \
        --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
        --region "${GARDENER_REGION}" -z "${GARDENER_ZONES}" -t "${MACHINE_TYPE}" \
        --scaler-max 4 --scaler-min 2 \
        --kube-version=${GARDENER_CLUSTER_VERSION}
)

shout "Installing Kyma"
date

# Parallel-install library installs cluster-essentials, istio, and xip-patch before kyma installation. That's why they should not exist on the InstallationCR.
# Once we figure out a way to fix this, this custom CR can be deleted from this script.
cat << EOF > "$PWD/kyma-parallel-install-installationCR.yaml"
apiVersion: "installer.kyma-project.io/v1alpha1"
kind: Installation
metadata:
  name: kyma-installation
  namespace: default
spec:
  components:
    - name: "testing"
      namespace: "kyma-system"
    - name: "knative-eventing"
      namespace: "knative-eventing"
    - name: "dex"
      namespace: "kyma-system"
    - name: "ory"
      namespace: "kyma-system"
    - name: "api-gateway"
      namespace: "kyma-system"
    - name: "rafter"
      namespace: "kyma-system"
    - name: "service-catalog"
      namespace: "kyma-system"
    - name: "service-catalog-addons"
      namespace: "kyma-system"
    - name: "helm-broker"
      namespace: "kyma-system"
    - name: "nats-streaming"
      namespace: "natss"
    - name: "core"
      namespace: "kyma-system"
    - name: "cluster-users"
      namespace: "kyma-system"
    - name: "logging"
      namespace: "kyma-system"
    - name: "permission-controller"
      namespace: "kyma-system"
    - name: "apiserver-proxy"
      namespace: "kyma-system"
    - name: "iam-kubeconfig-service"
      namespace: "kyma-system"
    - name: "serverless"
      namespace: "kyma-system"
    - name: "knative-provisioner-natss"
      namespace: "knative-eventing"
    - name: "event-sources"
      namespace: "kyma-system"
    - name: "application-connector"
      namespace: "kyma-integration"
    - name: "tracing"
      namespace: "kyma-system"
    - name: "monitoring"
      namespace: "kyma-system"
    - name: "kiali"
      namespace: "kyma-system"
    - name: "console"
      namespace: "kyma-system"
EOF

(
set -x
kyma alpha deploy \
    --ci \
    --resources "${KYMA_PROJECT_DIR}/kyma/resources" \
    --components "$PWD/kyma-parallel-install-installationCR.yaml"
)

shout "Checking the versions"
date
kyma version

# shout "Running Kyma tests"
# date

# # enable test-log-collector before tests; if prowjob fails before test phase we do not have any reason to enable it earlier
# if [[ "${BUILD_TYPE}" == "master" && -n "${LOG_COLLECTOR_SLACK_TOKEN}" ]]; then
#   export ENABLE_TEST_LOG_COLLECTOR=true
# fi

# readonly SUITE_NAME="testsuite-all-$(date '+%Y-%m-%d-%H-%M')"
# readonly CONCURRENCY=5
# (
# set -x
# kyma test run \
#     --name "${SUITE_NAME}" \
#     --concurrency "${CONCURRENCY}" \
#     --max-retries 1 \
#     --timeout 90m \
#     --watch \
#     --non-interactive
# )

shout "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
