#!/usr/bin/env bash


#Description: runs wss-unified-agent 
#The purpose is to run the wss-unified-agent and modify its config the 
#jar can be found 
#curl -LJO https://github.com/whitesource/unified-agent-distribution/raw/master/standAlone/wss-unified-agent.jar
#config can be found here
#curl -LJO https://github.com/whitesource/unified-agent-distribution/raw/master/standAlone/wss-unified-agent.config

#Expected vars: 
# - APIKEY- Key provided by SAP Whitesource Team
# - PRODUCTNAME - Product inside whitesource
# - USERKEY - Users specified key(should be a service account)
# - PROJECTNAME- Kyma component name, scans that directory and posts the results in whitesource

set -o errexit

gsutil cp "gs://kyma-prow-secrets/whitesource-userkey.encrypted" "." 
gsutil cp "gs://kyma-prow-secrets/whitesource-apikey.encrypted" "." 

"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/decrypt.sh" "whitesource-userkey.encrypted" "whitesource-userkey" 
USERKEY=$(cat "whitesource-userkey")

"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/decrypt.sh" "whitesource-apikey.encrypted" "whitesource-apikey"
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

 if [ "${DRYRUN}" -eq 0 ];then
    echo "***********************************"
    echo "***********Scanning ***************"
    echo "***********************************"
    java "${JAVA_OPTS}" -jar /wss/wss-unified-agent.jar -c /wss/wss-unified-agent.config
else 
    echo "***********************************"
    echo "********* DRYRUN Success **********"  
    echo "***********************************"
fi
echo "***********************************"
echo "*********Scanning Finished*********"
echo "***********************************"  