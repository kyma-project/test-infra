#!/usr/bin/env bash

# https://github.com/kubernetes-sigs/kind/issues/303
# https://github.com/kubernetes-sigs/kind/issues/717
# https://github.com/kubernetes-sigs/kind/issues/759

set -eo pipefail

readonly ARGS=("$@")
readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly CLUSTER_DIR="$( cd "${SCRIPT_DIR}/../cluster" && pwd )"
readonly CONFIG_DIR="$( cd "${SCRIPT_DIR}/../config" && pwd )"
readonly KYMA_SOURCES="${GOPATH}/src/github.com/kyma-project/kyma"
readonly KYMA_INSTALLER="kyma-installer:master"
readonly DOMAIN="kyma.local"
readonly HOSTNAMES="apiserver.${DOMAIN} console.${DOMAIN} catalog.${DOMAIN} instances.${DOMAIN} brokers.${DOMAIN} dex.${DOMAIN} docs.${DOMAIN} addons.${DOMAIN} lambdas-ui.${DOMAIN} console-backend.${DOMAIN} minio.${DOMAIN} jaeger.${DOMAIN} grafana.${DOMAIN} log-ui.${DOMAIN} loki.${DOMAIN} configurations-generator.${DOMAIN} gateway.${DOMAIN} connector-service.${DOMAIN} oauth2.${DOMAIN} oauth2-admin.${DOMAIN} kiali.${DOMAIN} compass-gateway.${DOMAIN} compass-gateway-mtls.${DOMAIN} compass-gateway-auth-oauth.${DOMAIN} compass.${DOMAIN} compass-mf.${DOMAIN}"
readonly INSTALLATION_TIMEOUT="${INSTALLATION_TIMEOUT:-"30m"}"
readonly LOCAL_PATH_PROVISIONER_VERSION="${LOCAL_PATH_PROVISIONER_VERSION:-"v0.0.11"}"

CLUSTER_IP=127.0.0.1
CLUSTER_PROVISIONED="false"

ENSURE_KUBECTL="false"
UPDATE_HOSTS="false"
DELETE_CLUSTER="false"
RUN_TESTS="false"
ONLY_CLUSTER="false"
SETUP_INOTIFY="false"

# shellcheck disable=SC1090
source "${SCRIPT_DIR}/common.sh"

function readFlags() {
    while test $# -gt 0; do
        case "$1" in
            -h|--help)
                shift
                echo "Script that installs Kyma on kind"
                echo " "
                echo "Options:"
                echo "  -h --help            Print usage."
                echo "     --ensure-kubectl  Update kubectl to the same version as cluster."
                echo "     --update-hosts    Append hosts file with Kyma hosts assigned to worker node."
                echo "     --delete-cluster  Deletes cluster at the end of script execution."
                echo "     --run-tests       Run Kyma integration tests after Kyma installation."
                echo "     --only-cluster    Creates only cluster without Kyma."
                echo "     --setup-inotify   Setup inotify instances and watches."
                echo " "
                echo "Environment variables:"
                echo "  BUILD_TYPE   If set to \"pr\" then Kubernetes image from Pull Request will be used, PULL_NUMBER variable is required"
                echo "  PULL_NUMBER  Number of Pull Request with Kubernetes image"
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
            --run-tests)
                shift
                RUN_TESTS="true"
                ;;
            --only-cluster)
                shift
                ONLY_CLUSTER="true"
                ;;
            --setup-inotify)
                shift
                SETUP_INOTIFY="true"
                ;;
            *)
                log "$1 is not a recognized flag, use --help flag for a list of avaiable options!"
                return 1
                ;;
        esac
    done

    readonly ENSURE_KUBECTL UPDATE_HOSTS DELETE_CLUSTER RUN_TESTS ONLY_CLUSTER
}

function finalize() {
    local -r EXIT_STATUS=$?
    local finalizationError="false"

    testStart "finalization"
    log "Finalization" 2>&1 | ${STORE_TEST_OUTPUT}

    log "Print all docker processes" 2>&1 | ${STORE_TEST_OUTPUT_APPEND}
    docker ps -a 2>&1 | ${STORE_TEST_OUTPUT_APPEND}

    if [[ ${CLUSTER_PROVISIONED} = "true" ]]; then
        log "Exporting cluster logs to ${ARTIFACTS_DIR}/cluster-logs" 2>&1 | ${STORE_TEST_OUTPUT_APPEND}
        mkdir -p "${ARTIFACTS_DIR}/cluster-logs" 2>&1 | ${STORE_TEST_OUTPUT_APPEND} || finalizationError="true"
        kind export logs "${ARTIFACTS_DIR}/cluster-logs" 2>&1 | ${STORE_TEST_OUTPUT_APPEND} || finalizationError="true"

        log "Creating archive ${ARTIFACTS_DIR}/cluster-logs.tar.gz with cluster logs"
        (cd "${ARTIFACTS_DIR}" && tar -zcf "${ARTIFACTS_DIR}/cluster-logs.tar.gz" "cluster-logs/") 2>&1 | ${STORE_TEST_OUTPUT_APPEND} || finalizationError="true"
        
        if [[ ${DELETE_CLUSTER} = "true" ]]; then
            log "Deleting cluster" 2>&1 | ${STORE_TEST_OUTPUT_APPEND}
            kind delete cluster 2>&1 | ${STORE_TEST_OUTPUT_APPEND} || finalizationError="true"  # https://github.com/kubernetes-sigs/kind/issues/759
        fi
    fi

    set +eo pipefail
    if [[ ${finalizationError} = "true" ]]; then
        testFailed
    else
        testPassed
    fi
    set -eo pipefail

    saveTestSuite
    return "${EXIT_STATUS}"
}

function setupInotify() {
    sysctl -w fs.inotify.max_user_watches=524288
    sysctl -w fs.inotify.max_user_instances=512
}

function ensureExpectedKubectlVersion() {
    if command -v kubectl >/dev/null 2>&1; then
        log "Removing built-in kubectl version"
        rm -rf "$(command -v kubectl)"
    fi

    log "Install kubectl in version ${KUBERNETES_VERSION}"
    curl -LO "https://storage.googleapis.com/kubernetes-release/release/${KUBERNETES_VERSION}/bin/linux/amd64/kubectl" --fail \
        && chmod +x kubectl \
        && mv kubectl /usr/local/bin/kubectl
}

function createCluster() {
    log "Create kind cluster"
    kind create cluster --config "${CLUSTER_DIR}/cluster.yaml" --wait 3m --image "${IMAGE_NAME}"
    readonly KUBECONFIG="$(kind get kubeconfig-path --name="kind")"
    cp "${KUBECONFIG}" "${HOME}/.kube/config"
    kubectl cluster-info
}

function updateHostsFile() {
    log "Update /etc/hosts with cluster IP"
    echo "${CLUSTER_IP} ${HOSTNAMES}" | tee -a /etc/hosts > /dev/null
}

function buildKymaImage() {
    log "Build Kyma Installer image"
    docker build -t "${KYMA_INSTALLER}" -f "${KYMA_SOURCES}/tools/kyma-installer/kyma.Dockerfile" "${GOPATH}/src/github.com/kyma-project/kyma"
}

function loadKymaImage() {
    log "Load Kyma Installer image to the cluster"
    kind load docker-image "${KYMA_INSTALLER}"
}

function installDefaultResources() {
    log "Make kubernetes.io/host-path Storage Class as non default"
    kubectl annotate storageclass standard storageclass.kubernetes.io/is-default-class="false" storageclass.beta.kubernetes.io/is-default-class="false" --overwrite

    log "Install default resources from ${CLUSTER_DIR}/resources/"
    kubectl apply -f "${CLUSTER_DIR}/resources/"
}

function installTiller() {
    log "Install Tiller"
    "${KYMA_SOURCES}/installation/scripts/install-tiller.sh"
}

function loadKymaConfiguration() {
    log "Create kyma-installer namespace"
    kubectl create namespace "kyma-installer"

    log "Install Kyma configuration"
    < "${CONFIG_DIR}/overrides.yaml" sed 's/\.minikubeIP: .*/\.minikubeIP: '"${CLUSTER_IP}"'/g' \
        | sed 's/\.domainName: .*/\.domainName: '"${DOMAIN}"'/g' \
        | kubectl apply -f-
}

function triggerInstallation() {
    log "Trigger installation"
    "${KYMA_SOURCES}/installation/scripts/concat-yamls.sh" "${KYMA_SOURCES}/installation/resources/installer.yaml" "${KYMA_SOURCES}/installation/resources/installer-cr-cluster.yaml.tpl" \
        | sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER};" \
        | sed -e "s/__VERSION__/0.0.1/g" \
        | sed -e "s/__.*__//g" \
        | kubectl apply -f-
}

function waitForInstallationFinish() {
    log "Waiting for Kyma installation to finish"
    "${KYMA_SOURCES}/installation/scripts/is-installed.sh" --timeout "${INSTALLATION_TIMEOUT}"
}

function testKyma() {
    log "Starting Kyma Integration tests"
    "${KYMA_SOURCES}/installation/scripts/testing.sh" --cleanup "false" --concurrency 5
}

readFlags "${ARGS[@]}"

trap testFailed ERR
trap finalize EXIT

initTestSuite "installKymaOnKind"

log "Cluster will be deployed with Kubernetes ${KUBERNETES_VERSION}"

testStart "setupInotify"
if [[ ${SETUP_INOTIFY} = "true" ]]; then
    setupInotify 2>&1 | ${STORE_TEST_OUTPUT}
    testPassed    
else
    testSkipped "Disabled"
fi

testStart "initializeDINDEnvironment"
if docker info > /dev/null 2>&1 ; then
    testSkipped "DIND already enabled"
else
    startDocker 2>&1 | ${STORE_TEST_OUTPUT}
    testPassed
fi

testStart "ensureExpectedKubectlVersion"
if [[ ${ENSURE_KUBECTL} = "true" ]]; then
    ensureExpectedKubectlVersion 2>&1 | ${STORE_TEST_OUTPUT}
    testPassed
else
    testSkipped "Disabled"
fi

testStart "createKindCluster"
createCluster 2>&1 | ${STORE_TEST_OUTPUT}
CLUSTER_PROVISIONED="true"
CLUSTER_IP="$(docker inspect -f "{{ .NetworkSettings.IPAddress }}" kind-worker)"
testPassed

testStart "installDefaultResources"
installDefaultResources 2>&1 | ${STORE_TEST_OUTPUT}
testPassed

if [[ ${ONLY_CLUSTER} = "true" ]]; then
    exit 0
fi

testStart "updateHostsFile"
if [[ ${UPDATE_HOSTS} = "true" ]]; then
    updateHostsFile 2>&1 | ${STORE_TEST_OUTPUT}
    testPassed    
else
    testSkipped "Disabled"
fi

testStart "buildKymaImage"
buildKymaImage 2>&1 | ${STORE_TEST_OUTPUT}
testPassed

testStart "loadKymaImageIntoCluster"
loadKymaImage 2>&1 | ${STORE_TEST_OUTPUT}
testPassed

testStart "installTiller"
installTiller 2>&1 | ${STORE_TEST_OUTPUT}
testPassed

testStart "loadKymaConfiguration"
loadKymaConfiguration 2>&1 | ${STORE_TEST_OUTPUT}
testPassed

testStart "triggerKymaInstallation"
triggerInstallation 2>&1 | ${STORE_TEST_OUTPUT}
testPassed

testStart "waitForInstallationFinish"
waitForInstallationFinish 2>&1 | ${STORE_TEST_OUTPUT}
testPassed

testStart "integrationTests"
if [[ ${RUN_TESTS} = "true" ]]; then
    testKyma 2>&1 | ${STORE_TEST_OUTPUT}
    testPassed
else
    testSkipped "Disabled"
fi