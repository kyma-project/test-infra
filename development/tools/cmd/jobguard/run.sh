#!/usr/bin/env bash

# this is only temporary solution
echo "PULL NUMBER: ${PULL_NUMBER}"
echo "PULL SHA: ${PULL_PULL_SHA}"

env INITIAL_SLEEP_TIME=5s PULL_NUMBER=${PULL_NUMBER} PULL_SHA=${PULL_PULL_SHA} JOB_NAME_PATTERN="(pre-master-kyma-components-.*)|(pre-master-kyma-tests-.*)" ./job-guard