#!/usr/bin/env bash

set -o errexit

############################################################
# REPO_OWNER, REPO_NAME and PULL_NUMBER are set by ProwJob #
############################################################

discoverUnsetVar=false

for var in REPO_OWNER REPO_NAME PULL_NUMBER CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_DNS_ZONE_NAME; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

#For reserve-ip-address.sh
export GCLOUD_REGION="${CLOUDSDK_COMPUTE_REGION}"

#For provision-gke-cluster.sh
export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"

export DNS_ZONE_NAME="${CLOUDSDK_DNS_ZONE_NAME}"

#For provision-gke-cluster.sh
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"

export GCLOUD_IP_ADDRESS_NAME=$(echo "pr-${PULL_NUMBER}-job-${PROW_JOB_ID}" | tr "[:upper:]" "[:lower:]")
export DNS_SUBDOMAIN="${GCLOUD_IP_ADDRESS_NAME}"
export CLUSTER_NAME="${REPO_OWNER}-${REPO_NAME}-${PULL_NUMBER}"
export IP_ADDRESS="will_be_generated"

trap cleanup EXIT

#!Put cleanup code in this function!
cleanup() {
    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    if [ -n "${CLEANUP_CLUSTER}" ]; then
      echo "################################################################################"
      echo "# Deprovision cluster: \"${CLUSTER_NAME}\""
      echo "################################################################################"
      "${KYMA_SOURCES_DIR}/prow/scripts/deprovision-gke-cluster.sh"
    fi

    if [ -n "${CLEANUP_DNS_RECORD}" ]; then
      echo "################################################################################"
      echo "# Delete DNS Record"
      echo "################################################################################"
      ${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/delete-dns-record.sh
    fi

    if [ -n "${CLEANUP_IP_ADDRESS}" ]; then
      echo "################################################################################"
      echo "# Release IP Address"
      echo "################################################################################"
      ${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/release-ip-address.sh
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
export IP_ADDRESS=$(${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/reserve-ip-address.sh)
CLEANUP_IP_ADDRESS="true"

echo "################################################################################"
echo "# Create DNS Record"
echo "################################################################################"
${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/create-dns-record.sh
CLEANUP_DNS_RECORD="true"


#echo "################################################################################"
#echo "# Generate certificate for the domain: \"${DOMAIN_NAME}\""
#echo "################################################################################"
#${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/generate-self-signed-cert.sh


echo "################################################################################"
echo "# Provision cluster: \"${CLUSTER_NAME}\""
echo "################################################################################"
${KYMA_SOURCES_DIR}/prow/scripts/provision-gke-cluster.sh
CLEANUP_CLUSTER="true"


echo "################################################################################"
echo "# MOCK: Installing Kyma, testing, etc..."
echo "################################################################################"

sleep 60

