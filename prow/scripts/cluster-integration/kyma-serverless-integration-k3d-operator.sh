#!/usr/bin/env bash

set -o errexit
set -o pipefail # Fail a pipe if any sub-command fails.

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=prow/scripts/cluster-integration/kyma-serverless-run-test.sh
source "${SCRIPT_DIR}/kyma-serverless-run-test.sh"
# shellcheck source=prow/scripts/lib/serverless-shared-k3s.sh
source "${SCRIPT_DIR}/../lib/serverless-shared-k3s.sh"

date

export DOMAIN=${KYMA_DOMAIN:-local.kyma.dev}
if [[ -z $REGISTRY_VALUES ]]; then
  export REGISTRY_VALUES="dockerRegistry.enableInternal=false,dockerRegistry.serverAddress=registry.localhost:5000,dockerRegistry.registryAddress=registry.localhost:5000"
fi

export INTEGRATION_SUITE=("$@")

make -C $KYMA_SOURCES_DIR/hack/ci run-without-lifecycle-manager-operator

run_tests "${INTEGRATION_SUITE[@]}"
TEST_STATUS=$?
set -o errexit

collect_results
echo "Exit code ${TEST_STATUS}"

exit ${TEST_STATUS}
