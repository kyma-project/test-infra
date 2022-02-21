#!/usr/bin/env bash

set -e
set -o pipefail

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

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration"

export CONTROL_PLANE_RECONCILER_DIR="/home/prow/go/src/github.com/kyma-project/control-plane/tools/reconciler"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/gardener/gardener.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gardener.sh"
# shellcheck source=prow/scripts/cluster-integration/helpers/reconciler.sh
source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/helpers/reconciler.sh"


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

export INPUT_CLUSTER_NAME="rec-wkly-lt"

# Provisioning gardener long lasting cluster
reconciler::provision_cluster

reconciler::export_shoot_cluster_kubeconfig

set +e
log::banner "Deploying Monitoring for load test"
git clone https://github.com/prometheus-operator/kube-prometheus.git
cd kube-prometheus
kubectl create -f manifests/setup
kubectl create -f manifests/
set -e

log::banner "Deploying Reconciler for load test"
cd "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"  || { echo "Failed to change dir to: ${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"; exit 1; }

mothership_latest_commit=$(curl  --silent "https://api.github.com/repos/kyma-incubator/reconciler/commits/main" | jq -r '.sha')
mothership_tag="${mothership_latest_commit::8}"
mothership_tag="PR-848"
mock_component_tag="PR-838"
sed -i "s/reconciler\/mothership:.\\{8\\}/reconciler\/mothership:${mothership_tag}/g" ./resources/reconciler-load-test.yaml
sed -i "s/reconciler\/component:.\\{8\\}/reconciler\/component:${mock_component_tag}/g" ./resources/reconciler-load-test.yaml

sed -i "s/namespace:  reconciler/namespace:  monitoring/g" ./resources/reconciler-load-test.yaml

echo "************* Current reconciler Image to be used **************"
cat ./resources/reconciler-load-test.yaml | grep -o 'reconciler\/mothership:.*'
cat ./resources/reconciler-load-test.yaml | grep -o 'reconciler\/component:.*' | head -1
echo "**************************************************************"

kubectl apply -f resources/reconciler-load-test.yaml

