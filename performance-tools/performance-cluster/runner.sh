#!/bin/bash

set -o pipefail

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

if [ -f "../../prow/scripts/library.sh" ]; then
    source "../../prow/scripts/library.sh"

elif [ -f "../test-infra/prow/scripts/library.sh" ]; then
    source "../test-infra/prow/scripts/library.sh"
    export LIBS_DIR="../test-infra/prow/scripts/library.sh"

else
    echo "File 'library.sh' can't be found."
    exit 1;
fi

# Set the kubeconfig

if [[ "${KUBERNETES_SERVICE_HOST}" == "" ]]; then
    shout "KUBERNETES_SERVICE_HOST not set !!!"
    exit 1
fi

if [[ "${SLACK_TOKEN}" == "" || "${SLACK_URL}" == "" || "${SLACK_CHANNEL}" == "" ]]; then
    shout "Slack details not set!! Exiting"
    exit 1
fi

if [[ "${GRAFANA_URL}" == "" ]]; then
    shout "Grafana URL not set!! Exiting"
    exit 1
fi

SLACK_URL=`echo $SLACK_URL`

export SRC_DIR="$(mktemp -d -t src.XXXXXX)"

# Create Kyma Cluster
${SCRIPTS_DIR}/cluster.sh --action create --cluster-grade production --infra "${INFRA}"
if [[ $? != 0 ]]; then
    shoutFail "Cluster creation failed!!"
    DATE="$(date)"
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

# Set kubernetes context
kubectl config set-context "gke_${CLOUDSDK_CORE_PROJECT}_${CLOUDSDK_COMPUTE_ZONE}_${INPUT_CLUSTER_NAME}" --namespace=default

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

if [ ! -d "${DIR_SHARED_CERT}" ]; then
    shoutFail "Directory ${DIR_SHARED_CERT} DOES NOT exists."
    exit 0
fi

CERT_DIR="$(mktemp -d -t cert.XXXXXX)"

TMP_FILE=$(mktemp $CERT_DIR/temp-cert.XXXXXX) \
   && kubectl get configmap  net-global-overrides -n kyma-installer -o jsonpath='{.data.global\.ingress\.tlsCrt}' | base64 -d > ${TMP_FILE} \
   && cp ${TMP_FILE} ${DIR_SHARED_CERT}/$(basename -- ${TMP_FILE}).crt \
   && update-ca-certificates \
   && update-ca-certificates --fresh \
   && rm ${TMP_FILE}


if [[ $TESTS_DIR == "" ]]; then
  shoutFail "TESTS Directory is not defined!!"
  exit 1
fi

export PREREQ_PATH="${SRC_DIR}/kyma-project/kyma/${TESTS_DIR}/prerequisites"
export TESTS_PATH="${SRC_DIR}/kyma-project/kyma/${TESTS_DIR}/components"

shout "Applying all prequisite files !!"
for f in $(find "${PREREQ_PATH}" -type f -name '*setup.sh'); do
  shout "Running following file: $f"
  source $f
done

shout "Running K6 Scripts"

if [[ "${RUNMODE}" == "all" ]]; then
    shout "Running the complete test suite"
    bash ${SCRIPTS_DIR}/scripts/k6-runner.sh all
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
    bash ${SCRIPTS_DIR}/scripts/k6-runner.sh $SCRIPTPATH
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

service docker stop

source "${SCRIPTS_DIR}/cluster.sh" "--action" "delete" "--infra" "${INFRA}"

DATE="$(date)"

shout "DATE: ${DATE}"

curl -X POST \
    -H 'Content-type: application/json; charset=utf-8' \
    --data '{"channel":"'"${SLACK_CHANNEL}"'","text":"Test Run: Success \n Date: '"${DATE}"' \n Revision: '"${REVISION}"' \n Grafana:'"${GRAFANA_URL}"'"}' \
    $SLACK_URL/$SLACK_TOKEN
