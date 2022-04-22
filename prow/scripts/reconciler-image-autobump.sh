#!/usr/bin/env bash

#Description: Generic image-autobump tool. This scripts implements a pipeline that updates the reconciler docker image tag in kyma-project/control-plane
#
#Expected common vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - TEST_INFRA_SOURCES_DIR - directory path with kyma-project/test-infra sources
# - K8S_TEST_INFRA_SOURCES_DIR - directory path with kubernetes/test-infra sources
# - RECONCILER_DIR - directory path with Kyma-incubator/reconciler sources
# - CONTROL_PLANE_DIR - directory path with Kyma-project/control-plane sources
#
#Please look in each provider script for provider specific requirements

## ---------------------------------------------------------------------------------------
## Configurations and Variables
## ---------------------------------------------------------------------------------------

# exit on error, and raise error when variable is not set when used
set -e

ENABLE_TEST_LOG_COLLECTOR=false

# Exported variables
export K8S_TEST_INFRA_SOURCES_DIR="/home/prow/go/src/github.com/kubernetes/test-infra"
export RECONCILER_DIR="/home/prow/go/src/github.com/kyma-incubator/reconciler"
export CONTROL_PLANE_DIR="/home/prow/go/src/github.com/kyma-project/control-plane"
export KYMA_TEST_INFRA_SOURCES_DIR="/home/prow/go/src/github.com/kyma-project/test-infra"
export BUMP_TOOL_CONFIG_FILE="${KYMA_TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/control-plane-autobump-reconciler-config.yaml"

# All provides require these values, each of them may check for additional variables
requiredVars=(
    K8S_TEST_INFRA_SOURCES_DIR
    RECONCILER_DIR
    CONTROL_PLANE_DIR
    BUMP_TOOL_CONFIG_FILE
)

# Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD

## ---------------------------------------------------------------------------------------
## Prow job functions definitions
## ---------------------------------------------------------------------------------------

# log::info prints message with info level
#
# Arguments:
#   $* - Message
function log::info {
    echo -e "$(date +"%Y/%m/%d %T %Z") [INFO] $*"
}

# utils::check_required_vars checks if all provided variables are initialized
#
# Arguments
# $1 - list of variables
function utils::check_required_vars() {
    local discoverUnsetVar=false
    for var in "$@"; do
      if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
      fi
    done
    if [ "${discoverUnsetVar}" = true ] ; then
      exit 1
    fi
}

function autobump::build() {
  log::info "Building k8s image autobump tool"
  cd "${KYMA_TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/generic-autobumper"
  go build -o /tools/generic-autobumper
}

function reconciler::fetch_latest_image_tag() {
  log::info "Fetching latest reconciler commit ID"
  cd "${RECONCILER_DIR}"
  RECONCILER_COMMIT_ID="$(git rev-parse HEAD)"
  export RECONCILER_IMAGE_TAG="$(echo "${RECONCILER_COMMIT_ID}" | cut -c1-8)"
  log::info "Reconciler image tag: ${RECONCILER_IMAGE_TAG}"
}

function autobump::update_reconciler_image_tag(){
  log::info "Update reconciler image tag in control plane"
  cd "${CONTROL_PLANE_DIR}"
  yq e -i '(.global.images.mothership_reconciler_version ) |= "'${RECONCILER_IMAGE_TAG}'"' ./resources/kcp/values.yaml
  yq e -i '(.global.images.components.[] | select(has("version")).["version"] ) |= "'${RECONCILER_IMAGE_TAG}'"' ./resources/kcp/values.yaml
}

function autobump::commit_changes(){
  log::info "Commit changes..."
  cd "${CONTROL_PLANE_DIR}"
  git add resources/kcp/values.yaml
  git commit -m 'Bumping Reconciler:\n\nNo eu.gcr.io/kyma-project/incubator/reconciler/ changes.\n\n' '--author' 'Kyma Bot <kyma.bot@sap.com>'
}

function autobump::run() {
  log::info "Running image auto-bump tool for reconciler"
  cd "${CONTROL_PLANE_DIR}"
  /tools/generic-autobumper --config="${BUMP_TOOL_CONFIG_FILE}"
}

## ---------------------------------------------------------------------------------------
## Prow job execution steps
## ---------------------------------------------------------------------------------------
# check if all the required ENVs are defined
utils::check_required_vars "${requiredVars[@]}"

# build generic-autobumper project in kubernetes/test-infra
autobump::build

# fetch latest reconciler image tag from reconciler commit ID
reconciler::fetch_latest_image_tag

autobump::update_reconciler_image_tag

autobump::commit_changes

# run autobump tool to update reconciler image tag in kyma-project/control-plane
autobump::run

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
