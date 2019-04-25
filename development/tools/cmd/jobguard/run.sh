#!/usr/bin/env bash

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

KYMA_PROJECT_DIR=${KYMA_PROJECT_DIR:-"/home/prow/go/src/github.com/kyma-project"}
echo "KYMA PROJECT DIR ${KYMA_PROJECT_DIR}"
echo "PULL NUMBER: ${PULL_NUMBER}"
echo "PULL SHA: ${PULL_PULL_SHA}"
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"

cd ${ROOT_PATH}

env GITHUB_TOKEN=${GITHUB_TOKEN} \
    INITIAL_SLEEP_TIME=1m \
    PULL_NUMBER=${PULL_NUMBER} \
    PULL_SHA=${PULL_PULL_SHA} \
    JOB_NAME_PATTERN="(pre-master-kyma-components-.*)|(pre-master-kyma-tests-.*)" \
    PROW_CONFIG_FILE="${TEST_INFRA_SOURCES_DIR}/prow/config.yaml" \
    PROW_JOBS_DIRECTORY="${TEST_INFRA_SOURCES_DIR}/prow/jobs" \
    ./job-guard