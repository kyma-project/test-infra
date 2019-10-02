#!/usr/bin/env bash

set -eo pipefail

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# shellcheck disable=SC1090
source "${SCRIPT_DIR}/common.sh"

function finalize() {
    local -r EXIT_STATUS=$?
    saveTestSuite
    exit "${EXIT_STATUS}"
}

trap testFailed ERR
trap finalize EXIT

initTestSuite "buildKubernetesImage"

log "Building Kubernetes ${KUBERNETES_VERSION} image"

testStart "initializeDINDEnvironment"
startDocker 2>&1 | ${STORE_TEST_OUTPUT}
testPassed

testStart "authenticateToDockerRegistry"
authenticate 2>&1 | ${STORE_TEST_OUTPUT}
testPassed

testStart "cloneKubernetesRepository"
log "Cloning Kubernetes repository to ${GOPATH}/src/k8s.io/kubernetes" 2>&1 | ${STORE_TEST_OUTPUT}
mkdir -p "${GOPATH}/src/k8s.io/kubernetes" 2>&1 | ${STORE_TEST_OUTPUT_APPEND}
git clone --branch "${KUBERNETES_VERSION}" https://github.com/kubernetes/kubernetes.git "${GOPATH}/src/k8s.io/kubernetes" 2>&1 | ${STORE_TEST_OUTPUT_APPEND}
testPassed

testStart "buildKubernetesImage"
log "Building Kubernetes image as ${IMAGE_NAME}" 2>&1 | ${STORE_TEST_OUTPUT}
kind build node-image --type docker --image "${IMAGE_NAME}" 2>&1 | ${STORE_TEST_OUTPUT_APPEND}
testPassed

testStart "pushKubernetesImage"
log "Pushing ${IMAGE_NAME} to Docker registry" 2>&1 | ${STORE_TEST_OUTPUT}
docker push "${IMAGE_NAME}" 2>&1 | ${STORE_TEST_OUTPUT_APPEND}
testPassed

testStart "createCluster"
log "Creating cluster with image ${IMAGE_NAME}" 2>&1 | ${STORE_TEST_OUTPUT}
"${SCRIPT_DIR}/install-kyma.sh" --only-cluster --delete-cluster 2>&1 | ${STORE_TEST_OUTPUT_APPEND}
testPassed