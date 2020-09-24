#!/usr/bin/env bash
CURRENT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
KYMA_TEST_TIMEOUT=${KYMA_TEST_TIMEOUT:=1h}

readonly TMP_DIR=$(mktemp -d)
readonly JUNIT_REPORT_PATH="${ARTIFACTS:-${TMP_DIR}}/junit_Kyma_octopus-test-suite.xml"
readonly CONCURRENCY=5
# Should be fixed name, it is displayed in TestGrid
readonly SUITE_NAME="testsuite-all"
readonly CONSOLE_SUITE_NAME="testsuite-console"

# shellcheck disable=SC1090
source "${CURRENT_PATH}/lib/testing-helpers.sh"

kc="kubectl $(context_arg)"

cleanup() {
    rm -rf "${TMP_DIR}"
}

trap cleanup EXIT

host::os() {
  local host_os
  case "$(uname -s)" in
    Darwin)
      host_os=darwin
      ;;
    Linux)
      host_os=linux
      ;;
    *)
      log::error "Unsupported host OS. Must be Linux or Mac OS X."
      exit 1
      ;;
  esac
  echo "${host_os}"
}

install::kyma_cli() {
    mkdir -p "${INSTALL_DIR}/bin"
    export PATH="${INSTALL_DIR}/bin:${PATH}"
    os=$(host::os)

    pushd "${INSTALL_DIR}/bin"

    log::info "- Install kyma CLI ${os} locally to a tempdir..."

    curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}?alt=media"
    chmod +x kyma

    log::success "OK"

    popd
}

function printImagesWithLatestTag() {
    retry=10
    while true; do
        # shellcheck disable=SC2046
        images=$(kubectl $(context_arg)  get pods --all-namespaces -o jsonpath="{..image}" |\
        tr -s "[:space:]" "\n" |\
        grep ":latest")

        # TODO(michal-hudy): it shoudn"t be done that way, grep returns 1 when no lines match, same bug in kyma repository....
        if [[ $? -lt 2 ]]; then
            break
        fi
        (( retry-- ))
        if [[ ${retry} -eq 0 ]]; then
            log::error "Reached maximum attempts, not trying any longer"
            return 1
        fi
        sleep 5
    done

    if [ ${#images} -ne 0 ]; then
        log::error "${images}"
        log::error "FAILED"
        return 1
    fi
    log::success "OK"
    return 0
}

function main() {
  echo "----------------------------"
  echo "- Testing Kyma..."
  echo "----------------------------"

  export INSTALL_DIR=${TMP_DIR}
  install::kyma_cli

  cts::check_crd_exist

  cts::delete

  inject_addons_if_necessary


  echo "Stopping CBS for testing purposes.."
  ${kc} scale --replicas=0 deployment/console-backend -n kyma-system

  echo "KYMA_TESTS: ${KYMA_TESTS}"
  TESTS=$(kyma test definitions --ci | tr "\r\n" " ")

  KYMA_TESTS_WITHOUT_CBS=${TESTS//"console-backend"/}
  KYMA_TESTS_WITHOUT_CONSOLE=${KYMA_TESTS_WITHOUT_CBS//"console-web"/}
  
  echo "TESTS: ${TESTS}"
  echo "KYMA_TESTS_WITHOUT_CONSOLE: ${KYMA_TESTS_WITHOUT_CONSOLE}"

  log::info "- Running Kyma tests"
  # match all tests besides console chart
  # shellcheck disable=SC2086
  kyma test run ${KYMA_TESTS_WITHOUT_CONSOLE} \
                --name "${SUITE_NAME}" \
                --concurrency "${CONCURRENCY}" \
                --max-retries 1 \
                --timeout "${KYMA_TEST_TIMEOUT}" \
                --watch \
                --non-interactive

  log::info "- Test summary"
  kyma test status "${SUITE_NAME}" -owide

  # TODO(mszostok): decide if this should be supported by `kyma test status`,
  #  right now we do not have the exit code
  statusSucceeded=$(${kc} get cts "${SUITE_NAME}"  -ojsonpath="{.status.conditions[?(@.type=="Succeeded")]}")
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


  ${kc} scale --replicas=1 deployment/console-backend -n kyma-system
  echo "Starting CBS for testing console domain"

  log::info "- Running Kyma Console tests"
  kyma test run console-web console-backend \
                --name "${CONSOLE_SUITE_NAME}" \
                --concurrency "${CONCURRENCY}" \
                --max-retries 1 \
                --timeout "${KYMA_TEST_TIMEOUT}" \
                --watch \
                --non-interactive

  log::info "- Test summary"
  kyma test status "${CONSOLE_SUITE_NAME}" -owide

  statusSucceeded=$(${kc} get cts "${CONSOLE_SUITE_NAME}"  -ojsonpath="{.status.conditions[?(@.type=="Succeeded")]}")
  if [[ "${statusSucceeded}" != *"True"* ]]; then
    log::info "- Fetching logs due to test suite failure"
    consoleTestExitCode=1

    echo "- Fetching logs from testing pods in Failed status..."
    kyma test logs "${CONSOLE_SUITE_NAME}" --test-status Failed

    echo "- Fetching logs from testing pods in Unknown status..."
    kyma test logs "${CONSOLE_SUITE_NAME}" --test-status Unknown

    echo "- Fetching logs from testing pods in Running status due to running afer test suite timeout..."
    kyma test logs "${CONSOLE_SUITE_NAME}" --test-status Running
  fi


  log::info "- Generate JUnit test summary"
  kyma test status "${SUITE_NAME}" -ojunit | sed "s/ (executions: [0-9]*)"/"/g" > "${JUNIT_REPORT_PATH}"
  kyma test status "${SUITE_NAME}" -ojunit | sed "s/ (executions: [0-9]*)"/"/g" > "${JUNIT_REPORT_PATH}"

  log::info "All test pods should be terminated. Checking..."
  waitForTestPodsTermination "${SUITE_NAME}"
  cleanupExitCode=$?
  waitForTestPodsTermination "${CONSOLE_SUITE_NAME}"
  consoleCleanupExitCode=$?

  log::info "- ClusterTestSuites details"
  ${kc} get cts "${SUITE_NAME}" -oyaml
  ${kc} get cts "${CONSOLE_SUITE_NAME}" -oyaml

  # TODO (mhudy): cts shouldn"t be deleted because all test pods are deleted too and kind export will not store them
  # cts::delete

  log::info "Images with tag latest are not allowed. Checking..."
  printImagesWithLatestTag
  latestTagExitCode=$?

  exit $((testExitCode + consoleTestExitCode + cleanupExitCode + consoleCleanupExitCode + latestTagExitCode))
}

main
