#!/bin/bash

# This script is designed to run SKR integration test with own_cluster plan.

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
  log::info "Execute Job Guard"
  "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

export SHOOT_DEFINITION_PATH="${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/skr-test/shoot-template.yaml"
export GARDENER_SHOOT_TEMPLATE=$(envsubst < "$SHOOT_DEFINITION_PATH" | base64)
export GARDENER_KUBECONFIG=$GARDENER_KYMA_PROW_KUBECONFIG
export KCP_GARDENER_NAMESPACE="garden-$GARDENER_KYMA_PROW_PROJECT_NAME"

pushd /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration

make ci-skr-own-cluster

log::success "all done"
