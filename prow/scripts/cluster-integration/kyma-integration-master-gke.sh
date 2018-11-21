#!/usr/bin/env bash

set -o errexit

############################################################
# REPO_OWNER, REPO_NAME and PULL_NUMBER are set by ProwJob #
############################################################
# MACHINE_TYPE (optional): GKE machine type                #
############################################################

discoverUnsetVar=false

for var in REPO_OWNER REPO_NAME PULL_NUMBER KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

readonly TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
readonly KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
readonly KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
readonly KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

readonly INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
readonly INSTALLER_CONFIG="${KYMA_RESOURCES_DIR}/installer-config-cluster.yaml.tpl"
readonly INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"

IP_ADDRESS_NAME=$(echo "pr-${PULL_NUMBER}-job-${PROW_JOB_ID}" | tr "[:upper:]" "[:lower:]")
export IP_ADDRESS_NAME
export DNS_SUBDOMAIN="${IP_ADDRESS_NAME}"
export CLUSTER_NAME="${REPO_OWNER}-${REPO_NAME}-${PULL_NUMBER}"
export IP_ADDRESS="will_be_generated"

#For provision-gke-cluster.sh
export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
#For provision-gke-cluster.sh
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"

source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

shout "Authenticate"
init

shout "Reserve IP Address"
IP_ADDRESS=$("${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/reserve-ip-address.sh)
export IP_ADDRESS
CLEANUP_IP_ADDRESS="true"
shout "IP Address: ${IP_ADDRESS} created"


shout "Create DNS Record"
date
DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"
export DNS_DOMAIN
"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/create-dns-record.sh
CLEANUP_DNS_RECORD="true"

shout "Provision cluster: \"${CLUSTER_NAME}\""
date
export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
if [ -z "$MACHINE_TYPE" ]; then
      export MACHINE_TYPE="n1-standard-2"
fi

"${KYMA_SOURCES_DIR}"/prow/scripts/provision-gke-cluster.sh
CLEANUP_CLUSTER="true"

shout "Install Tiller"
date
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
"${KYMA_SCRIPTS_DIR}"/install-tiller.sh

shout "Generate self-signed certificate"
date
export DOMAIN=${DNS_DOMAIN%?}
CERT_KEY=$("${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/generate-self-signed-cert.sh)
TLS_CERT=$(echo "${CERT_KEY}" | head -1)
TLS_KEY=$(echo "${CERT_KEY}" | tail -1)

shout "Apply Kyma config"
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

shout "Trigger installation"
date
kubectl label installation/kyma-installation action=install
"${KYMA_SCRIPTS_DIR}"/is-installed.sh