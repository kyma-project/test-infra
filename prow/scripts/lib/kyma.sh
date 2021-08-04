#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"

# kyma::deploy_kyma starts Kyma deployment using new installation method
# Arguments:
# optional:
# s - Kyma sources directory
# p - execution profile
# u - upgrade (this will not reuse helm values which is already set)
function kyma::deploy_kyma() {

    local OPTIND
    local executionProfile=
    local kymaSourcesDir=""
    local upgrade=

    while getopts ":p:s:u:" opt; do
        case $opt in
            p)
                if [ -n "$OPTARG" ]; then
                    executionProfile="$OPTARG"
                fi ;;
            s)
                kymaSourcesDir="$OPTARG" ;;
            u)
                upgrade="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    log::info "Deploying Kyma"

    if [[ -n "$executionProfile" ]]; then
        if [[ -n "$upgrade" ]]; then
            kyma deploy --reuse-values=false --ci --profile "$executionProfile" --source=local --workspace "${kymaSourcesDir}" --verbose
        else
            kyma deploy --ci --profile "$executionProfile" --source=local --workspace "${kymaSourcesDir}" --verbose
        fi
    else
        if [[ -n "$upgrade" ]]; then
            kyma deploy --reuse-values=false --ci --source=local --workspace "${kymaSourcesDir}" --verbose
        else
            kyma deploy --ci --source=local --workspace "${kymaSourcesDir}" --verbose
        fi
    fi
}

# kyma::delete_kyma uninstalls Kyma using new deletion method
function kyma::delete_kyma() {
  log::info "Uninstalling Kyma"

  kyma delete --ci --verbose
}

# kyma::get_last_release_version returns latest Kyma release version
#
# Arguments:
#   t - GitHub token
# Returns:
#   Last Kyma release version
function kyma::get_last_release_version {

    local OPTIND
    local githubToken

    while getopts ":t:" opt; do
        case $opt in
            t)
                githubToken="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$githubToken" "Github token was not provided. Exiting..."
    
    # shellcheck disable=SC2034
    kyma_get_last_release_version_return_version=$(curl --silent --fail --show-error -H "Authorization: token $githubToken" "https://api.github.com/repos/kyma-project/kyma/releases" \
        | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-1].tag_name')
}

kyma::install_cli() {
    if ! [[ -x "$(command -v kyma)" ]]; then
        local settings
        local kyma_version
        settings="$(set +o); set -$-"
        mkdir -p "/tmp/bin"
        export PATH="/tmp/bin:${PATH}"
        os=$(host::os)

        pushd "/tmp/bin" || exit

        echo "--> Install kyma CLI ${os} locally to /tmp/bin"

        curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}?alt=media"
        chmod +x kyma
        kyma_version=$(kyma version --client)
        echo "--> Kyma CLI version: ${kyma_version}"
        echo "OK"
        popd || exit
        eval "${settings}"
    else
        log::info "Kyma CLI is already installed: $(kyma version -c)"
    fi
}

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
      echo "Unsupported host OS. Must be Linux or Mac OS X."
      exit 1
      ;;
  esac
  echo "${host_os}"
}

kyma::run_test_log_collector(){
    if [ "${ENABLE_TEST_LOG_COLLECTOR:-}" = true ] && [[ -n "${LOG_COLLECTOR_SLACK_TOKEN:-}" ]]; then
        if [[ "$BUILD_TYPE" == "master" ]] || [[ -z "${BUILD_TYPE:-}" ]]; then
            log::info "Install test-log-collector"
            export PROW_JOB_NAME=$1
            (
                TLC_DIR="${TEST_INFRA_SOURCES_DIR}/development/test-log-collector"

                helm install test-log-collector --set slackToken="${LOG_COLLECTOR_SLACK_TOKEN}" \
                --set prowJobName="${PROW_JOB_NAME}" \
                "${TLC_DIR}/chart/test-log-collector" \
                --namespace=kyma-system \
                --wait \
                --timeout=600s || true # we want it to work on "best effort" basis, which does not interfere with cluster
            )
        fi
    fi
}

# Arguments:
# required:
# s - suite name
#
# Returns
# kyma_test_summary_return_exit_code - exit code
#
kyma::test_summary() {

    local OPTIND
    local suiteName
    local tests_exit=0
    local statusSucceeded

    while getopts ":s:" opt; do
        case $opt in
            s)
                suiteName="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$suiteName" "Suite name was not provided. Exiting..."

    echo "Test Summary"
    kyma test status "${suiteName}" -owide

    statusSucceeded=$(kubectl get cts "${suiteName}" -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
    if [[ "${statusSucceeded}" != *"True"* ]]; then
        tests_exit=1
        echo "- Fetching logs due to test suite failure"

        echo "- Fetching logs from testing pods in Failed status..."
        kyma test logs "${suiteName}" --test-status Failed

        echo "- Fetching logs from testing pods in Unknown status..."
        kyma test logs "${suiteName}" --test-status Unknown

        echo "- Fetching logs from testing pods in Running status due to running afer test suite timeout..."
        kyma test logs "${suiteName}" --test-status Running
    fi

    echo "ClusterTestSuite details"
    kubectl get cts "${suiteName}" -oyaml
    if [ $tests_exit -ne 0 ]; then
        log::error "Tests have failed"
    else
        log::success "Tests completed"
    fi

    # shellcheck disable=SC2034
    kyma_test_summary_return_exit_code="$tests_exit"
}
