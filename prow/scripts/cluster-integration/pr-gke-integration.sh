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

#!Put cleanup code in this function!
cleanup() {
    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    if [ -n "${CLEANUP_CLUSTER}" ]; then
      echo "################################################################################"
      echo "# Deprovisioning cluster: \"${CLUSTER_NAME}\""
      echo "################################################################################"
      "${KYMA_SOURCES_DIR}/prow/scripts/deprovision-gke-cluster.sh"
    fi

    if [ -n "${CLEANUP_DNS_ENTRY}" ]; then
      echo "################################################################################"
      echo "# Remove DNS Entry"
      echo "################################################################################"
      #TODO: script!
      ${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/remove-dns-entry.sh
    fi

    if [ -n "${CLEANUP_IP_ADDRESS}" ]; then
      echo "################################################################################"
      echo "# Remove IP Address"
      echo "################################################################################"
      #TODO: script!
      ${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/remove-ip-address.sh
    fi

    echo "################################################################################"
    echo "# Job is finished "
    echo "################################################################################"
    set -e
}

#TODO: Externalize this path somewhere (Job Config?)
SOURCES_DIR="/home/prow/go/src/github.com/kyma-project"
TEST_INFRA_SOURCES_DIR="${SOURCES_DIR}/test-infra"
KYMA_SOURCES_DIR="${SOURCES_DIR}/kyma"

echo "################################################################################"
echo "# Reserve IP Address"
echo "################################################################################"
#TODO: script!
${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/reserve-ip-address.sh
CLEANUP_IP_ADDRESS="true"


echo "################################################################################"
echo "# Add DNS Entry"
echo "################################################################################"
#TODO: script!
${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/setup-dns-entry.sh
CLEANUP_DNS_ENTRY="true"


echo "################################################################################"
echo "# Provisioning cluster: \"${CLUSTER_NAME}\""
echo "################################################################################"
export CLUSTER_NAME="${REPO_OWNER}-${REPO_NAME}-${PULL_NUMBER}"
${KYMA_SOURCES_DIR}/prow/scripts/provision-gke-cluster.sh
CLEANUP_CLUSTER="true"


echo "################################################################################"
echo "# MOCK: Installing Kyma, testing, etc..."
echo "################################################################################"

sleep 60

