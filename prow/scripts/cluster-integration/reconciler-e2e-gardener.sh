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
export RECONCILER_SOURCES_DIR="/home/prow/go/src/github.com/kyma-incubator/reconciler"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
#set to dummy value, required by gardener/gcp.sh
export KYMA_SOURCE="main"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/gardener/gardener.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gardener.sh"

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

readonly COMMON_NAME_PREFIX="grd"
utils::generate_commonName -n "${COMMON_NAME_PREFIX}"
COMMON_NAME="${utils_generate_commonName_return_commonName:?}"
export COMMON_NAME

export CLUSTER_NAME="${COMMON_NAME}"

# checks required vars and initializes gcloud/docker if necessary
gardener::init

# if MACHINE_TYPE is not set then use default one
gardener::set_machine_type

kyma::install_cli

# currently only Azure generates overrides, but this may change in the future
gardener::generate_overrides

gardener::provision_cluster

log::info "Building Reconciler CLI"
date
cd "${RECONCILER_SOURCES_DIR}"
make deploy

# Wait until reconciler is ready
timeout=1200 # in secs
delay=10 # in secs
iterationsLeft=$(( timeout/delay ))


# Run a test pod
kubectl run -n reconciler --image=alpine:3.14.1 --restart=Never test-pod -- sh -c "sleep 36000"

# Wait until test pod is ready
timeout=10 # in secs
delay=2 # in secs
iterationsLeft=$(( timeout/delay ))
while : ; do
  testPodStatus=$(kubectl get po -n reconciler test-pod -ojsonpath='{.status.containerStatuses[*].ready}')
  if [ "${testPodStatus}" = "true" ]; then
    echo "Test pod is ready"
    break
  fi
  if [ "$timeout" -ne 0 ] && [ "$iterationsLeft" -le 0 ]; then
    echo "Timeout reached on initializing test pod. Exiting"
    exit 1
  fi
  echo "Waiting for test pod to be ready..."
  sleep 2
  iterationsLeft=$(( iterationsLeft-1 ))
done

# Copy the payload with kubeconfig to the test pod
# shellcheck disable=SC2016
# shellcheck source=/dev/null
kc="$(cat ${KUBECONFIG})"; jq --arg kubeconfig "${kc}" '.kubeconfig = $kubeconfig' ./scripts/e2e-test/template.json > body.json
kubectl cp body.json reconciler/test-pod:/tmp
kubectl cp  ./scripts/e2e-test/reconcile-kyma.sh reconciler/test-pod:/tmp

log::banner "Reconcile Kyma in the same cluster until it is ready"
kubectl exec -it -n reconciler test-pod -- sh -c ". /tmp/reconcile-kyma.sh"

### Once Kyma is installed run the fast integration test
log::banner "Execute tests"
gardener::test_fast_integration_kyma
#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"