#!/usr/bin/env bash

#Description: runs wss-unified-agent
#The purpose is to run the wss-unified-agent

#Expected vars:
# - WS_APIKEY- Key provided by SAP Whitesource Team
# - WS_USERKEY - Users specified key(should be a service account)
# - WS_PRODUCTNAME - Product inside whitesource
# - SCAN_LANGUAGE - Scan language is used to set the correct values in the whitesource config for golang / golang-mod / javascript / python
# Optional vars:
# - CREATE_SUBPROJECTS - Find all projects/modules based on the SCAN_LANGUAGE and scan each to a separate Whitesource project
# - INTERNAL_GITHUB_TOKEN - Token to authenticate against internal github to fetch private go modules
# - INTERNAL_GITHUB_URL - URL for internal github to authenticate against using INTERNAL_GITHUB_TOKEN, required weh using INTERNAL_GITHUB_TOKEN

set -e

# whitesource config
GO_MOD_CONFIG_PATH="/wss/go-mod-wss-unified-agent.config"
JAVASCRIPT_CONFIG_PATH="/wss/javascript-wss-unified-agent.config"
PYTHON_CONFIG_PATH="/wss/python-wss-unified-agent.config"

if [[ -n "$INTERNAL_GITHUB_TOKEN" && -n "$INTERNAL_GITHUB_URL" ]]; then
  git config --global url."https://${INTERNAL_GITHUB_TOKEN}:x-oauth-basic@${INTERNAL_GITHUB_URL}/".insteadOf "https://${INTERNAL_GITHUB_URL}/"
  export GOPRIVATE=${INTERNAL_GITHUB_URL}/kyma
fi

if [[ -z "$PROJECT" ]]; then
  PROJECT="$REPO_NAME"
fi

# pass values to Whitesource binary through WS_* variables
if [[ -z "$WS_USERKEY" ]]; then
  export WS_USERKEY=$(cat "${WHITESOURCE_USERKEY}")
fi

if [[ -z "$WS_APIKEY" ]]; then
  export WS_APIKEY=$(cat "${WHITESOURCE_APIKEY}")
fi

# don't stop scans on first failure, but fail the whole job after all scans have finished
export scan_failed

case "${SCAN_LANGUAGE}" in
golang-mod)
  echo "SCAN: golang-mod"
  go version
  CONFIG_PATH=$GO_MOD_CONFIG_PATH
  export GO111MODULE=on
  COMPONENT_DEFINITION="go.mod"
  ;;

javascript)
  echo "SCAN: javascript"
  CONFIG_PATH=$JAVASCRIPT_CONFIG_PATH
  COMPONENT_DEFINITION="package.json"
  ;;

python)
  echo "SCAN: python"
  CONFIG_PATH=$PYTHON_CONFIG_PATH
  COMPONENT_DEFINITION="pyproject.toml"
  ;;

*)
  echo "can only be golang, javascript or python"
  exit 1
  ;;
esac

echo "scanComment=$(date)" >> "$CONFIG_PATH"

echo "💩 Starting Scan"

# scanFolder scans single folder to a Whitesource project
# parameters:
# $1 - path to a folder to scan
# $2 - name of the Whitesource project
# variables:
# WS_PRODUCTNAME - name of the Whitesource product
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


  if [[ -n "$CUSTOM_EXCLUDE" ]]; then
    export WS_EXCLUDES="${WS_EXCLUDES} ${CUSTOM_EXCLUDE}"
  fi

  # WS_PRODUCTNAME is treat as a input
  # it's set outside the script
  # shellcheck disable=SC2153
  echo "Product name - $WS_PRODUCTNAME"
  echo "Project name - $WS_PROJECTNAME"

  if [ "${DRYRUN}" = false ]; then
    echo "⏳ Scanning $FOLDER"
    java -jar /wss/wss-unified-agent.jar -c "$CONFIG_PATH"
    scan_result="$?"
  else
    echo "✅ DRYRUN Successful for $FOLDER"
  fi
  popd
  if [[ "$scan_result" != 0 ]]; then
    return 1
  else
    return 0
  fi
}


if [[ "$CREATE_SUBPROJECTS" == "true" ]]; then
  # find all go.mod / Gopkg.toml / package.json projects and scan them individually
  if [[ -n "$CUSTOM_EXCLUDE" ]]; then
    found_components=$(find . -name "$COMPONENT_DEFINITION" -not -path "./tests/*" -not -path "./docs/*" -not -path "${CUSTOM_EXCLUDE}")
  else
    found_components=$(find . -name "$COMPONENT_DEFINITION" -not -path "./tests/*" -not -path "./docs/*" )
  fi

  while read -r component_definition_path; do
    component_path="${component_definition_path%/*}"
    # keep only the last diretrory in the tree as a name
    component="${component_path##*/}"

    set +e
    scanFolder "${component_path}" "${PROJECT}-${component}"
    scan_result="$?"
    set -e

    if [[ "$scan_result" -ne 0 ]]; then
      echo "🛑 Scan for ${FOLDER} has failed"
      scan_failed=1
    fi
  done <<< "$found_components"
else
  set +e
  scanFolder "." "${PROJECT}"
  scan_result="$?"
  set -e

  if [[ "$scan_result" -ne 0 ]]; then
    echo "🛑 Scan for $(pwd) has failed"
    scan_failed=1
  fi
fi

if [[ "$scan_failed" -eq 1 ]]; then
  echo "🛑 One or more of the scans have failed"
  exit 1
else
  echo "💩 Scanning Finished"
fi
