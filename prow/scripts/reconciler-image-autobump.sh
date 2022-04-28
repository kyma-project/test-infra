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
  cd "${KYMA_TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/autobumper"
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
  # support old image tag update, should be removed after PR https://github.com/kyma-project/control-plane/pull/1601 merged.
  if $(yq eval '.global.images | has("mothership_reconciler")' ./resources/kcp/values.yaml); then
    yq e -i '(.global.images.mothership_reconciler ) = "eu.gcr.io/kyma-project/incubator/reconciler/mothership:'${RECONCILER_IMAGE_TAG}'"' ./resources/kcp/values.yaml
    yq e -i '(.global.images.component_reconciler ) = "eu.gcr.io/kyma-project/incubator/reconciler/component:'${RECONCILER_IMAGE_TAG}'"' ./resources/kcp/values.yaml
  fi

  if $(yq eval '.global.images | has("mothership_reconciler_version")' ./resources/kcp/values.yaml); then
    yq e -i '(.global.images.mothership_reconciler_version ) = "'${RECONCILER_IMAGE_TAG}'"' ./resources/kcp/values.yaml
  fi
  if $(yq eval '.global.images | has("components")' ./resources/kcp/values.yaml); then
    yq e -i '(.global.images.components.[] | select(has("version")).["version"] ) = "'${RECONCILER_IMAGE_TAG}'"' ./resources/kcp/values.yaml
  fi
}

function autobump::commit_changes_and_create_pr(){
  log::info "Commit changes..."
  cd "${CONTROL_PLANE_DIR}"
  if [[ $(git status --porcelain) ]]; then
    git add resources/kcp/values.yaml
    git commit -m 'Bumping Reconciler:\n\nNo eu.gcr.io/kyma-project/incubator/reconciler/ changes.\n\n' '--author' 'Kyma Bot <kyma.bot@sap.com>'
    log::info "Create PR to control plane"
    /tools/generic-autobumper --config="${BUMP_TOOL_CONFIG_FILE}"
  else
    log::info "Nothing changed, stopped."
  fi
}

## ---------------------------------------------------------------------------------------
## Prow job execution steps
## ---------------------------------------------------------------------------------------
# check if all the required ENVs are defined
utils::check_required_vars "${requiredVars[@]}"

autobump::build

# fetch latest reconciler image tag from reconciler commit ID
reconciler::fetch_latest_image_tag

autobump::update_reconciler_image_tag

autobump::commit_changes_and_create_pr

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
