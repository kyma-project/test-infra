#!/bin/bash

set -o errexit
set -o pipefail

source "${CURRENT_PATH}/scripts/library.sh"


# Set the kubeconfig

if [[ "${KUBERNETES_SERVICE_HOST}" == "" ]]; then
    shout "KUBERNETES_SERVICE_HOST not set !!!"
    exit 1
fi


gcloud container clusters get-credentials $LOADGEN_CLUSTER_NAME --zone $LOADGEN_COMPUTE_ZONE --project $CLOUDSDK_CORE_PROJECT

# Create Kyma Cluster
source "${CURRENT_PATH}/scripts/cluster.sh" "--action" "create" "--cluster-grade" "production"


# Get virtualservice
# Switch to kubeconfig from kyma cluster
gcloud container clusters get-credentials "${INPUT_CLUSTER_NAME}" --zone="${CLOUDSDK_COMPUTE_ZONE}" --project="${CLOUDSDK_CORE_PROJECT}"

# Get Virtual Service Host Name
export VIRTUAL_SERVICE_NAME="$(kubectl get virtualservice core-console -n kyma-system -o jsonpath='{.spec.hosts[0]}')"

export VIRTUAL_SERVICE_HOST_NAME="${VIRTUAL_SERVICE_NAME#*.}"

if [[ "${VIRTUAL_SERVICE_HOST_NAME}" == "" ]]; then
    shoutFail "Environment VIRTUAL_SERVICE_HOST_NAME is empty"
    exit 0
fi

shout "Virtual Service Host Name ${VIRTUAL_SERVICE_HOST_NAME}"

# Add the xip.io self-signed certificate to Linux trusted certificates -- Linux (Ubuntu, Debian)
DIR_SHARED_CERT="/usr/local/share/ca-certificates"

if [ ! -d "${DIR_SHARED_CERT}" ]
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

# Get test scripts


# Run K6 scripts
shout "Running K6 Scripts"

if [[ "${RUNMODE}" == "all" ]]; then
  shout "Running the complete test suite"
  source "k6-runner.sh" "all"
  set -o errexit
elif [[ "${RUNMODE}" == "" && "${SCRIPTPATH}" != "" ]]; then
  shout "Running following Script: $SCRIPTPATH"
  source "k6-runner.sh" "${SCRIPTPATH}"
  set -o errexit
fi

shout "Finished all k6 tests!!"

shout "Deleting the deployed kyma cluster!!"

source "${CURRENT_PATH}/scripts/cluster.sh" "--action" "delete" "--cluster-grade" "production"