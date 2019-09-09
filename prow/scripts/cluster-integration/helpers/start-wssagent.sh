#!/usr/bin/env bash


#Description: runs wss-unified-agent 
#The purpose is to run the wss-unified-agent

#Expected vars: 
# - APIKEY- Key provided by SAP Whitesource Team
# - PRODUCTNAME - Product inside whitesource
# - USERKEY - Users specified key(should be a service account)
# - PROJECTNAME- Kyma component name, scans that directory and posts the results in whitesource

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

echo "***********************************"
echo "***********Starting Scan***********"
echo "***********************************"

sed -i.bak "s/apiKey=/apiKey=${APIKEY}/g
s/productName=/productName=${PRODUCTNAME}/g
s/userKey=/userKey=${USERKEY}/g
s/projectName=/projectName=${PROJECTNAME}/g" /wss/wss-unified-agent.config

echo "Dryrun state - $DRYRUN"
echo "Product name - $PRODUCTNAME"
echo "Project name - $PROJECTNAME"
echo "Java Options - '$JAVA_OPTS'"

cd "${KYMA_PROJECT_DIR}/${PROJECTNAME}" # run from project directory

 if [ "${DRYRUN}" = false ];then
    echo "***********************************"
    echo "***********Scanning ***************"
    echo "***********************************"
    if [ -z "$JAVA_OPTS" ]; then
        echo "no additional java_opts set"
        java -jar /wss/wss-unified-agent.jar -c /wss/wss-unified-agent.config
    else
        java "${JAVA_OPTS}" -jar /wss/wss-unified-agent.jar -c /wss/wss-unified-agent.config
    fi
else 
    echo "***********************************"
    echo "********* DRYRUN Success **********"  
    echo "***********************************"
fi
echo "***********************************"
echo "*********Scanning Finished*********"
echo "***********************************"  