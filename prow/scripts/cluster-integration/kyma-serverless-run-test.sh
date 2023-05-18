#!/usr/bin/env bash

set -o pipefail # Fail a pipe if any sub-command fails.

run_tests() {
  export SERVERLESS_OVERRIDES_DIR="./overrides"
  export INTEGRATION_SUITE=${1:-serverless-integration}
  export KYMA_SOURCES_DIR=${2:-./kyma}

  date

  # check for test secrets.
  if [[ -e "${SERVERLESS_OVERRIDES_DIR}/git-auth.env" ]]; then
    # shellcheck source=/dev/null
    source "${SERVERLESS_OVERRIDES_DIR}/git-auth.env"
  fi

  #https://github.com/kyma-project/test-infra/issues/6513
  export PATH=${PATH}:/usr/local/go/bin

  export APP_TEST_CLEANUP="onSuccessOnly"
  (cd "${KYMA_SOURCES_DIR}/tests/function-controller" && make "${INTEGRATION_SUITE}")
  return $?
}
