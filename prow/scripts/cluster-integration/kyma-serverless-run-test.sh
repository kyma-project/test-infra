#!/usr/bin/env bash

set -o pipefail # Fail a pipe if any sub-command fails.

run_tests() {
  export SERVERLESS_OVERRIDES_DIR="./overrides"
  export INTEGRATION_SUITE=("$@")

  date

  # check for test secrets.
  if [[ -e "${SERVERLESS_OVERRIDES_DIR}/git-auth.env" ]]; then
    # shellcheck source=/dev/null
    source "${SERVERLESS_OVERRIDES_DIR}/git-auth.env"
  fi

  #https://github.com/kyma-project/test-infra/issues/6513
  export PATH=${PATH}:/usr/local/go/bin

  export APP_TEST_CLEANUP="onSuccessOnly"
  (cd "${SERVERLESS_SOURCES}/tests/serverless" && make "${INTEGRATION_SUITE[@]}")
  return $?
}
