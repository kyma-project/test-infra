#!/usr/bin/env bash
CURRENT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

readonly TMP_DIR=$(mktemp -d)
readonly JUNIT_REPORT_PATH="${ARTIFACTS:-${TMP_DIR}}/junit_Kyma_octopus-test-suite.xml"
readonly CONCURRENCY=5
readonly SUITE_NAME="testsuite-all-$(date '+%Y-%m-%d-%H-%M')"

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

    curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-develop-version/kyma-${os}?alt=media"
    chmod +x kyma

    log::success "OK"

    popd
}

cts::check_crd_exist() {
  ${kc} get clustertestsuites.testing.kyma-project.io > /dev/null 2>&1
  if [[ $? -eq 1 ]]
  then
     echo "ERROR: script requires ClusterTestSuite CRD"
     exit 1
  fi
}

cts::delete() {
  existingCTSs=$(${kc} get cts -o custom-columns=NAME:.metadata.name --no-headers=true)
  for cts in ${existingCTSs}
  do
    kyma test delete "${cts}"
  done

}

function main() {
  echo "----------------------------"
  echo "- Testing Kyma..."
  echo "----------------------------"

  export INSTALL_DIR=${TMP_DIR}
  install::kyma_cli

  cts::check_crd_exist

  cts::delete

  matchTests="" # match all tests

  ${kc} get cm dex-config -n kyma-system -ojsonpath="{.data}" | grep --silent "#__STATIC_PASSWORDS__"
  if [[ $? -eq 1 ]]
  then
    # if static users are not available, do not execute tests which requires them
    matchTests=$(${kc} get testdefinitions --all-namespaces -l 'require-static-users!=true' -o=go-template='{{- range .items}} {{.metadata.name}}{{- end}}')
    echo "WARNING: following tests will be skipped due to the lack of static users:"
    ${kc} get testdefinitions --all-namespaces -l 'require-static-users=true' -o=go-template --template='{{- range .items}}{{printf " - %s\n" .metadata.name}}{{- end}}'
  fi

  log::info "- Creating ClusterAddonsConfiguration which provides the testing addons"
  injectTestingAddons
  if [[ $? -eq 1 ]]; then
    exit 1
  fi

  trap removeTestingAddons EXIT

  log::info "- Running Kyma tests"
  # shellcheck disable=SC2086
  kyma test run ${matchTests} \
                --name "${SUITE_NAME}" \
                --concurrency "${CONCURRENCY}" \
                --max-retries 1 \
                --timeout "1h" \
                --watch \
                --non-interactive

  log::info "- Test summary"
  kyma test status "${SUITE_NAME}" -owide

  # TODO(mszostok): decide if this should be supported by `kyma test status`,
  #  right now we do not have the exit code
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

  log::info "- Generate JUnit test summary"
  kyma test status "${SUITE_NAME}" -ojunit > "${JUNIT_REPORT_PATH}"

  log::info "All test pods should be terminated. Checking..."
  waitForTestPodsTermination "${SUITE_NAME}"
  cleanupExitCode=$?

  log::info "- ClusterTestSuite details"
  kubectl get cts "${SUITE_NAME}" -oyaml

  cts::delete

  log::info "Images with tag latest are not allowed. Checking..."
  printImagesWithLatestTag
  latestTagExitCode=$?

  exit $((testExitCode + cleanupExitCode + latestTagExitCode))
}

main
