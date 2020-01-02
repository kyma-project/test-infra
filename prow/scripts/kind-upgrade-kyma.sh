#!/usr/bin/env bash

set -eo pipefail

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TMP_DIR="$(mktemp -d)"
readonly TMP_BIN_DIR="${TMP_DIR}/bin"
readonly LIB_DIR="$( cd "${SCRIPT_DIR}/lib" && pwd )"
readonly ARTIFACTS_DIR="${ARTIFACTS:-"${TMP_DIR}/artifacts"}"
readonly KIND_RESOURCES_DIR="${SCRIPT_DIR}/kind/resources"
mkdir -p "${ARTIFACTS_DIR}"
mkdir -p "${TMP_BIN_DIR}"
export PATH="${TMP_BIN_DIR}:${PATH}"

# shellcheck disable=SC1090
source "${LIB_DIR}/testing-helpers.sh"
# shellcheck disable=SC1090
source "${LIB_DIR}/host.sh"
# shellcheck disable=SC1090
source "${LIB_DIR}/junit.sh"
# shellcheck disable=SC1090
source "${LIB_DIR}/docker.sh"
# shellcheck disable=SC1090
source "${LIB_DIR}/kind.sh"
# shellcheck disable=SC1090
source "${LIB_DIR}/kubernetes.sh"
# shellcheck disable=SC1090
source "${LIB_DIR}/helm.sh"
# shellcheck disable=SC1090
source "${LIB_DIR}/kyma.sh"

# finalize stores logs, saves JUnit report and removes cluster
function finalize {
    local -r exit_status=$?
    local finalization_failed="false"

    junit::test_start "Finalization"
    log::info "Finalizing job" 2>&1 | junit::test_output

    log::info "Printing all docker processes" 2>&1 | junit::test_output
    docker::print_processes 2>&1 | junit::test_output || finalization_failed="true"

    if [[ ${CLUSTER_PROVISIONED} = "true" ]]; then
        log::info "Exporting cluster logs to ${ARTIFACTS_DIR}" 2>&1 | junit::test_output
        kind::export_logs "${CLUSTER_NAME}" "${ARTIFACTS_DIR}" 2>&1 | junit::test_output || finalization_failed="true"

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

    log::info "Deleting temporary dir ${TMP_DIR}"
    rm -rf "${TMP_DIR}" || true

    if [[ ${exit_status} -eq 0 ]]; then
        log::success "Job finished with success"
    else
        log::error "Job finished with error"
    fi

    return "${exit_status}"
}

trap finalize EXIT

CLUSTER_PROVISIONED="false"

ENSURE_KUBECTL="false"
UPDATE_HOSTS="false"
DELETE_CLUSTER="false"
TUNE_INOTIFY="false"
START_DOCKER="false"

ENSURE_HELM="false"
readonly HELM_VERSION="v2.10.0"

CLUSTER_NAME="kyma"
KUBERNETES_VERSION="v1.14.6"
CLUSTER_CONFIG="${KIND_RESOURCES_DIR}/../cluster.yaml"
KYMA_SOURCES="${GOPATH}/src/github.com/kyma-project/kyma"
readonly KYMA_OVERRIDES="${KYMA_SOURCES}/installation/resources/installer-config-kind.yaml.tpl"
readonly KYMA_INSTALLER="${KYMA_SOURCES}/installation/resources/installer.yaml"
readonly KYMA_INSTALLATION_CR="${KYMA_SOURCES}/installation/resources/installer-cr-cluster.yaml.tpl"
KYMA_INSTALLATION_TIMEOUT="30m"
readonly KYMA_DOMAIN="kyma.local"
readonly KYMA_INSTALLER_NAME="kyma-installer:kind"

export UPGRADE_TEST_PATH="${KYMA_SOURCES}/tests/end-to-end/upgrade/chart/upgrade"
# timeout in sec for helm operation install/test
export UPGRADE_TEST_HELM_TIMEOUT_SEC=10000
# timeout in sec for e2e upgrade test pods until they reach the terminating state
export UPGRADE_TEST_TIMEOUT_SEC=600
export UPGRADE_TEST_NAMESPACE="e2e-upgrade-test"
export UPGRADE_TEST_RELEASE_NAME="${UPGRADE_TEST_NAMESPACE}"
export UPGRADE_TEST_RESOURCE_LABEL="kyma-project.io/upgrade-e2e-test"
export UPGRADE_TEST_LABEL_VALUE_PREPARE="prepareData"
export UPGRADE_TEST_LABEL_VALUE_EXECUTE="executeTests"
export TEST_CONTAINER_NAME="tests"

# read_flags analyzes provided arguments as flags
#
# Arguments:
#   All arguments ar treated as flags
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
                echo "     --ensure-helm                Update helm client to the version expected by Kyma."
                echo "     --update-hosts               Append hosts file with Kyma hosts assigned to worker node."
                echo "     --delete-cluster             Deletes cluster at the end of script execution."
                echo "     --tune-inotify               Tune inotify instances and watches."
                echo "     --start-docker               Start the Docker Daemon."
                echo "     --cluster-name               Name of the kind cluster, default \`kyma\`."
                echo "     --cluster-config             Path to kind cluster configuration, default \`kind/cluster.yaml\`."
                echo "     --kubernetes-version         Kubernetes version, default \`v1.14.6\`."
                echo "     --kyma-sources               Path to the Kyma sources, default \`\${GOPATH}/src/github.com/kyma-project/kyma\`."
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
            --ensure-helm)
                shift
                ENSURE_HELM="true"
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
        ENSURE_HELM \
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

# tune_inotify configures INotify settings
function tune_inotify {
    log::info "Increasing limits for inotify"
    sysctl -w fs.inotify.max_user_watches=524288
    sysctl -w fs.inotify.max_user_instances=512
}

function get_helm_certs {
    log::info "Downloading Helm certificates from cluster"
    "${SCRIPT_DIR}/cluster-integration/helpers/get-helm-certs.sh"
}

function create_test_resources {
    log::info "Create e2e upgrade test resources"

    injectTestingAddons

    if [  -f "$(helm home)/ca.pem" ]; then
        local HELM_ARGS="--tls"
    fi

    helm install "${UPGRADE_TEST_PATH}" \
        --name "${UPGRADE_TEST_RELEASE_NAME}" \
        --namespace "${UPGRADE_TEST_NAMESPACE}" \
        --timeout "${UPGRADE_TEST_HELM_TIMEOUT_SEC}" \
        --wait ${HELM_ARGS}

    prepareResult=$?
    if [ "${prepareResult}" != 0 ]; then
        echo "Helm install operation failed: ${prepareResult}"
        return "${prepareResult}"
    fi

    set +o errexit
    checkTestPodTerminated
    prepareTestResult=$?
    set -o errexit

    echo "Logs for prepare data operation to test e2e upgrade: "
    kubectl logs -n "${UPGRADE_TEST_NAMESPACE}" -l "${UPGRADE_TEST_RESOURCE_LABEL}=${UPGRADE_TEST_LABEL_VALUE_PREPARE}" -c "${TEST_CONTAINER_NAME}"
    if [ "${prepareTestResult}" != 0 ]; then
        echo "Exit status for prepare upgrade e2e tests: ${prepareTestResult}"
        return "${prepareTestResult}"
    fi
}

remove_addons_if_necessary() {
    tdWithAddon=$(kubectl get td --all-namespaces -l testing.kyma-project.io/require-testing-addon=true -o custom-columns=NAME:.metadata.name --no-headers=true)

    if [ -z "$tdWithAddon" ]
    then
        echo "- Removing ClusterAddonsConfiguration which provides the testing addons"
        removeTestingAddons
        if [[ $? -eq 1 ]]; then
            return 1
        fi
    else
        echo "- Skipping removing ClusterAddonsConfiguration"
    fi
}

# main is the entrypoint function
function main {
    trap junit::test_fail ERR
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
        kubernetes::ensure_kubectl "${KUBERNETES_VERSION}" "$(host::os)" "${TMP_BIN_DIR}" 2>&1 | junit::test_output
        junit::test_pass
    else
        junit::test_skip
    fi

    junit::test_start "Ensure_Helm_Client"
    if [[ ${ENSURE_HELM} = "true" ]]; then
        log::info "Ensuring Helm client version" 2>&1 | junit::test_output
        helm::ensure_client "${HELM_VERSION}" "$(host::os)" "${TMP_BIN_DIR}" 2>&1 | junit::test_output
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

    junit::test_start "Get_Latest_Release"
    log::info "Checking latest Kyma release" 2>&1 | junit::test_output
    local -r latest_release="$(kyma::get_last_release_version "${BOT_GITHUB_TOKEN}" 2>&1 | junit::test_output)"
    log::info "Latest Kyma release is ${latest_release}" 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Install_Tiller_From_Release"
    log::info "Installing Tiller from version ${latest_release}" 2>&1 | junit::test_output
    kubectl apply -f "https://raw.githubusercontent.com/kyma-project/kyma/${latest_release}/installation/resources/tiller.yaml" 2>&1 | junit::test_output
    kubernetes::is_pod_ready "${KYMA_SOURCES}" kube-system name tiller 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Load_Kyma_Configuration_From_Release"
    log::info "Downloading Kyma Configuration from version ${latest_release}" 2>&1 | junit::test_output
    curl -L --silent --fail --show-error "https://raw.githubusercontent.com/kyma-project/kyma/${latest_release}/installation/resources/installer-config-kind.yaml.tpl" --output "${TMP_DIR}/installer-config-kind.yaml.tpl" 2>&1 | junit::test_output
    log::info "Loading Kyma Configuration from version ${latest_release}" 2>&1 | junit::test_output
    kyma::load_config "${worker_ip}" "${KYMA_DOMAIN}" "${TMP_DIR}/installer-config-kind.yaml.tpl" 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Install_Kyma_From_Release"
    log::info "Downloading Kyma artifacts from version ${latest_release}" 2>&1 | junit::test_output
    curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${latest_release}/kyma-installer-cluster.yaml" --output "${TMP_DIR}/last-release-installer.yaml"
    log::info "Installing Kyma from version ${latest_release}" 2>&1 | junit::test_output
    kubectl apply -f "${TMP_DIR}/last-release-installer.yaml" 2>&1 | junit::test_output
    kyma::is_installed "${KYMA_SOURCES}" "${KYMA_INSTALLATION_TIMEOUT}" 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Update_Hosts_File"
    if [[ ${UPDATE_HOSTS} = "true" ]]; then
        log::info "Updating hosts file" 2>&1 | junit::test_output
        kyma::update_hosts 2>&1 | junit::test_output
        junit::test_pass
    else
        junit::test_skip
    fi

    junit::test_start "Download_Helm_Certs"
    get_helm_certs 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Create_Test_Resources"
    create_test_resources 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Delete_Kyma_Installation_CR"
    # Remove the finalizer form kyma-installation the merge type is used because strategic is not supported on CRD.
    # More info about merge strategy can be found here: https://tools.ietf.org/html/rfc7386
    log::info "Deleting the kyma-installation CR" 2>&1 | junit::test_output
    kubectl patch Installation kyma-installation -n default --patch '{"metadata":{"finalizers":null}}' --type=merge 2>&1 | junit::test_output
    kubectl delete Installation -n default kyma-installation 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Build_Kyma_Installer"
    log::info "Building Kyma Installer Docker Image" 2>&1 | junit::test_output
    kyma::build_installer "${KYMA_SOURCES}" "${KYMA_INSTALLER_NAME}" 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Load_Kyma_Installer"
    log::info "Loading Kyma Installer to cluster" 2>&1 | junit::test_output
    kind::load_image "${CLUSTER_NAME}" "${KYMA_INSTALLER_NAME}" 2>&1 | junit::test_output
    junit::test_pass

    junit::test_start "Update_Tiller"
    log::info "Updating Tiller" 2>&1 | junit::test_output
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
        kyma::update_hosts 2>&1 | junit::test_output
        junit::test_pass
    else
        junit::test_skip
    fi

    junit::test_start "Remove_Addons"
    remove_addons_if_necessary 2>&1 | junit::test_output
    junit::test_pass

    # TODO(michal-hudy): not everything is covered with JUnit in the kyma-testing.sh script
    junit::test_start "Testing_Kyma"
    kyma::test "${SCRIPT_DIR}" 2>&1 | junit::test_output
    junit::test_pass
}

read_flags "$@"
main
