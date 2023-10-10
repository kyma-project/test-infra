#!/usr/bin/env bash

set -o errexit
set -o pipefail # Fail a pipe if any sub-command fails.

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=prow/scripts/cluster-integration/kyma-serverless-run-test.sh
source "${SCRIPT_DIR}/kyma-serverless-run-test.sh"
# shellcheck source=prow/scripts/lib/serverless-shared-k3s.sh
source "${SCRIPT_DIR}/../lib/serverless-shared-k3s.sh"

date

make -C "$SERVERLESS_SOURCES"/hack/ci run-without-lifecycle-manager

export INTEGRATION_SUITE=("$@")
run_tests "${INTEGRATION_SUITE[@]}"
TEST_STATUS=$?
set -o errexit

collect_results
echo "Exit code ${TEST_STATUS}"

exit ${TEST_STATUS}
