#!/bin/bash

set -o pipefail

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

if [ -f "../../prow/scripts/library.sh" ]; then
    TEST_INFRA_SOURCES_DIR="../.."

elif [ -f "../test-infra/prow/scripts/library.sh" ]; then
    TEST_INFRA_SOURCES_DIR="../test-infra"

else
	echo "File '/prow/scripts/library.sh' does not exists."
    exit 1;
fi

source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

# Set the kubeconfig

if [[ "${KUBERNETES_SERVICE_HOST}" == "" ]]; then
    shout "KUBERNETES_SERVICE_HOST not set !!!"
    exit 1
fi

if [[ "${SLACK_TOKEN}" == "" || "${SLACK_URL}" == "" || "${SLACK_CHANNEL}" == ""]]; then
  shout "Slack details not set!! Exiting"
  exit 1
fi

if [[ "${GRAFANA_URL}" == "" ]]; then
  shout "Grafana URL not set!! Exiting"
  exit 1
fi

export SRC_DIR="$(mktemp -d -t src.XXXXXX)"


# Create Kyma Cluster
cluster.sh --action create --cluster-grade production
if [[ $? != 0 ]]; then
shoutFail "Cluster creation failed!!"
curl -X POST \
 -H 'Content-type: application/json; charset=utf-8' \
 --data '{"channel":"'"${SLACK_CHANNEL}"'","text":"Test Run: Failure \n Date: '"${DATE}"' \n Revision: '"${REVISION}"' \n Reason: Cluster Creation Failed"}' \
 $SLACK_URL/$SLACK_TOKEN
 exit 1
fi

export REVISION="$(cd /${SRC_DIR}/kyma-project/kyma && git rev-parse --short HEAD)"

# Get virtualservice
# Switch to kubeconfig from kyma cluster
gcloud container clusters get-credentials "${INPUT_CLUSTER_NAME}" --zone="${CLOUDSDK_COMPUTE_ZONE}" --project="${CLOUDSDK_CORE_PROJECT}"

# Get Virtual Service Host Name
export VIRTUAL_SERVICE_NAME="$(kubectl get virtualservice core-console -n kyma-system -o jsonpath='{.spec.hosts[0]}')"

export CLUSTER_DOMAIN_NAME="${VIRTUAL_SERVICE_NAME#*.}"

if [[ "${CLUSTER_DOMAIN_NAME}" == "" ]]; then
    shoutFail "Environment CLUSTER_DOMAIN_NAME is empty"
    exit 1
fi

shout "Domain name of the kyma cluster is: ${CLUSTER_DOMAIN_NAME}"

# Add the xip.io self-signed certificate to Linux trusted certificates -- Linux (Ubuntu, Debian)
DIR_SHARED_CERT="/usr/local/share/ca-certificates"

if [ ! -d "${DIR_SHARED_CERT}" ];
then
    shoutFail "Directory ${DIR_SHARED_CERT} DOES NOT exists."
    exit 0
fi

CERT_DIR="$(mktemp -d -t cert.XXXXXX)"

TMP_FILE=$(mktemp $CERT_DIR/temp-cert.XXXXXX) \
   && kubectl get configmap  net-global-overrides -n kyma-installer -o jsonpath='{.data.global\.ingress\.tlsCrt}' | base64 -d > ${TMP_FILE} \
   && cp ${TMP_FILE} ${DIR_SHARED_CERT}/$(basename -- ${TMP_FILE}) \
   && update-ca-certificates \
   && update-ca-certificates --fresh \
   && rm ${TMP_FILE}

if [[ $TESTS_DIR == "" ]]; then
  shoutFail "TESTS Directory is not defined!!"
  exit 1
fi

shout "Applying all prequisite files !!"
for f in $(find "${TESTS_DIR}/prerequisites" -type f -name *.yaml); do
  shout "Applying following file: $f"
  kubectl apply -f $f
done


# Run K6 scripts
shout "Running K6 Scripts"

if [[ "${RUNMODE}" == "all" ]]; then
  shout "Running the complete test suite"
  bash ${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/k6-runner.sh all
  if [[ $? != 0 ]]; then
    shoutFail "K6 test scripts run failed!!"
    curl -X POST \
    -H 'Content-type: application/json; charset=utf-8' \
    --data '{"channel":"'"${SLACK_CHANNEL}"'","text":"Test Run: Failure \n Date: '"${DATE}"' \n Revision: '"${REVISION}"' \n Reason: K6 test scripts run failed"}' \
    $SLACK_URL/$SLACK_TOKEN
    exit 1
  fi
elif [[ "${RUNMODE}" == "" && "${SCRIPTPATH}" != "" ]]; then
  shout "Running following Script: $SCRIPTPATH"
  bash  ${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/k6-runner.sh $SCRIPTPATH
  if [[ $? != 0 ]]; then
    shoutFail "K6 test scripts run failed!!"
    curl -X POST \
    -H 'Content-type: application/json; charset=utf-8' \
    --data '{"channel":"'"${SLACK_CHANNEL}"'","text":"Test Run: Failure \n Date: '"${DATE}"' \n Revision: '"${REVISION}"' \n Reason: K6 test scripts run failed"}' \
    $SLACK_URL/$SLACK_TOKEN
    exit 1
  fi
fi

shout "Finished all k6 tests!!"

shout "Deleting the deployed kyma cluster!!"

source "cluster.sh" "--action" "delete"

DATE="$(date)"

curl -X POST \
 -H 'Content-type: application/json; charset=utf-8' \
 --data '{"channel":"'"${SLACK_CHANNEL}"'","text":"Test Run: Success \n Date: '"${DATE}"' \n Revision: '"${REVISION}"' \n Grafana:'"${GRAFANA_URL}"'"}' \
 $SLACK_URL/$SLACK_TOKEN