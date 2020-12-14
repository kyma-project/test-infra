#!/usr/bin/env bash

set -o errexit

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"


# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"

#ENABLE_TEST_LOG_COLLECTOR=false
#TEST_LOG_COLLECTOR_PROW_JOB_NAME="post-master-kyma-gke-upgrade"
#
#discoverUnsetVar=false
#for var in REPO_OWNER \
#  REPO_NAME \
#  DOCKER_PUSH_REPOSITORY \
#  KYMA_PROJECT_DIR \
#  CLOUDSDK_CORE_PROJECT \
#  CLOUDSDK_COMPUTE_REGION \
#  CLOUDSDK_DNS_ZONE_NAME \
#  GOOGLE_APPLICATION_CREDENTIALS \
#  KYMA_ARTIFACTS_BUCKET \
#  BOT_GITHUB_TOKEN \
#  DOCKER_IN_DOCKER_ENABLED \
#  GCR_PUSH_GOOGLE_APPLICATION_CREDENTIALS; do
#  if [[ -z "${!var}" ]]; then
#    echo "ERROR: $var is not set"
#    discoverUnsetVar=true
#  fi
#done
#if [[ "${discoverUnsetVar}" = true ]]; then
#  exit 1
#fi
#
#if [[ "${BUILD_TYPE}" == "master" ]]; then
#  if [ -z "${LOG_COLLECTOR_SLACK_TOKEN}" ]; then
#    echo "ERROR: LOG_COLLECTOR_SLACK_TOKEN is not set"
#    exit 1
#  fi
#fi
#
##Exported variables
#export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
#export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
#export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
#export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
#export KYMA_INSTALL_TIMEOUT="30m"
#export KYMA_UPDATE_TIMEOUT="40m"
#export UPGRADE_TEST_PATH="${KYMA_SOURCES_DIR}/tests/end-to-end/upgrade/chart/upgrade"
#export UPGRADE_TEST_NAMESPACE="e2e-upgrade-test"
#export UPGRADE_TEST_RELEASE_NAME="${UPGRADE_TEST_NAMESPACE}"
#export UPGRADE_TEST_RESOURCE_LABEL="kyma-project.io/upgrade-e2e-test"
#export EXTERNAL_SOLUTION_TEST_PATH="${KYMA_SOURCES_DIR}/tests/end-to-end/external-solution-integration/chart/external-solution"
#export EXTERNAL_SOLUTION_TEST_NAMESPACE="integration-test"
#export EXTERNAL_SOLUTION_TEST_RELEASE_NAME="${EXTERNAL_SOLUTION_TEST_NAMESPACE}"
#export EXTERNAL_SOLUTION_TEST_RESOURCE_LABEL="kyma-project.io/external-solution-e2e-test"
#export TEST_RESOURCE_LABEL_VALUE_PREPARE="prepareData"
#export HELM_TIMEOUT_SEC=10000s # timeout in sec for helm install/test operation
#export TEST_TIMEOUT_SEC=600    # timeout in sec for test pods until they reach the terminating state
#export TEST_CONTAINER_NAME="tests"
#
#TMP_DIR=$(mktemp -d)
#KYMA_LABEL_PREFIX="kyma-project.io"
#KYMA_TEST_LABEL_PREFIX="${KYMA_LABEL_PREFIX}/test"
#BEFORE_UPGRADE_LABEL_QUERY="${KYMA_TEST_LABEL_PREFIX}.before-upgrade=true"
#POST_UPGRADE_LABEL_QUERY="${KYMA_TEST_LABEL_PREFIX}.after-upgrade=true"

TMP_DIR=$(mktemp -d)

CLOUD_FUNCTION_URL="https://europe-west3-kyma-project.cloudfunctions.net/get-scan-result"
RESPONSE_FILE=${TMP_DIR}/response.json

if [[ "${BUILD_TYPE}" == "pr" ]]; then
  log::info "Execute Job Guard"
  "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

PR_NAME="PR-${PULL_NUMBER}"
echo "Protecode scan result for ${PR_NAME}:"
curl -s "${CLOUD_FUNCTION_URL}?tag=${PR_NAME}" -H "Content-Type:application/json" > ${RESPONSE_FILE}
cat ${RESPONSE_FILE} | jq '.'

SUCCESS=$(cat ${RESPONSE_FILE} | jq '.success')

if [[ "$SUCCESS" == "true" ]]; then
   log::success "All images are green!"
   exit 0
fi

log::error "Some images contain security vulnerabilities"
log::error "For more details please check json output"
exit -1

