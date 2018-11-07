#!/usr/bin/env bash

set -o errexit

############################################################
# REPO_OWNER, REPO_NAME and PULL_NUMBER are set by ProwJob #
############################################################

discoverUnsetVar=false

for var in REPO_OWNER REPO_NAME PULL_NUMBER; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi


trap cleanup EXIT

#Put cleanup code in this function
cleanup() {
    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    echo "
    ################################################################################
    # Deprovisioning cluster: \"${CLUSTER_NAME}\"
    ################################################################################
    "
    ${KYMA_SOURCES_DIR}/prow/scripts/deprovision-gke-cluster.sh


    #TODO: Add more deprovisioning steps here


    echo "
    ################################################################################
    # Job is finished
    ################################################################################
    "
    set -e
}


SOURCES_DIR="/home/prow/go/src/github.com/kyma-project"
KYMA_SOURCES_DIR="${SOURCES_DIR}/kyma"

export CLUSTER_NAME="${REPO_OWNER}-${REPO_NAME}-${PULL_NUMBER}"

echo "
################################################################################
# Provisioning cluster: \"${CLUSTER_NAME}\"
################################################################################
"

${KYMA_SOURCES_DIR}/prow/scripts/provision-gke-cluster.sh

echo "
################################################################################
# MOCK: Installing Kyma, testing, etc...
################################################################################
"
sleep 60

