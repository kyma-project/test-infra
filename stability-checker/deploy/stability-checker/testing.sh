#!/usr/bin/env bash
CURRENT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
KYMA_TEST_TIMEOUT=${KYMA_TEST_TIMEOUT:=1h}

readonly CONCURRENCY=5
# Should be fixed name, it is displayed in TestGrid
readonly SUITE_NAME="testsuite-all"

# shellcheck disable=SC1090
source "${CURRENT_PATH}/testing-helpers.sh"

kc="kubectl $(context_arg)"

function main() {
  echo "----------------------------"
  echo "- Testing Kyma..."
  echo "----------------------------"

  cts::check_crd_exist

  cts::delete

  inject_addons_if_necessary

  log::info "- Running Kyma tests"
  # match all tests
  # shellcheck disable=SC2086
  kyma test run \
                --name "${SUITE_NAME}" \
                --concurrency "${CONCURRENCY}" \
                --max-retries 1 \
                --timeout "${KYMA_TEST_TIMEOUT}" \
                --watch \
                --non-interactive

  echo "- Tests results"
  kubectl get cts  ${SUITE_NAME} -o=go-template --template='{{range .status.results}}{{printf "Test status: %s - %s" .name .status }}{{ if gt (len .executions) 1 }}{{ print " (Retried)" }}{{end}}{{print "\n"}}{{end}}'

  log::info "All test pods should be terminated. Checking..."
  waitForTestPodsTermination "${SUITE_NAME}"
  cleanupExitCode=$?

  log::info "- ClusterTestSuite details"
  ${kc} get cts "${SUITE_NAME}" -oyaml

  statusSucceeded=$(${kc} get cts "${SUITE_NAME}"  -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
  if [[ "${statusSucceeded}" != *"True"* ]]; then
    log::info "- Fetching logs due to test suite failure"
    testExitCode=1

    echo "- Fetching logs from testing pods in Failed status..."
    kyma test logs "${SUITE_NAME}" --test-status Failed

    echo "- Fetching logs from testing pods in Unknown status..."
    kyma test logs "${SUITE_NAME}" --test-status Unknown

    echo "- Fetching logs from testing pods in Running status due to running afer test suite timeout..."
    kyma test logs "${SUITE_NAME}" --test-status Running
  fi

  exit $((testExitCode + cleanupExitCode))
}

main
