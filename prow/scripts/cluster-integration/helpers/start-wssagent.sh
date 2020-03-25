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

# whitesource config
GO_CONFIG_PATH="/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/whitesource-scanner/go-wss-unified-agent.config"
JAVASCRIPT_CONFIG_PATH="/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/whitesource-scanner/javascript-wss-unified-agent.config"

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
        CONFIG_PATH=$GO_CONFIG_PATH
        sed -i.bak "s|go.dependencyManager=|go.dependencyManager=dep|g" $CONFIG_PATH
        ;;

    golang-mod)
        echo "SCAN: golang-mod"
        CONFIG_PATH=$GO_CONFIG_PATH
        sed -i.bak "s|go.dependencyManager=|go.dependencyManager=modules|g" $CONFIG_PATH
        ;;
        
    javascript)
        echo "SCAN: javascript"
        CONFIG_PATH=$JAVASCRIPT_CONFIG_PATH
        ;;
        
    *)
        echo "can only be golang or javascript"
        exit 1
esac

echo "***********************************"
echo "***********Starting Scan***********"
echo "***********************************"

KYMA_SRC="${GITHUB_ORG_DIR}/${PROJECTNAME}"

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
    cd "${FOLDER}" # change to passed parameter
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


    if [ "${DRYRUN}" = false ];then
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
}

scanFolder "${KYMA_SRC}" "${PROJECTNAME}"

echo "***********************************"
echo "*********Scanning Finished*********"
echo "***********************************"