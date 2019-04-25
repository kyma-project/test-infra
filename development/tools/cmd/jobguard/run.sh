#!/usr/bin/env bash

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# this is only temporary solution
echo "PULL NUMBER: ${PULL_NUMBER}"
echo "PULL SHA: ${PULL_PULL_SHA}"

env GITHUB_TOKEN=${GITHUB_TOKEN} \
    INITIAL_SLEEP_TIME=5s \
    PULL_NUMBER=${PULL_NUMBER} \
    PULL_SHA=${PULL_PULL_SHA} \
    JOB_NAME_PATTERN="(pre-master-kyma-components-.*)|(pre-master-kyma-tests-.*)" \
    PROW_CONFIG_FILE="/Users/i303785/go/src/github.com/kyma-project/test-infra/prow/config.yaml" \
    PROW_JOBS_DIRECTORY="/Users/i303785/go/src/github.com/kyma-project/test-infra/prow/jobs" \
     go run  ${ROOT_PATH}/main.go