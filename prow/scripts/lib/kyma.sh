#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"

# kyma::install starts Kyma installation on the cluster
#
# Arguments:
#   $1 - Path to the Kyma sources
#   $2 - Name of the installer Docker Image
#   $3 - Path to the installer resource
#   $4 - Path to the installer custom resource
#   $5 - Installation timeout
function kyma::install {
    "${1}/installation/scripts/concat-yamls.sh" "${3}" "${4}" \
        | sed -e 's;image: eu.gcr.io/kyma-project/.*installer:.*$;'"image: ${2};" \
        | sed -e "s/__VERSION__/0.0.1/g" \
        | sed -e "s/__.*__//g" \
        | kubectl apply -f-
    
    kyma::is_installed "${1}" "${5}"
}

# kyma::alpha_deploy_kyma starts Kyma deployment using alpha feature
function kyma::alpha_deploy_kyma() {
  log::info "Deploying Kyma"

  if [[ "$EXECUTION_PROFILE" == "evaluation" ]]; then
    kyma alpha deploy --ci --profile evaluation --value global.isBEBEnabled=true --source=local --workspace "${KYMA_SOURCES_DIR}" --verbose
  elif [[ "$EXECUTION_PROFILE" == "production" ]]; then
    kyma alpha deploy --ci --profile production --value global.isBEBEnabled=true --source=local --workspace "${KYMA_SOURCES_DIR}" --verbose
  else
    kyma alpha deploy --ci --value global.isBEBEnabled=true --source=local --workspace "${KYMA_SOURCES_DIR}" --verbose
  fi
}

# kyma::is_installed waits for Kyma installation finish
#
# Arguments:
#   $1 - Path to the Kyma sources
#   $2 - Installation timeout
function kyma::is_installed {
    "${1}/installation/scripts/is-installed.sh" --timeout "${2}"
}

# kyma::load_config loads Kyma overrides to the cluster. Also sets domain and cluster IP
#
# Arguments:
#   $1 - IP of the cluster
#   $2 - Domain name
#   $3 - Path to the overrides file
function kyma::load_config {
    kubectl create namespace "kyma-installer" || echo "Ignore namespace creation"

    < "${3}" sed 's/\.minikubeIP: .*/\.minikubeIP: '"${1}"'/g' \
        | sed 's/\.domainName: .*/\.domainName: '"${2}"'/g' \
        | kubectl apply -f-
}

# kyma::test starts the Kyma integration tests
#
# Arguments:
#   $1 - Path to the scripts (../) directory
function kyma::test {
    "${1}/kyma-testing.sh"
}

# kyma::build_installer builds Kyma Installer Docker image
#
# Arguments:
#   $1 - Path to the Kyma sources
#   $2 - Installer Docker image name
function kyma::build_installer {
    docker build -t "${2}" -f "${1}/tools/kyma-installer/kyma.Dockerfile" "${1}"
}

# kyma::build_and_push_installer builds Kyma Installer Docker image
#
# Arguments:
#   $1 - Path to the Kyma sources
#   $2 - Installer Docker image name
function kyma::build_and_push_installer {
    docker build -t "${2}" -f "${1}/tools/kyma-installer/kyma.Dockerfile" "${1}"
    docker push "${2}"
}

# kyma::update_hosts appends hosts file with Kyma DNS records
function kyma::update_hosts {
    # TODO(michal-hudy):  Switch to local DNS server if possible
    local -r hosts="$(kubectl get virtualservices --all-namespaces -o jsonpath='{.items[*].spec.hosts[*]}')"
    echo "127.0.0.1 ${hosts}" | tee -a /etc/hosts > /dev/null
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
        | jq -r "del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(\".\") | map(tonumber)) | reverse | [ .[] | select(.tag_name|test(\"^1\\\.22\\\.\"))] | .[0].tag_name")

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
