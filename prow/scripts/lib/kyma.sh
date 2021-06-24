#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"

# kyma::alpha_deploy_kyma starts Kyma deployment using alpha feature
# Arguments:
# optional:
# s - Kyma sources directory
# p - execution profile
function kyma::alpha_deploy_kyma() {

    local OPTIND
    local executionProfile=
    local kymaSourcesDir=""

    while getopts ":p:s:" opt; do
        case $opt in
            p)
                if [ -n "$OPTARG" ]; then
                    executionProfile="$OPTARG"
                fi ;;
            s)
                kymaSourcesDir="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    log::info "Deploying Kyma"

    if [[ -n "$executionProfile" ]]; then
        kyma alpha deploy --ci --profile "$executionProfile" --value global.isBEBEnabled=true --source=local --workspace "${kymaSourcesDir}" --verbose
    else
        kyma alpha deploy --ci --value global.isBEBEnabled=true --source=local --workspace "${kymaSourcesDir}" --verbose
    fi
}

# kyma::alpha_delete_kyma uninstalls Kyma using alpha feature
function kyma::alpha_delete_kyma() {
  log::info "Uninstalling Kyma"

  kyma alpha delete --ci --verbose
}

# kyma::get_last_release_version returns latest Kyma release version
#
# Arguments:
#   $1 - GitHub token
# Returns:
#   Last Kyma release version
function kyma::get_last_release_version {
    if [[ -z "$1" ]]; then
        log::error "Github token is missing, please provide token"
        exit 1
    fi
    
    version=$(curl --silent --fail --show-error -H "Authorization: token ${1}" "https://api.github.com/repos/kyma-project/kyma/releases" \
        | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-1].tag_name')

    echo "${version}"
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

kyma::test_summary() {
    local tests_exit=0
    if [[ -n "${SUITE_NAME:-}" ]]; then
        echo "Test Summary"
        kyma test status "${SUITE_NAME}" -owide

        statusSucceeded=$(kubectl get cts "${SUITE_NAME}" -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
        if [[ "${statusSucceeded}" != *"True"* ]]; then
            tests_exit=1
            echo "- Fetching logs due to test suite failure"

            echo "- Fetching logs from testing pods in Failed status..."
            kyma test logs "${SUITE_NAME}" --test-status Failed

            echo "- Fetching logs from testing pods in Unknown status..."
            kyma test logs "${SUITE_NAME}" --test-status Unknown

            echo "- Fetching logs from testing pods in Running status due to running afer test suite timeout..."
            kyma test logs "${SUITE_NAME}" --test-status Running
        fi

        echo "ClusterTestSuite details"
        kubectl get cts "${SUITE_NAME}" -oyaml
        return $tests_exit
    else
        tests_exit=1
        echo "SUITE_NAME was not set"
        return $tests_exit
    fi
}
