#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"

# kyma::deploy_kyma starts Kyma deployment using new installation method
# Arguments:
# optional:
# s - Kyma source
# d - Kyma workspace directory
# p - execution profile
# u - upgrade (this will not reuse helm values which is already set)
function kyma::deploy_kyma() {

    local OPTIND
    local executionProfile=
    local kymaSource=""
    local kymaSourcesDir=""
    local upgrade=

  log::info "Checking Kyma optional arguments"
    while getopts ":s:p:d:u:" opt; do
        case $opt in
            s)
                kymaSource="$OPTARG"
    						log::info "Kyma Source to install: ${kymaSource}"
                ;;
            p)
                if [ -n "$OPTARG" ]; then
                    executionProfile="$OPTARG"
                    log::info "Execution Profile: ${executionProfile}"
                fi ;;
            d)
                kymaSourcesDir="$OPTARG"
                log::info "Kyma Source Directory: ${kymaSourcesDir}"
                 ;;
            u)
                upgrade="$OPTARG"
                log::info "Kyma upgrade option: ${upgrade}"
                ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    log::info "Deploying Kyma"

    if [[ -n "$kymaSource" ]]; then
        kyma deploy --ci --concurrency=8 --profile "$executionProfile" --source="${kymaSource}" --workspace "${kymaSourcesDir}" --verbose
    else
        if [[ -n "$executionProfile" ]]; then
            kyma deploy --ci --concurrency=8 --profile "$executionProfile" --source=local --workspace "${kymaSourcesDir}" --verbose
        else
            kyma deploy --ci --concurrency=8 --source=local --workspace "${kymaSourcesDir}" --verbose
        fi
    fi


}

# kyma::undeploy_kyma uninstalls Kyma
function kyma::undeploy_kyma() {
  log::info "Uninstalling Kyma"

  kyma undeploy --ci --verbose
}

# kyma::get_last_release_version returns latest Kyma release version
#
# Arguments:
#   t - GitHub token
#   v - searched version as a regular expression, e.g. "^1\." (optional)
# Returns:
#   Last Kyma release version
function kyma::get_last_release_version {

    local OPTIND
    local githubToken
    local searchedVersion=""

    log::info "Checking Github token and searched version arguments"
    while getopts ":t:v:" opt; do
        case $opt in
            t)
                utils::mask_debug_output; githubToken="$OPTARG"; utils::unmask_debug_output ;;
            v)
                searchedVersion="$OPTARG" ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::mask_debug_output
    utils::check_empty_arg "$githubToken" "Github token was not provided. Exiting..."
    utils::unmask_debug_output

    if [[ -n "${searchedVersion}" ]]; then
        utils::mask_debug_output
        # shellcheck disable=SC2034
        kyma_get_last_release_version_return_version=$(curl --silent --fail --show-error -H "Authorization: token $githubToken" "https://api.github.com/repos/kyma-project/kyma/releases" \
            | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | [.[]| select( .tag_name | match("'"${searchedVersion}"'"))] | .[-1].tag_name')
        utils::unmask_debug_output
    else
        utils::mask_debug_output
        # shellcheck disable=SC2034
        kyma_get_last_release_version_return_version=$(curl --silent --fail --show-error -H "Authorization: token $githubToken" "https://api.github.com/repos/kyma-project/kyma/releases" \
            | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-1].tag_name')
        utils::unmask_debug_output
    fi
}

function kyma::get_offset_minor_releases() {
    while getopts ":v:" opt; do
        case $opt in
        v)
            base="$OPTARG" ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
        :)
            echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    RE='[^0-9]*\([0-9]*\)[.]\([0-9]*\)[.]\([0-9]*\)\([0-9A-Za-z-]*\)'

    # shellcheck disable=SC2001
    MAJOR=$(echo "$base" | sed -e "s#$RE#\\1#")
    # shellcheck disable=SC2001
    MINOR=$(echo "$base" | sed -e "s#$RE#\\2#")
    # shellcheck disable=SC2001
    PATCH=$(echo "$base" | sed -e "s#$RE#\\3#")

    local index=0
    minor_release_versions[index]=$base

    # PREVIOUS_MINOR_VERSION_COUNT - Count of last Kyma2 minor versions to be upgraded from
    while [ $index -lt "$PREVIOUS_MINOR_VERSION_COUNT" ]; do
        # do not decrease the PATCH initially, first decrease MINOR then PATCH
        if [ "$PATCH" -gt 0 ] && [ "$index" -gt 0 ]; then
          PATCH=$((PATCH-1))
          newVersion="$MAJOR.$MINOR.$PATCH"

        elif [ "$MINOR" -gt 0 ]; then
          MINOR=$((MINOR-1))
          newVersion="^$MAJOR.$MINOR."
          PATCH=-1
        else
            break
        fi

        kyma::get_last_release_version \
        -t "${BOT_GITHUB_TOKEN}" \
        -v "${newVersion}"

        if [[ -z "$kyma_get_last_release_version_return_version" ]] || [[ "$kyma_get_last_release_version_return_version" = "null" ]] ; then
            log::info "### The last release version returned from the offset is ${kyma_get_last_release_version_return_version} and thus invalid"
            continue
        fi

        if [ "$PATCH" -lt 0 ]; then
          # shellcheck disable=SC2001
          PATCH=$(echo "$kyma_get_last_release_version_return_version" | sed -e "s#$RE#\\3#")
        fi

        newVersion="$MAJOR.$MINOR.$PATCH"

        index=$((index+1))

        # shellcheck disable=SC2034
        minor_release_versions[index]=$newVersion
    done

    log::info "#### Valid minor versions to be tested:" "${minor_release_versions[@]}"
}

# kyma::get_previous_release_version returns previous Kyma release version (i.e. one version before the latest released version)
#
# Arguments:
#   t - GitHub token
# Returns:
#   Previous Kyma release version
function kyma::get_previous_release_version {
    local OPTIND
    local githubToken

    while getopts ":t:" opt; do
        case $opt in
            t)
                githubToken="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&1; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&1 ;;
        esac
    done

    utils::check_empty_arg "$githubToken" "Github token was not provided. Exiting..."

    # shellcheck disable=SC2034
    kyma_get_previous_release_version_return_version=$(curl --silent --fail --show-error -H "Authorization: token $githubToken" "https://api.github.com/repos/kyma-project/kyma/releases" \
        | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-2].tag_name')
}

kyma::install_unstable_cli() {
    local settings
    local kyma_version
    settings="$(set +o); set -$-"
    mkdir -p "/tmp/bin"
    export PATH="/tmp/bin:${PATH}"
    os=$(host::os)

    pushd "/tmp/bin" || exit

    log::info "--> Install kyma CLI (unstable version) ${os} locally to /tmp/bin"

    curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-unstable/kyma-${os}?alt=media"
    chmod +x kyma
    kyma_version=$(kyma version --client)
    log::info "--> Kyma CLI (unstable) version: ${kyma_version}"
    log::info "OK"
    popd || exit
    eval "${settings}"
}


kyma::install_old_cli() {
    local settings
    local kyma_version
    settings="$(set +o); set -$-"
    mkdir -p "/tmp/bin"
    export PATH="/tmp/bin:${PATH}"
    os=$(host::os)

    pushd "/tmp/bin" || exit

    log::info "--> Install kyma CLI ${os} locally to /tmp/bin"

    if [[ "${KYMA_MAJOR_VERSION-}" == "1" ]]; then
        curl -sSLo kyma.tar.gz "https://github.com/kyma-project/cli/releases/download/1.24.8/kyma_${os}_x86_64.tar.gz"
        tar xvzf kyma.tar.gz
    else
        curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-unstable/kyma-${os}?alt=media"
    fi

    chmod +x kyma
    kyma_version=$(kyma version --client)
    log::info "--> Kyma CLI version: ${kyma_version}"
    log::info "OK"
    popd || exit
    eval "${settings}"
}

kyma::install_cli() { #latest CLI release
    local settings
    settings="$(set +o); set -$-"

    mkdir -p "/tmp/bin"
    export PATH="/tmp/bin:${PATH}"
    pushd "/tmp/bin" || exit

    local os
    os="$(uname -s)"
    if [[ -z "$os" || ! "$os" =~ ^(Darwin|Linux)$ ]]; then
        echo >&2 -e "Unsupported host OS. Must be Linux or Mac OS X."
        exit 1
    else
        readonly os
    fi

    curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/latest/download/kyma_${os}_x86_64.tar.gz" \
    && tar -zxvf kyma.tar.gz && chmod +x kyma \
    && rm -f kyma.tar.gz

    kyma_version=$(kyma version --client)
    log::info "--> Kyma CLI version: ${kyma_version}"
    log::info "OK"
    popd || exit
    eval "${settings}"
}

kyma::install_cli_from_reconciler_pr() {
  local install_dir
  declare -r install_dir="/usr/local/bin"
  mkdir -p "$install_dir"

  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  if [[ -z "$os" || ! "$os" =~ ^(linux)$ ]]; then
    log::error "Unsupported host OS. Must be Linux."
    exit 1
  else
    readonly os
  fi

  kyma_cli_url="https://storage.googleapis.com/kyma-cli-pr/kyma-${os}-pr-${PULL_NUMBER}"

  pushd "$install_dir" || exit
  log::info "Downloading Kyma CLI from: ${kyma_cli_url}"
  curl -Lo kyma "${kyma_cli_url}"
  chmod +x kyma
  popd || exit

  kyma version --client
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
