#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on a real Gardener cluster.
#
#Expected common vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - GARDENER_REGION - Gardener compute region
# - GARDENER_ZONES - Gardener compute zones inside the region
# - GARDENER_CLUSTER_VERSION - Version of the Kubernetes cluster
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME - Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME - Name of the secret configured in the gardener project to access the cloud provider
# - MACHINE_TYPE - (optional) machine type
#
#Please look in each provider script for provider specific requirements




# exit on error, and raise error when variable is not set when used
set -e

ENABLE_TEST_LOG_COLLECTOR=false

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"

# All provides require these values, each of them may check for additional variables
requiredVars=(
    GARDENER_PROVIDER
    KYMA_PROJECT_DIR
    GARDENER_REGION
    GARDENER_ZONES
    GARDENER_CLUSTER_VERSION
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
)

utils::check_required_vars "${requiredVars[@]}"

if [[ $GARDENER_PROVIDER == "azure" ]]; then
    # shellcheck source=prow/scripts/lib/gardener/azure.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/azure.sh"
elif [[ $GARDENER_PROVIDER == "aws" ]]; then
    # shellcheck source=prow/scripts/lib/gardener/aws.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/aws.sh"
elif [[ $GARDENER_PROVIDER == "gcp" ]]; then
    # shellcheck source=prow/scripts/lib/gardener/gcp.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gcp.sh"
else
    ## TODO what should I put here? Is this a backend?
    log::error "GARDENER_PROVIDER ${GARDENER_PROVIDER} is not yet supported"
    exit 1
fi

# nice cleanup on exit, be it succesful or on fail
trap gardener::cleanup EXIT INT

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
readonly COMMON_NAME_PREFIX="grd"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
export COMMON_NAME

### Cluster name must be less than 10 characters!
export CLUSTER_NAME="${COMMON_NAME}"

# checks required vars and initializes gcloud/docker if necessary
gardener::init

# if MACHINE_TYPE is not set then use default one
gardener::set_machine_type

kyma::install_cli

# currently only Azure generates overrides, but this may change in the future
gardener::generate_overrides

gardener::provision_cluster

host::prepare_components_file(){
cat << EOF > "$PWD/components.yaml"
defaultNamespace: kyma-system
prerequisites:
  - name: "cluster-essentials"
  - name: "istio"
    namespace: "istio-system"
  - name: "certificates"
    namespace: "istio-system"
components:
  - name: "dex"
  - name: "ory"
  - name: "api-gateway"
EOF
}

log::info "Installing Kyma with Dex, Ory and API-Gateway"

if [[ "${KYMA_ALPHA}" == "true" ]]; then
  host::prepare_components_file
  kyma alpha deploy --ci --components-file "$PWD/components.yaml" --value global.isBEBEnabled=true --source=local --workspace "${KYMA_SOURCES_DIR}" --verbose
else
  host::prepare_components_file
  kyma install --ci --components "$PWD/components.yaml" --override global.isBEBEnabled=true --source=local --src-path "${KYMA_SOURCES_DIR}" --verbose
fi

# generate pod-security-policy list in json
utils::save_psp_list "${ARTIFACTS}/kyma-psp.json"

log::info "Prepare test environment variables"

# Preparing needed environment variables for API Gateway tests, these can be moved later on.
export TEST_HYDRA_ADDRESS="https://oauth2.${CLUSTER_NAME}.kyma-prow.shoot.canary.k8s-hana.ondemand.com"
TEST_USER_EMAIL="$(kubectl get secret -n kyma-system admin-user --template="{{.data.email}}" | base64 --decode)"
export TEST_USER_EMAIL
TEST_USER_PASSWORD="$(kubectl get secret -n kyma-system admin-user --template="{{.data.password}}" | base64 --decode)"
export TEST_USER_PASSWORD
export TEST_REQUEST_TIMEOUT="120"
export TEST_REQUEST_DELAY="10"
export INGRESSGATEWAY_ADDRESS="istio-ingressgateway.istio-system.svc.cluster.local"
export TEST_DOMAIN="${CLUSTER_NAME}.kyma-prow.shoot.canary.k8s-hana.ondemand.com"
export TEST_CLIENT_TIMEOUT=30s
export TEST_RETRY_MAX_ATTEMPTS="5"
export TEST_RETRY_DELAY="5"
export TEST_GATEWAY_NAME="kyma-gateway"
export TEST_GATEWAY_NAMESPACE="kyma-system"

log::info "Running Kyma API-Gateway tests"

pushd "${KYMA_SOURCES_DIR}/tests/integration/api-gateway/gateway-tests"
go test -v ./main_test.go
popd

log::success "Tests completed"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
