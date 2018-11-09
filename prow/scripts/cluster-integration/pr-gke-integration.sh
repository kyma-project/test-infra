#!/usr/bin/env bash

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

#Setup variables
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

echo "################################################################################"
echo "# Authenticate"
echo "################################################################################"
date
# shellcheck source=/dev/null
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
init


echo "################################################################################"
echo "# Build Kyma-Installer Docker image"
echo "################################################################################"
date
#DOCKER_TAG is created by library.sh script
export KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-integration/${REPO_OWNER}/${REPO_NAME}:${DOCKER_TAG}"
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
    | sed -E ";s;develop\/installer:.+;rc/kyma-installer:0.5-rc;" \
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
