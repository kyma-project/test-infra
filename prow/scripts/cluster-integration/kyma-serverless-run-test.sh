#!/usr/bin/env bash

set -o errexit
set -o pipefail # Fail a pipe if any sub-command fails.

run_tests() {
  # shellcheck source=prow/scripts/lib/serverless-shared-k3s.sh
  source "${SCRIPT_DIR}/../lib/serverless-shared-k3s.sh"

  export KYMA_SOURCES_DIR="./kyma"
  export SERVERLESS_OVERRIDES_DIR="./overrides"
  export INTEGRATION_SUITE=${1:-serverless-integration}

  date

  if [[ ${INTEGRATION_SUITE} == "git-auth-integration" ]]; then
    echo "--> Cloning Serverless integration tests from kyma:main"
    git clone https://github.com/kyma-project/kyma "${KYMA_SOURCES_DIR}"
  fi

  # check for test secrets.
  if [[ -e "${SERVERLESS_OVERRIDES_DIR}/git-auth.env" ]]; then
    # shellcheck source=/dev/null
    source "${SERVERLESS_OVERRIDES_DIR}/git-auth.env"
  fi

  #https://github.com/kyma-project/test-infra/issues/6513
  export PATH=${PATH}:/usr/local/go/bin

  export APP_TEST_CLEANUP="onSuccessOnly"
  set +o errexit
  (cd ${KYMA_SOURCES_DIR}/tests/function-controller && make "${INTEGRATION_SUITE}")
  job_status=$?
  set -o errexit

  collect_results

  echo "Exit code ${job_status}"

  exit $job_status
}
