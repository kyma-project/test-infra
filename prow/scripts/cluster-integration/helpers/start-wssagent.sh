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


if [ -f "../../prow/scripts/library.sh" ]; then
    export TEST_INFRA_SOURCES_DIR="../.."

elif [ -f "../test-infra/prow/scripts/library.sh" ]; then
    export TEST_INFRA_SOURCES_DIR="../test-infra"

else
	echo "File 'library.sh' can't be found."
    exit 1;
fi

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

# authenticate gcloud client
init

export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

gsutil cp "gs://kyma-prow-secrets/whitesource-userkey.encrypted" "." 
gsutil cp "gs://kyma-prow-secrets/whitesource-apikey.encrypted" "." 

"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/decrypt.sh" "whitesource-userkey" "whitesource-userkey.encrypted"
USERKEY=$(cat "whitesource-userkey")

"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/decrypt.sh" "whitesource-apikey" "whitesource-apikey.encrypted"
APIKEY=$(cat "whitesource-apikey")

case "${SCAN_LANGUAGE}" in
    golang)
        echo "SCAN: golang (dep)"
        sed -i.bak "s|go.dependencyManager=godep|go.dependencyManager=dep|g" /wss/wss-unified-agent.config
        ;;

    golang-mod)
        echo "SCAN: golang-mod"
        sed -i.bak "s|go.dependencyManager=godep|go.dependencyManager=modules|g" /wss/wss-unified-agent.config
        ;;
        
    javascript)
        echo "SCAN: JAVASTRIPT, disable scanning for go dependencies"
        sed -i.bak "s|go.resolveDependencies=true|# go.resolveDependencies=true|g" /wss/wss-unified-agent.config
        sed -i.bak "s|go.collectDependenciesAtRuntime=false|# go.collectDependenciesAtRuntime=false|g" /wss/wss-unified-agent.config
        sed -i.bak "s|go.dependencyManager=godep|# go.dependencyManager=godep|g" /wss/wss-unified-agent.config
        ;;
        
    *)
        echo "can only be golang or javascript"
        exit 1
esac

# resolve deps for console repository
if [ "${PROJECTNAME}" == "console" ]; then
    cd "$KYMA_SRC"
    make resolve
fi    

echo "***********************************"
echo "***********Starting Scan***********"
echo "***********************************"

KYMA_SRC="${GITHUB_ORG_DIR}/${PROJECTNAME}"
   

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
    cd "${FOLDER}" # change to passed parameter
    PROJNAME=$2

    # adjust global setting ignoreSourceFiles=true
    sed -i.bak "s|#ignoreSourceFiles=true|ignoreSourceFiles=false|g;" /wss/wss-unified-agent.config

    if [[ $CUSTOM_PROJECTNAME == "" ]]; then 
    # use custom projectname for kyma-mod scans in order not to override kyma (dep) scan results
        sed -i.bak "s|apiKey=|apiKey=${APIKEY}|g; s|productName=|productName=${PRODUCTNAME}|g; s|userKey=|userKey=${USERKEY}|g; s|projectName=|projectName=${PROJNAME}|g" /wss/wss-unified-agent.config
    else
        sed -i.bak "s|apiKey=|apiKey=${APIKEY}|g; s|productName=|productName=${PRODUCTNAME}|g; s|userKey=|userKey=${USERKEY}|g; s|projectName=|projectName=${CUSTOM_PROJECTNAME}|g" /wss/wss-unified-agent.config
    fi

    echo "Product name - $PRODUCTNAME"
    echo "Project name - $PROJNAME"
    echo "Java Options - '$JAVA_OPTS'"


    if [ "${DRYRUN}" = false ];then
        echo "***********************************"
        echo "******** Scanning $FOLDER ***"
        echo "***********************************"
        if [ -z "$JAVA_OPTS" ]; then
            echo "no additional java_opts set"
            java -jar /wss/wss-unified-agent.jar -c /wss/wss-unified-agent.config
        else
            java "${JAVA_OPTS}" -jar /wss/wss-unified-agent.jar -c /wss/wss-unified-agent.config
        fi
    else 
        echo "***********************************"
        echo "******** DRYRUN Successful for $FOLDER ***"  
        echo "***********************************"
    fi
}

scanFolder "${KYMA_SRC}" "${PROJECTNAME}"

echo "***********************************"
echo "*********Scanning Finished*********"
echo "***********************************"