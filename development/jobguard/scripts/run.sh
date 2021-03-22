#!/bin/bash

ROOT_PATH=$( dirname "$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" )

KYMA_PROJECT_DIR=${KYMA_PROJECT_DIR:-"/home/prow/go/src/github.com/kyma-project"}
JOB_NAME_PATTERN=${JOB_NAME_PATTERN:-"(pre-master-kyma-components-.*)|(pre-master-kyma-tests-.*)|(pre-kyma-components-.*)|(pre-kyma-tests-.*)|(pre-main-kyma-components-.*)|(pre-main-kyma-tests-.*)"}
TIMEOUT=${JOBGUARD_TIMEOUT:-"15m"}
echo "KYMA PROJECT DIR ${KYMA_PROJECT_DIR}"
echo "COMMIT SHA: ${PULL_PULL_SHA}"
echo "JOBGUARD_TIMEOUT: ${TIMEOUT}"

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"

if [ -x "/prow-tools/jobguard" ]; then
  env GITHUB_TOKEN="${BOT_GITHUB_TOKEN}" \
      INITIAL_SLEEP_TIME=1m \
      TIMEOUT="${TIMEOUT}" \
      COMMIT_SHA="${PULL_PULL_SHA}" \
      JOB_NAME_PATTERN="${JOB_NAME_PATTERN}" \
      PROW_CONFIG_FILE="${TEST_INFRA_SOURCES_DIR}/prow/config.yaml" \
      PROW_JOBS_DIRECTORY="${TEST_INFRA_SOURCES_DIR}/prow/jobs" \
      GO111MODULE=on \
      /prow-tools/jobguard
else
  cd "${ROOT_PATH}/cmd/jobguard" || exit 1
  env GITHUB_TOKEN="${BOT_GITHUB_TOKEN}" \
      INITIAL_SLEEP_TIME=1m \
      TIMEOUT="${TIMEOUT}" \
      COMMIT_SHA="${PULL_PULL_SHA}" \
      JOB_NAME_PATTERN="${JOB_NAME_PATTERN}" \
      PROW_CONFIG_FILE="${TEST_INFRA_SOURCES_DIR}/prow/config.yaml" \
      PROW_JOBS_DIRECTORY="${TEST_INFRA_SOURCES_DIR}/prow/jobs" \
      GO111MODULE=on \
      go run main.go
fi
