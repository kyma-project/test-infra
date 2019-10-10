#!/usr/bin/env bash

readonly KUBERNETES_VERSION=$( cat "${SCRIPT_DIR}/../KIND_KUBERNETES_VERSION" )
if [[ "${BUILD_TYPE}" == "pr" ]]; then
    readonly IMAGE_NAME="${DOCKER_PUSH_REPOSITORY:-"eu.gcr.io/kyma-project/test-infra"}${DOCKER_PUSH_DIRECTORY:-"/pr"}/kubernetes-kind:PR-${PULL_NUMBER}"
else
    readonly IMAGE_NAME="${DOCKER_PUSH_REPOSITORY:-"eu.gcr.io/kyma-project/test-infra"}${DOCKER_PUSH_DIRECTORY}/kubernetes-kind:${KUBERNETES_VERSION}"
fi
ARTIFACTS_DIR="${ARTIFACTS:-"$( cd "${SCRIPT_DIR}/.." && pwd )/tmp"}" # ARTIFACTS environment variable is provided by Prow
if [[ ! -d ${ARTIFACTS_DIR} ]]; then
    mkdir -p "${ARTIFACTS_DIR}"
fi
readonly ARTIFACTS_DIR="$( cd "${ARTIFACTS_DIR}" && pwd )"
readonly STORE_TEST_OUTPUT="tee ${ARTIFACTS_DIR}/lastTestOutput.log"
readonly STORE_TEST_OUTPUT_APPEND="tee -a ${ARTIFACTS_DIR}/lastTestOutput.log"

function log() {
    echo "$(date +"%Y/%m/%d %T %Z"): ${1}"
}

function startDocker() {
    log "Docker in Docker enabled, initializing..."
    printf '=%.0s' {1..80}; echo
    # If we have opted in to docker in docker, start the docker daemon,
    service docker start
    # the service can be started but the docker socket not ready, wait for ready
    local WAIT_N=0
    local MAX_WAIT=20
    while true; do
        # docker ps -q should only work if the daemon is ready
        docker ps -q > /dev/null 2>&1 && break
        if [[ ${WAIT_N} -lt ${MAX_WAIT} ]]; then
            WAIT_N=$((WAIT_N+1))
            log "Waiting for docker to be ready, sleeping for ${WAIT_N} seconds."
            sleep ${WAIT_N}
        else
            log "Reached maximum attempts, not waiting any longer..."
            exit 1
        fi
    done
    printf '=%.0s' {1..80}; echo

    docker-credential-gcr configure-docker
    log "Done setting up docker in docker."
}

function authenticate() {
    log "Authenticating"
    gcloud auth activate-service-account --key-file "${GOOGLE_APPLICATION_CREDENTIALS}"
}

TESTS_COUNT=0
FAILED_TESTS_COUNT=0
SKIPPED_TESTS_COUNT=0
CURRENT_TEST_CASE=
CURRENT_TEST_START_TIME=
TEST_SUITE_START_TIME=
TEST_SUITE_NAME=

function initTestSuite() {
    TEST_SUITE_START_TIME=$(date +%s)
    TEST_SUITE_NAME="${1}"
    readonly TEST_SUITE_NAME TEST_SUITE_START_TIME
    rm -rf "${ARTIFACTS_DIR}/junit_${TEST_SUITE_NAME}_suite.xml"
    rm -rf "${ARTIFACTS_DIR}/lastTestOutput.log"
}

function testStart() {
    TESTS_COUNT=$((++TESTS_COUNT))
    CURRENT_TEST_CASE="${TEST_SUITE_NAME}/${1}"
    CURRENT_TEST_START_TIME=$(date +%s)
    echo "=== RUN: ${CURRENT_TEST_CASE} 🤔"
}

function testPassed() {
    local -r duration=$(($(date +%s)-CURRENT_TEST_START_TIME))
    echo "--- PASS: ${CURRENT_TEST_CASE} (${duration}s) 😍"
    echo "        <testcase name=\"${CURRENT_TEST_CASE}\" time=\"${duration}\"></testcase>" >> "${ARTIFACTS_DIR}/junit_${TEST_SUITE_NAME}_suite.xml"
}

function testFailed() {
    local -r duration=$(($(date +%s)-CURRENT_TEST_START_TIME))
    FAILED_TESTS_COUNT=$((++FAILED_TESTS_COUNT))
    echo "--- FAIL: ${CURRENT_TEST_CASE} (${duration}s) 💩"
    echo "        <testcase name=\"${CURRENT_TEST_CASE}\" time=\"${duration}\"><failure>$(< "${ARTIFACTS_DIR}/lastTestOutput.log" tr -cd '\11\12\15\40-\176' | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g; s/"/\&quot;/g; s/'"'"'/\&#39;/g')</failure></testcase>" >> "${ARTIFACTS_DIR}/junit_${TEST_SUITE_NAME}_suite.xml"
    return 1
}

function testSkipped() {
    local -r duration=$(($(date +%s)-CURRENT_TEST_START_TIME))
    SKIPPED_TESTS_COUNT=$((++SKIPPED_TESTS_COUNT))
    echo "--- SKIP: ${CURRENT_TEST_CASE} (${duration}s) 🙈"
    echo "        <testcase name=\"${CURRENT_TEST_CASE}\" time=\"${duration}\"><skipped>${1}</skipped></testcase>" >> "${ARTIFACTS_DIR}/junit_${TEST_SUITE_NAME}_suite.xml"
}

function saveTestSuite() {
    local -r suiteDuration=$(($(date +%s)-TEST_SUITE_START_TIME))
    echo "<testsuite failures=\"${FAILED_TESTS_COUNT}\" name=\"${TEST_SUITE_NAME}\" skipped=\"${SKIPPED_TESTS_COUNT}\" tests=\"${TESTS_COUNT}\" time=\"${suiteDuration}\">
$(cat "${ARTIFACTS_DIR}/junit_${TEST_SUITE_NAME}_suite.xml")
</testsuite>" > "${ARTIFACTS_DIR}/junit_${TEST_SUITE_NAME}_suite.xml"
}