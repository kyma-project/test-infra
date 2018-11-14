#!/usr/bin/env bash

set -o errexit

############################################################
# REPO_OWNER, REPO_NAME and PULL_NUMBER are set by ProwJob #
############################################################

discoverUnsetVar=false

for var in REPO_OWNER REPO_NAME PULL_NUMBER SOURCES_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS; do
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
      echo "# Deprovision cluster: \"${CLUSTER_NAME}\""
      echo "################################################################################"
      date
      "${KYMA_SOURCES_DIR}"/prow/scripts/deprovision-gke-cluster.sh
    fi

    if [ -n "${CLEANUP_DNS_RECORD}" ]; then
      echo "################################################################################"
      echo "# Delete DNS Record"
      echo "################################################################################"
      date
      "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/delete-dns-record.sh
    fi

    if [ -n "${CLEANUP_IP_ADDRESS}" ]; then
      echo "################################################################################"
      echo "# Release IP Address"
      echo "################################################################################"
      date
      "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/release-ip-address.sh
    fi

    echo "################################################################################"
    echo "# Job is finished "
    echo "################################################################################"
    date
    set -e
}

TEST_INFRA_SOURCES_DIR="${SOURCES_DIR}/test-infra"
KYMA_SOURCES_DIR="${SOURCES_DIR}/kyma"

#Setup variables
IP_ADDRESS_NAME=$(echo "pr-${PULL_NUMBER}-job-${PROW_JOB_ID}" | tr "[:upper:]" "[:lower:]")
export IP_ADDRESS_NAME
export DNS_SUBDOMAIN="${IP_ADDRESS_NAME}"
export CLUSTER_NAME="${REPO_OWNER}-${REPO_NAME}-${PULL_NUMBER}"
export IP_ADDRESS="will_be_generated"

#For provision-gke-cluster.sh
export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
#For provision-gke-cluster.sh
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"


echo "################################################################################"
echo "# Authenticate"
echo "################################################################################"
date
export BUILD_TYPE="pr"
# shellcheck source=/dev/null
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
init


echo "################################################################################"
echo "# Reserve IP Address"
echo "################################################################################"
date
IP_ADDRESS=$("${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/reserve-ip-address.sh)
export IP_ADDRESS
CLEANUP_IP_ADDRESS="true"
echo "IP Address: ${IP_ADDRESS} created"


echo "################################################################################"
echo "# Create DNS Record"
echo "################################################################################"
date
"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/create-dns-record.sh
CLEANUP_DNS_RECORD="true"


echo "################################################################################"
echo "# Provision cluster: \"${CLUSTER_NAME}\""
echo "################################################################################"
date
export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
"${KYMA_SOURCES_DIR}"/prow/scripts/provision-gke-cluster.sh
CLEANUP_CLUSTER="true"


echo "################################################################################"
echo "# MOCK: Installing Kyma, testing, etc..."
echo "################################################################################"
date
echo "I'm pretending I'm doing something for the next 60 seconds..."
sleep 60
kubectl cluster-info
kubectl get pods --all-namespaces
