#!/usr/bin/env bash

#Description: runs wss-unified-agent
#The purpose is to run the wss-unified-agent

#Expected vars:
# - APIKEY- Key provided by SAP Whitesource Team
# - PRODUCTNAME - Product inside whitesource
# - USERKEY - Users specified key(should be a service account)
# - PROJECTNAME - Kyma component name, scans that directory and posts the results in whitesource
# - GITHUB_ORG_DIR - Project directory to scan
# - SCAN_LANGUAGE - Scan language is used to set the correct values in the whitesource config for golang / javascript

set -o errexit
export TEST_INFRA_SOURCES_DIR="/home/prow/go/src/github.com/kyma-project/test-infra/"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"

# whitesource config
GO_CONFIG_PATH="$TEST_INFRA_SOURCES_DIR/prow/images/whitesource-scanner/go-wss-unified-agent.config"
JAVASCRIPT_CONFIG_PATH="$TEST_INFRA_SOURCES_DIR/prow/images/whitesource-scanner/javascript-wss-unified-agent.config"

# authenticate gcloud client
gcp::authenticate \
  -c "${GOOGLE_APPLICATION_CREDENTIALS}"

export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

USERKEY=$(cat "${WHITESOURCE_USERKEY}")

APIKEY=$(cat "${WHITESOURCE_APIKEY}")

#exclude components based on dependency management
function filterFolders() {
  local DEPENDENCY_FILE_TO_EXCLUDE
  DEPENDENCY_FILE_TO_EXCLUDE=$1
  local FOLDER_TO_SCAN
  FOLDER_TO_SCAN=$2
  local EXCLUDES
  EXCLUDES=$({ cd "${FOLDER_TO_SCAN}" && find . -iname "${DEPENDENCY_FILE_TO_EXCLUDE}"; } | grep -v vendor | grep -v tests | xargs -n 1 dirname | sed 's/$/\/**/' | sed 's/^.\//**\//' | paste -s -d" " -)
  EXCLUDES="excludes=**/tests/** ${EXCLUDES}"
  echo "$EXCLUDES"
}

function prepareDependencies() {
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

# shellcheck disable=SC2153
KYMA_SRC="${GITHUB_ORG_DIR}/${PROJECTNAME}"

case "${SCAN_LANGUAGE}" in
golang)
  echo "SCAN: golang (dep)"
  go version
  CONFIG_PATH=$GO_CONFIG_PATH
  sed -i.bak "s|go.dependencyManager=|go.dependencyManager=dep|g" $CONFIG_PATH
  sed -i.bak '/^excludes=/d' $CONFIG_PATH
  echo "scanComment=$(date)" >> $CONFIG_PATH
  # exclude gomod based folders
  filterFolders go.mod "${KYMA_SRC}" >>${CONFIG_PATH}
  prepareDependencies gopkg.toml "${KYMA_SRC}"
  COMPONENT_DEFINITION="Gopkg.toml"
  ;;

golang-mod)
  echo "SCAN: golang-mod"
  go version
  CONFIG_PATH=$GO_CONFIG_PATH
  export GO111MODULE=on
  sed -i.bak "s|go.dependencyManager=|#go.dependencyManager=|g" $CONFIG_PATH
  sed -i.bak "s|go.collectDependenciesAtRuntime=true|#go.collectDependenciesAtRuntime=true|g" $CONFIG_PATH
  sed -i.bak "s|go.resolveDependencies=true|go.resolveDependencies=false|g" $CONFIG_PATH
  sed -i.bak "s|go.ignoreSourceFiles=true|go.modules.ignoreSourceFiles=true|g" $CONFIG_PATH
  sed -i.bak '/^excludes=/d' $CONFIG_PATH
  echo "go.modules.resolveDependencies=true" >> $CONFIG_PATH
  echo "scanComment=$(date)" >> $CONFIG_PATH
  # exclude godep based folders
  filterFolders gopkg.toml "${KYMA_SRC}" >>${CONFIG_PATH}
  COMPONENT_DEFINITION="go.mod"
  ;;

javascript)
  echo "SCAN: javascript"
  CONFIG_PATH=$JAVASCRIPT_CONFIG_PATH
  echo "scanComment=$(date)" >> $CONFIG_PATH
  COMPONENT_DEFINITION="package.json"
  ;;

*)
  echo "can only be golang or javascript"
  exit 1
  ;;
esac

echo "***********************************"
echo "***********Starting Scan***********"
echo "***********************************"

# resolve deps for console repository
#if [ "${PROJECTNAME}" == "console" ]; then
#    cd "$KYMA_SRC"
#    make resolve
#fi

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
  PROJNAME=$2

  if [[ $CUSTOM_PROJECTNAME == "" ]]; then
    # use custom projectname for kyma-mod scans in order not to override kyma (dep) scan results
    sed -i.bak "s|apiKey=|apiKey=${APIKEY}|g; s|productName=|productName=${PRODUCTNAME}|g; s|userKey=|userKey=${USERKEY}|g; s|projectName=|projectName=${PROJNAME}|g" $CONFIG_PATH
  else
    sed -i.bak "s|apiKey=|apiKey=${APIKEY}|g; s|productName=|productName=${PRODUCTNAME}|g; s|userKey=|userKey=${USERKEY}|g; s|projectName=|projectName=${CUSTOM_PROJECTNAME}|g" $CONFIG_PATH
  fi

  echo "Product name - $PRODUCTNAME"
  echo "Project name - $PROJNAME"
  echo "Java Options - '$JAVA_OPTS'"

  if [ "${DRYRUN}" = false ]; then
    echo "***********************************"
    echo "******** Scanning $FOLDER ***"
    echo "***********************************"
    if [ -z "$JAVA_OPTS" ]; then
      echo "no additional java_opts set"
      java -jar /wss/wss-unified-agent.jar -c $CONFIG_PATH
    else
      java "${JAVA_OPTS}" -jar /wss/wss-unified-agent.jar -c $CONFIG_PATH
    fi
  else
    echo "***********************************"
    echo "******** DRYRUN Successful for $FOLDER ***"
    echo "***********************************"
  fi
  popd
}

function scanSubprojects() {
  if [[ $1 == "" ]]; then
    echo "path cannot be empty"
    exit 1
  fi
  FOLDER=$1

  if [[ $2 == "" ]]; then
    echo "component definition cannot be empty"
    exit 1
  fi
  # TODO better name
  local component_definition=$2
  
  if [[ $3 == "" ]]; then
    echo "component name cannot be empty"
    exit 1
  fi
  pushd "${FOLDER}" # change to passed parameter
  PROJNAME=$3

  
  find . -name "$component_definition" -not -path "./tests/*" | while read component; do
    # TODO what about excludes?
    scanFolder "${folder_name}" "${component}"
  done
  popd
}

if [[ "$CREATE_SUBPROJECTS" == "true" ]]; then
  scanSubprojects "${KYMA_SRC}" "${COMPONENT_DEFINITION}" "${PROJECTNAME}"
else
  scanFolder "${KYMA_SRC}" "${PROJECTNAME}"
fi

echo "***********************************"
echo "*********Scanning Finished*********"
echo "***********************************"
