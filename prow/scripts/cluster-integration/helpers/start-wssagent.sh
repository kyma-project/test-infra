#!/usr/bin/env bash

#Description: runs wss-unified-agent
#The purpose is to run the wss-unified-agent

#Expected vars:
# - WS_APIKEY- Key provided by SAP Whitesource Team
# - WS_USERKEY - Users specified key(should be a service account)
# - WS_PRODUCTNAME - Product inside whitesource
# - REPOSITORY - Component name, scans that directory and posts the results in whitesource
# - GITHUB_ORG_DIR - Project directory to scan
# - SCAN_LANGUAGE - Scan language is used to set the correct values in the whitesource config for golang / golang-mod / javascript
# Optional vars:
# - CUSTOM_PROJECTNAME - Use custom project name instead of REPOSITORY
# - CREATE_SUBPROJECTS - Find all projects/modules based on the SCAN_LANGUAGE and scan each to a separate Whitesource project

set -e
export TEST_INFRA_SOURCES_DIR="/home/prow/go/src/github.com/kyma-project/test-infra"
# shellcheck source=prow/scripts/lib/log.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/log.sh"

# whitesource config
GO_DEP_CONFIG_PATH="$TEST_INFRA_SOURCES_DIR/prow/images/whitesource-scanner/go-wss-unified-agent.config"
GO_MOD_CONFIG_PATH="$TEST_INFRA_SOURCES_DIR/prow/images/whitesource-scanner/go-mod-wss-unified-agent.config"
JAVASCRIPT_CONFIG_PATH="$TEST_INFRA_SOURCES_DIR/prow/images/whitesource-scanner/javascript-wss-unified-agent.config"


export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

# shellcheck disable=SC2153
PROJECT_SRC="${GITHUB_ORG_DIR}/${REPOSITORY}"

PROJECTNAME="${REPOSITORY}"
if [[ $CUSTOM_PROJECTNAME != "" ]]; then
  PROJECTNAME="${CUSTOM_PROJECTNAME}"
fi

# pass values to Whitesource binary through WS_* variables
export WS_USERKEY=$(cat "${WHITESOURCE_USERKEY}")
export WS_APIKEY=$(cat "${WHITESOURCE_APIKEY}")

# don't stop scans on first failure, but fail the whole job after all scans have finished
export scan_failed

#exclude components based on dependency management
function filterFolders() {
  local DEPENDENCY_FILE_TO_EXCLUDE
  DEPENDENCY_FILE_TO_EXCLUDE=$1
  local FOLDER_TO_SCAN
  FOLDER_TO_SCAN=$2
  local EXCLUDES
  EXCLUDES=$({ cd "${FOLDER_TO_SCAN}" && find . -iname "${DEPENDENCY_FILE_TO_EXCLUDE}"; } | grep -v vendor | grep -v tests | xargs -n 1 dirname | sed 's/$/\/**/' | sed 's/^.\//**\//' | paste -s -d" " -)
  EXCLUDES="excludes=**/docs/** **/tests/** ${EXCLUDES}"
  echo "$EXCLUDES"
}

function prepareDepDependencies() {
  local DEPENDENCY_FILE
  DEPENDENCY_FILE=$1
  local FOLDER_TO_SCAN
  FOLDER_TO_SCAN=$2

  for COMPFOLDER in $({ find "${FOLDER_TO_SCAN}" -iname "${DEPENDENCY_FILE}"; } | grep -v vendor | grep -v tests | xargs -n 1 dirname); do
    {
      echo "$COMPFOLDER"
      cd "$COMPFOLDER"
      # a little trick to enforce `dep ensure` over `dep init`
      mkdir -p vendor
    }
  done
}

case "${SCAN_LANGUAGE}" in
golang)
  echo "SCAN: golang (dep)"
  go version
  CONFIG_PATH=$GO_DEP_CONFIG_PATH
  COMPONENT_DEFINITION="Gopkg.toml"
  EXCLUDE_PROJECT_CONFIG="go.mod"
  prepareDepDependencies gopkg.toml "${PROJECT_SRC}"
  ;;

golang-mod)
  echo "SCAN: golang-mod"
  go version
  CONFIG_PATH=$GO_MOD_CONFIG_PATH
  export GO111MODULE=on
  EXCLUDE_PROJECT_CONFIG="Gopkg.toml"
  COMPONENT_DEFINITION="go.mod"
  ;;

javascript)
  echo "SCAN: javascript"
  CONFIG_PATH=$JAVASCRIPT_CONFIG_PATH
  COMPONENT_DEFINITION="package.json"
  ;;

*)
  echo "can only be golang or javascript"
  exit 1
  ;;
esac

echo "scanComment=$(date)" >> $CONFIG_PATH

log::banner "Starting Scan"

# scanFolder scans single folder to a Whitesource project
# parameters:
# $1 - path to a folder to scan
# $2 - name of the Whitesource project
# variables:
# WS_PRODUCTNAME - name of the Whitesource product
# EXCLUDE_PROJECT_CONFIG (optional) - name of the config fileson which folders should be excluded
# DRYRUN (optional) - don't run the Whitesource unified agent binary
# function returns 0 on success or 1 on fail
function scanFolder() { # expects to get the fqdn of folder passed to scan
  if [[ $1 == "" ]]; then
    echo "path cannot be empty"
    exit 1
  fi
  FOLDER=$1
  if [[ $2 == "" ]]; then
    echo "component name cannot be empty"
    exit 1
  fi
  pushd "${FOLDER}" # change to passed parameter
  WS_PROJECTNAME=$2
  export WS_PROJECTNAME

  if [[ -n "$EXCLUDE_PROJECT_CONFIG" ]]; then
    export WS_EXCLUDES=$(filterFolders "${EXCLUDE_PROJECT_CONFIG}" "$(pwd)")
  fi

  # shellcheck disable=SC2153
  echo "Product name - $WS_PRODUCTNAME"
  echo "Project name - $WS_PROJECTNAME"

  if [ "${DRYRUN}" = false ]; then
    log::banner "Scanning $FOLDER"
    if [ -z "$JAVA_OPTS" ]; then
      echo "no additional java_opts set"
      java -jar /wss/wss-unified-agent.jar -c $CONFIG_PATH
      scan_result="$?"
    else
      echo "Java Options - '$JAVA_OPTS'"
      java "${JAVA_OPTS}" -jar /wss/wss-unified-agent.jar -c $CONFIG_PATH
      scan_result="$?"
    fi
  else
    log::banner "DRYRUN Successful for $FOLDER"
  fi
  popd
  if [[ "$scan_result" != 0 ]]; then
    return 1
  else
    return 0
  fi
}


if [[ "$CREATE_SUBPROJECTS" == "true" ]]; then
  # treat every found Go / JS project as a separate Whitesource project
  pushd "${PROJECT_SRC}" # change to passed parameter

  # find all go.mod / Gopkg.toml / package.json projects and scan them individually
  while read -r component_definition_path; do
    # TODO what about excludes?
    # remove go.mod / Gopkg.toml part
    component_path="${component_definition_path%/*}"
    # keep only the last diretrory in the tree as a name
    component="${component_path##*/}"

    set +e
    scanFolder "${component_path}" "${PROJECTNAME}-${component}"
    scan_result="$?"
    set -e

    if [[ "$scan_result" -ne 0 ]]; then
      log::error "Scan for ${FOLDER} has failed"
      scan_failed=1
    fi
  done <<< "$(find . -name "$COMPONENT_DEFINITION" -not -path "./tests/*" -not -path "./docs/*")"
  popd
else
  # scan PROJECT_SRC directory as a single project
  set +e
  scanFolder "${PROJECT_SRC}" "${PROJECTNAME}"
  scan_result="$?"
  set -e

  if [[ "$scan_result" -ne 0 ]]; then
    log::error "Scan for ${PROJECT_SRC} has failed"
    scan_failed=1
  fi
fi

if [[ "$scan_failed" -eq 1 ]]; then
  log::error "One or more of the scans have failed"
  exit 1
else
  log::banner "Scanning Finished"
fi
