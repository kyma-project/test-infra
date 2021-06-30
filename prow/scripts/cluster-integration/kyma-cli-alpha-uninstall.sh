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
# - GARDENER_CLUSTER_VERSION - Version of the Kubernetes cluster
# - MACHINE_TYPE (optional): GCP machine type
#
#Permissions: In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
# - Compute Admin
# - Service Account User
# - Service Account Admin
# - Service Account Token Creator
# - Make sure the service account is enabled for the Google Identity and Access Management API.

set -e

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/testing-helpers.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gardener/gcp.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gcp.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"

requiredVars=(
    KYMA_PROJECT_DIR
    GARDENER_REGION
    GARDENER_ZONES
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
    GARDENER_CLUSTER_VERSION
)

utils::check_required_vars "${requiredVars[@]}"

# nice cleanup on exit, be it succesful or on fail
trap gardener::cleanup EXIT INT

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c4)
readonly COMMON_NAME_PREFIX="grdnr"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")

### Cluster name must be less than 10 characters!
export CLUSTER_NAME="${COMMON_NAME}"

# Local variables

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

log::info "Building Kyma CLI"
cd "${KYMA_PROJECT_DIR}/cli"
make build-linux
mv "${KYMA_PROJECT_DIR}/cli/bin/kyma-linux" "${KYMA_PROJECT_DIR}/cli/bin/kyma"
export PATH="${KYMA_PROJECT_DIR}/cli/bin:${PATH}"

log::info "Provision cluster: \"${CLUSTER_NAME}\""

gardener::set_machine_type

gardener::provision_cluster

log::info "Installing Kyma"

(
cd "${KYMA_PROJECT_DIR}/kyma"
kyma::alpha_deploy_kyma
)

sleep 1m

log::info "Uninstall Kyma"
(
cd "${KYMA_PROJECT_DIR}/kyma"
kyma alpha delete --ci
)

sleep 1m

log::info "Install Kyma again"
(
cd "${KYMA_PROJECT_DIR}/kyma"
kyma::alpha_deploy_kyma
)

log::info "Run Kyma tests"
SUITE_NAME="testsuite-alpha-$(date '+%Y-%m-%d-%H-%M')"
export SUITE_NAME
set +e
(
cd "${KYMA_PROJECT_DIR}/kyma"
kyma test run \
    --name "$SUITE_NAME" \
    --concurrency 6 \
    --max-retries 1 \
    --timeout 60m \
    --watch \
    --non-interactive \
    istio-kyma-validate application-connector application-operator application-registry \
    connection-token-handler connector-service api-gateway \
    kiali logging monitoring \
    rafter serverless serverless-long service-catalog
)

# collect logs from failed tests before deprovisioning
kyma::run_test_log_collector "kyma-cli-alpha-uninstall-gke"

if ! kyma::test_summary; then
    log::error "Tests have failed"
    set -e
    exit 1
fi
set -e

log::info "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
