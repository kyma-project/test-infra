#!/usr/bin/env bash

set -eo pipefail

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly LIB_DIR="$( cd "${SCRIPT_DIR}/lib" && pwd )"
readonly ARTIFACTS_DIR="${ARTIFACTS:-"$( pwd )/tmp"}"
readonly KIND_RESOURCES_DIR="${SCRIPT_DIR}/kind/resources"
mkdir -p "${ARTIFACTS_DIR}"

# shellcheck disable=SC1090
source "${LIB_DIR}/log.sh"
# shellcheck disable=SC1090
source "${LIB_DIR}/junit.sh"
# shellcheck disable=SC1090
source "${LIB_DIR}/docker.sh"
# shellcheck disable=SC1090
source "${LIB_DIR}/kind.sh"
# shellcheck disable=SC1090
source "${LIB_DIR}/kubernetes.sh"
# shellcheck disable=SC1090
source "${LIB_DIR}/kyma.sh"

function finalize {
    local -r exit_status=$?
    local finalization_failed="false"

    junit::test_start "Finalization"
    log::info "Finalizing job" 2>&1 | junit::test_output

    log::info "Printing all docker processes" 2>&1 | junit::test_output
    docker::print_processes 2>&1 | junit::test_output || finalization_failed="true"

    if [[ ${CLUSTER_PROVISIONED} = "true" ]]; then
        log::info "Exporting cluster logs to ${ARTIFACTS_DIR}" 2>&1 | junit::test_output
        kind::export_logs "${CLUSTER_NAME}" 2>&1 | junit::test_output || finalization_failed="true"

        if [[ ${DELETE_CLUSTER} = "true" ]]; then
            log::info "Deleting cluster" 2>&1 | junit::test_output
            kind::delete_cluster "${CLUSTER_NAME}" 2>&1 | junit::test_output || finalization_failed="true"
        fi
    fi

    if [[ ${finalization_failed} = "true" ]]; then
        junit::test_fail || true
    else
        junit::test_pass
    fi

    junit::suite_save

    return "${exit_status}"
}

trap junit::test_fail ERR
trap finalize EXIT

CLUSTER_PROVISIONED="false"

ENSURE_KUBECTL="false"
UPDATE_HOSTS="false"
DELETE_CLUSTER="false"
TUNE_INOTIFY="false"
START_DOCKER="false"

CLUSTER_NAME="kyma"
KUBERNETES_VERSION="v1.14.6"
CLUSTER_CONFIG="${KIND_RESOURCES_DIR}/../cluster.yaml"
KYMA_SOURCES="${GOPATH}/src/github.com/kyma-project/kyma"
KYMA_OVERRIDES="${KYMA_SOURCES}/installation/resources/installer-config-local.yaml.tpl"
KYMA_INSTALLER="${KYMA_SOURCES}/installation/resources/installer-local.yaml"
KYMA_INSTALLATION_CR="${KYMA_SOURCES}/installation/resources/installer-cr.yaml.tpl"
KYMA_INSTALLATION_TIMEOUT="30m"
readonly KYMA_DOMAIN="kyma.local"
readonly KYMA_INSTALLER_NAME="kyma-installer:kind"

function read_flags {
    while test $# -gt 0; do
        case "$1" in
            -h|--help)
                shift
                echo "Script that installs Kyma on kind"
                echo " "
                echo "Options:"
                echo "  -h --help                       Print usage."
                echo "     --ensure-kubectl             Update kubectl to the same version as cluster."
                echo "     --update-hosts               Append hosts file with Kyma hosts assigned to worker node."
                echo "     --delete-cluster             Deletes cluster at the end of script execution."
                echo "     --tune-inotify               Tune inotify instances and watches."
                echo "     --start-docker               Start the Docker Daemon."
                echo "     --cluster-name               Name of the kind cluster, default \`kyma\`."
                echo "     --cluster-config             Path to kind cluster configuration, default \`kind/cluster.yaml\`."
                echo "     --kubernetes-version         Kubernetes version, default \`v1.14.6\`."
                echo "     --kyma-sources               Path to the Kyma sources, default \`\${GOPATH}/src/github.com/kyma-project/kyma\`."
                echo "     --kyma-overrides             Path to the Kyma overrides, default \`\${KYMA_SOURCES}/installation/resources/installer-config-local.yaml.tpl\`."
                echo "     --kyma-installer             Path to the Kyma Installer yaml, default \`\${KYMA_SOURCES}/installation/resources/installer-local.yaml\`."
                echo "     --kyma-installation-cr       Path to the Kyma Installation CR, default \`\${KYMA_SOURCES}/installation/resources/installer-cr.yaml.tpl\`."
                echo "  -t --kyma-installation-timeout  The installation timeout, default \`30m\`."
                echo " "
                echo "Environment variables:"
                echo "  ARTIFACTS  If not set, all artifacts are stored in \`tmp\` directory"
                exit 0
                ;;
            --ensure-kubectl)
                shift
                ENSURE_KUBECTL="true"
                ;;
            --update-hosts)
                shift
                UPDATE_HOSTS="true"
                ;;
            --delete-cluster)
                shift
                DELETE_CLUSTER="true"
                ;;
            --tune-inotify)
                shift
                TUNE_INOTIFY="true"
                ;;
            --start-docker)
                shift
                START_DOCKER="true"
                ;;
            --cluster-name)
                shift
                CLUSTER_NAME="${1}"
                shift
                ;;
            --cluster-config)
                shift
                CLUSTER_CONFIG="${1}"
                shift
                ;;
            --kubernetes-version)
                shift
                KUBERNETES_VERSION="${1}"
                shift
                ;;
            --kyma-sources)
                shift
                KYMA_SOURCES="${1}"
                shift
                ;;
            --kyma-overrides)
                shift
                KYMA_OVERRIDES="${1}"
                shift
                ;;
            --kyma-installer)
                shift
                KYMA_INSTALLER="${1}"
                shift
                ;;
            --kyma-installation-cr)
                shift
                KYMA_INSTALLATION_CR="${1}"
                shift
                ;;
            -t|--kyma-installation-timeout)
                shift
                KYMA_INSTALLATION_TIMEOUT="${1}"
                shift
                ;;
            *)
                log::error "$1 is not a recognized flag, use --help flag for a list of avaiable options!"
                return 1
                ;;
        esac
    done

    readonly ENSURE_KUBECTL \
        UPDATE_HOSTS \
        DELETE_CLUSTER \
        TUNE_INOTIFY \
        START_DOCKER \
        CLUSTER_NAME \
        KUBERNETES_VERSION \
        CLUSTER_CONFIG \
        KYMA_SOURCES \
        KYMA_OVERRIDES \
        KYMA_INSTALLER \
        KYMA_INSTALLATION_CR \
        KYMA_INSTALLATION_TIMEOUT
}

function tune_inotify {
    log::info "Increasing limits for inotify"
    sysctl -w fs.inotify.max_user_watches=524288
    sysctl -w fs.inotify.max_user_instances=512
}

function main {
    junit::suite_init "Kyma_Integration"

    junit::test_start "Tune_Inotify"
    if [[ ${TUNE_INOTIFY} = "true" ]]; then
        tune_inotify 2>&1 | junit::test_output
        junit::test_pass
    else
        junit::test_skip "Disabled"
    fi

    junit::test_start "Start_Docker_Daemon"
    if [[ ${START_DOCKER} = "true" ]]; then
        log::info "Starting Docker daemon" 2>&1 | junit::test_output
        docker::start 2>&1 | junit::test_output
        junit::test_pass
    else
        junit::test_skip "Disabled"
    fi

    junit::test_start "Ensure_Kubectl"
    if [[ ${ENSURE_KUBECTL} = "true" ]]; then
        log::info "Ensuring kubectl version" 2>&1 | junit::test_output
        kubernetes::ensure_kubectl "${KUBERNETES_VERSION}" 2>&1 | junit::test_output
        junit::test_pass
    else
        junit::test_skip
    fi

    junit::test_start "Create_Cluster"
    log::info "Creating cluster" 2>&1 | junit::test_output
    kind::create_cluster "${CLUSTER_NAME}" "${KUBERNETES_VERSION}" "${CLUSTER_CONFIG}" 2>&1 | junit::test_output
    CLUSTER_PROVISIONED="true"
    readonly CLUSTER_PROVISIONED
    local -r worker_ip="$(kind::worker_ip "${CLUSTER_NAME}")"
    junit::test_pass

    junit::test_start "Install_Default_Resources"
    log::info "Installing default resources" 2>&1 | junit::test_output
    kind::install_default "${KIND_RESOURCES_DIR}" 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Build_Kyma_Installer"
    log::info "Building Kyma Installer Docker Image" 2>&1 | junit::test_output
    kyma::build_installer "${KYMA_SOURCES}" "${KYMA_INSTALLER_NAME}" 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Load_Kyma_Installer"
    log::info "Loading Kyma Installer to cluster" 2>&1 | junit::test_output
    kind::load_image "${CLUSTER_NAME}" "${KYMA_INSTALLER_NAME}" 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Install_Tiller"
    log::info "Installing Tiller" 2>&1 | junit::test_output
    kyma::install_tiller "${KYMA_SOURCES}" 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Load_Kyma_Configuration"
    log::info "Loading Kyma Configuration" 2>&1 | junit::test_output
    kyma::load_config "${worker_ip}" "${KYMA_DOMAIN}" "${KYMA_OVERRIDES}" 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Install_Kyma"
    log::info "Installing Kyma" 2>&1 | junit::test_output
    kyma::install "${KYMA_SOURCES}" "${KYMA_INSTALLER_NAME}" "${KYMA_INSTALLER}" "${KYMA_INSTALLATION_CR}" "${KYMA_INSTALLATION_TIMEOUT}" 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Update_Hosts_File"
    if [[ ${UPDATE_HOSTS} = "true" ]]; then
        log::info "Updating hosts file" 2>&1 | junit::test_output
        kyma::update_hosts "${worker_ip}" 2>&1 | junit::test_output
        junit::test_pass
    else
        junit::test_skip
    fi

    junit::test_start "Test_Kyma"
    log::info "Testing Kyma" 2>&1 | junit::test_output
    kyma::test "${KYMA_SOURCES}" 2>&1 | junit::test_output
    junit::test_pass
}

read_flags "$@"
main
