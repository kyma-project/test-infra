#!/usr/bin/env bash

set -e
set -o pipefail

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on a real Gardener cluster.
#Please look in each provider script for provider specific requirements

# THIS SCRIPT WILL START FROM ROOT OF RECONCILER

# shellcheck source=prow/scripts/reconciler/common.sh
source "../../kyma-project/test-infra/prow/scripts/reconciler/common.sh"

# All provides require these values, each of them may check for additional variables
requiredVars=(
    GARDENER_REGION
    GARDENER_ZONES
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
    INPUT_CLUSTER_NAME
)

utils::check_required_vars "${requiredVars[@]}"

if  ! command -v helm &> /dev/null ; then
  echo "helm not found"
  mkdir -p bin
  HELM_VERSION=v3.12.1
  curl -Lo helm.tar.gz "https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz" && \
  tar -xzOf helm.tar.gz linux-amd64/helm > bin/helm && \
  chmod +x bin/helm
  export PATH="$PWD/bin:$PATH"
fi
reconciler::delete_cluster_if_exists

reconciler::export_nightly_cluster_name

reconciler::provision_cluster

reconciler::export_shoot_cluster_kubeconfig

reconciler::deploy

reconciler::disable_sidecar_injection_reconciler_ns

reconciler::wait_until_is_ready