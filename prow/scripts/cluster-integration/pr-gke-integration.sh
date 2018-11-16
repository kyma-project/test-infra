#!/usr/bin/env bash

#Description: Kyma Integration plan on GKE. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma on real GKE cluster.
#
#
#Expected vars:
#
# - REPO_OWNER - Set up by prow, repository owner/organization
# - REPO_NAME - Set up by prow, repository name
# - PULL_NUMBER - Set up by prow, Pull request number
# - DOCKER_PUSH_REPOSITORY - Docker repository hostname
# - DOCKER_PUSH_DIRECTORY - Docker "top-level" directory (with leading "/")
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - CLOUDSDK_CORE_PROJECT - GCP project for all GCP resources used during execution (Service Account, IP Address, DNS Zone, image registry etc.)
# - CLOUDSDK_COMPUTE_REGION - GCP compute region
# - CLOUDSDK_DNS_ZONE_NAME - GCP zone name (not its DNS name!)
# - GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path
#
#Permissions: In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
# - Compute Network Admin
# - Kubernetes Engine Admin
# - Kubernetes Engine Cluster Admin
# - DNS Administrator
# - Service Account User
# - Storage Admin

set -o errexit

############################################################
# REPO_OWNER, REPO_NAME and PULL_NUMBER are set by ProwJob #
############################################################
# MACHINE_TYPE (optional): GKE machine type                #
############################################################

discoverUnsetVar=false

for var in REPO_OWNER REPO_NAME PULL_NUMBER DOCKER_PUSH_REPOSITORY DOCKER_PUSH_DIRECTORY KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS; do
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
    #Try to preserve exit status unless a new error occurs
    EXIT_STATUS=$?

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    if [ -n "${CLEANUP_CLUSTER}" ]; then
        echo "################################################################################"
        echo "# Deprovision cluster: \"${CLUSTER_NAME}\""
        echo "################################################################################"
        date
        "${KYMA_SOURCES_DIR}"/prow/scripts/deprovision-gke-cluster.sh
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    fi

    if [ -n "${CLEANUP_DNS_RECORD}" ]; then
        echo "################################################################################"
        echo "# Delete DNS Record"
        echo "################################################################################"
        date
        "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/delete-dns-record.sh
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    fi

    if [ -n "${CLEANUP_IP_ADDRESS}" ]; then
        echo "################################################################################"
        echo "# Release IP Address"
        echo "################################################################################"
        date
        "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/release-ip-address.sh
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    fi

    if [ -n "${CLEANUP_DOCKER_IMAGE}" ]; then
        echo "################################################################################"
        echo "# Delete temporary Kyma-Installer Docker image"
        echo "################################################################################"
        date
        "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/delete-image.sh
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    fi


    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    echo "################################################################################"
    echo "# Job is finished ${MSG}"
    echo "################################################################################"
    date
    set -e

    exit "${EXIT_STATUS}"
}

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
IP_ADDRESS_NAME=$(echo "pr-${PULL_NUMBER}-job-${PROW_JOB_ID}" | tr "[:upper:]" "[:lower:]")
export IP_ADDRESS_NAME
export DNS_SUBDOMAIN="${IP_ADDRESS_NAME}"
export CLUSTER_NAME="${REPO_OWNER}-${REPO_NAME}-${PULL_NUMBER}"
export IP_ADDRESS="will_be_generated"

#For provision-gke-cluster.sh
export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"

#Local variables
KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
INSTALLER_CONFIG="${KYMA_RESOURCES_DIR}/installer-config-cluster.yaml.tpl"
INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"

echo "################################################################################"
echo "# Authenticate"
echo "################################################################################"
date
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
init


echo "################################################################################"
echo "# Build Kyma-Installer Docker image"
echo "################################################################################"
date
export KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-integration/${REPO_OWNER}/${REPO_NAME}:PR-${PULL_NUMBER}"
"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/create-image.sh
CLEANUP_DOCKER_IMAGE="true"


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
DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"
export DNS_DOMAIN
"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/create-dns-record.sh
CLEANUP_DNS_RECORD="true"


echo "################################################################################"
echo "# Provision cluster: \"${CLUSTER_NAME}\""
echo "################################################################################"
date
export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
if [ -z "$MACHINE_TYPE" ]; then
      export MACHINE_TYPE="n1-standard-2"
fi
"${KYMA_SOURCES_DIR}"/prow/scripts/provision-gke-cluster.sh
CLEANUP_CLUSTER="true"


echo "################################################################################"
echo "Install Tiller"
echo "################################################################################"
date
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
"${KYMA_SCRIPTS_DIR}"/install-tiller.sh


echo "################################################################################"
echo "Generate self-signed certificate"
echo "################################################################################"
date
export DOMAIN=${DNS_DOMAIN%?}
CERT_KEY=$("${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/generate-self-signed-cert.sh)
TLS_CERT=$(echo "${CERT_KEY}" | head -1)
TLS_KEY=$(echo "${CERT_KEY}" | tail -1)

echo "################################################################################"
echo "Apply Kyma config"
echo "################################################################################"
date
"${KYMA_SCRIPTS_DIR}"/concat-yamls.sh "${INSTALLER_YAML}" "${INSTALLER_CONFIG}" "${INSTALLER_CR}" \
    | sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" \
    | sed -e "s/__DOMAIN__/${DOMAIN}/g" \
    | sed -e "s/__TLS_CERT__/${TLS_CERT}/g" \
    | sed -e "s/__TLS_KEY__/${TLS_KEY}/g" \
    | sed -e "s/__EXTERNAL_PUBLIC_IP__/${IP_ADDRESS}/g" \
    | sed -e "s/__SKIP_SSL_VERIFY__/true/g" \
    | sed -e "s/__VERSION__/0.0.1/g" \
    | sed -e "s/__.*__//g" \
    | kubectl apply -f-


echo "################################################################################"
echo "Trigger installation"
echo "################################################################################"
date
kubectl label installation/kyma-installation action=install
"${KYMA_SCRIPTS_DIR}"/is-installed.sh
