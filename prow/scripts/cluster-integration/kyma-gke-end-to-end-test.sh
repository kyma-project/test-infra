#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

discoverUnsetVar=false

for var in INPUT_CLUSTER_NAME REPO_OWNER_GIT REPO_NAME_GIT DOCKER_PUSH_REPOSITORY KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS KYMA_ARTIFACTS_BUCKET KYMA_BACKUP_RESTORE_BUCKET KYMA_BACKUP_CREDENTIALS CLOUDSDK_COMPUTE_ZONE; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"

### For provision-gke-cluster.sh
export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"
export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"

### For generate-cluster-backup-config.sh
export BACKUP_CREDENTIALS="${KYMA_BACKUP_CREDENTIALS}"
export BACKUP_RESTORE_BUCKET="${KYMA_BACKUP_RESTORE_BUCKET}"

readonly REPO_OWNER=$(echo "${REPO_OWNER_GIT}" | tr "[:upper:]" "[:lower:]")
readonly REPO_NAME_GIT=$(echo "${REPO_NAME_GIT}" | tr "[:upper:]" "[:lower:]")
readonly CURRENT_TIMESTAMP=$(date +%Y%m%d)

readonly STANDARIZED_NAME=$(echo "${INPUT_CLUSTER_NAME}" | tr "[:upper:]" "[:lower:]")
readonly DNS_SUBDOMAIN="${STANDARIZED_NAME}"
export CLUSTER_NAME="${STANDARIZED_NAME}"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

removeCluster() {
    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    EXIT_STATUS=$?

    shout "Fetching OLD_TIMESTAMP from cluster to be deleted"
	readonly OLD_TIMESTAMP=$(gcloud container clusters describe "${CLUSTER_NAME}" --zone="${GCLOUD_COMPUTE_ZONE}" --project="${GCLOUD_PROJECT_NAME}" --format=json | jq --raw-output '.resourceLabels."created-at"')


    shout "Deprovision cluster: \"${CLUSTER_NAME}\""
    date

    #save disk names while the cluster still exists to remove them later
    DISKS=$(kubectl get pvc --all-namespaces -o jsonpath="{.items[*].spec.volumeName}" | xargs -n1 echo)
    export DISKS

    #Delete cluster
    shout "Delete cluster $CLUSTER_NAME"
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/deprovision-gke-cluster.sh
    TMP_STATUS=$?
    if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

    #Delete orphaned disks
    shout "Delete orphaned PVC disks..."
    date
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-disks.sh
    TMP_STATUS=$?
    if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi


    shout "Delete Gateway DNS Record"
    date
    IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh
    TMP_STATUS=$?
    if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

    shout "Release Gateway IP Address"
    date
    IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/release-ip-address.sh
    TMP_STATUS=$?
    if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

    shout "Delete Remote Environments DNS Record"
    date
    IP_ADDRESS=${REMOTEENVS_IP_ADDRESS} DNS_FULL_NAME=${REMOTEENVS_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh
    TMP_STATUS=$?
    if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

    shout "Release Remote Environments IP Address"
    date
    IP_ADDRESS_NAME=${REMOTEENVS_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/release-ip-address.sh
    TMP_STATUS=$?
    if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

    shout "Delete temporary Kyma-Installer Docker image"
    date
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-image.sh
    TMP_STATUS=$?
    if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    shout "Job is finished ${MSG}"
    date
    set -e

    exit "${EXIT_STATUS}"
}

function cleanup() {
    OLD_CLUSTERS=$(gcloud container clusters list --filter="name~^${CLUSTER_NAME}" --format json | jq '.[].name' | tr -d '"')
    CLUSTERS_SIZE=$(echo "$OLD_CLUSTERS" | wc -l)
    if [[ "$CLUSTERS_SIZE" -gt 0 ]]; then
	    shout "Delete old cluster"
	    date
	    for CLUSTER in $OLD_CLUSTERS; do
		    removeCluster "${CLUSTER}"
	    done
    fi

}

# As is a periodic job executed on master, operate on triggering commit id
readonly COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)
KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/${STANDARIZED_NAME}/${REPO_OWNER}/${REPO_NAME}:COMMIT-${COMMIT_ID}"
export KYMA_INSTALLER_IMAGE

#Local variables
KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
INSTALLER_CONFIG="${KYMA_RESOURCES_DIR}/installer-config-cluster.yaml.tpl"
INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"

shout "Authenticate"
date
init
DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"

shout "Build Kyma-Installer Docker image"
date
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-image.sh"

shout "Cleanup"
date
cleanup

shout "Create new cluster"
date

shout "Reserve IP Address for Ingressgateway"
date
GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"
GATEWAY_IP_ADDRESS=$(IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/reserve-ip-address.sh")
echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"


shout "Create DNS Record for Ingressgateway IP"
date
GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"


shout "Reserve IP Address for Remote Environments"
date
REMOTEENVS_IP_ADDRESS_NAME="remoteenvs-${STANDARIZED_NAME}"
REMOTEENVS_IP_ADDRESS=$(IP_ADDRESS_NAME=${REMOTEENVS_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/reserve-ip-address.sh")
echo "Created IP Address for Remote Environments: ${REMOTEENVS_IP_ADDRESS}"


shout "Create DNS Record for Remote Environments IP"
date
REMOTEENVS_DNS_FULL_NAME="gateway.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
IP_ADDRESS=${REMOTEENVS_IP_ADDRESS} DNS_FULL_NAME=${REMOTEENVS_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"


shout "Provision cluster: \"${CLUSTER_NAME}\""
date

if [ -z "$MACHINE_TYPE" ]; then
      export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
fi
if [ -z "${CLUSTER_VERSION}" ]; then
      export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
fi

"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/provision-gke-cluster.sh"

shout "Install Tiller"
date
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
"${KYMA_SCRIPTS_DIR}"/install-tiller.sh


shout "Generate self-signed certificate"
date
DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
export DOMAIN
CERT_KEY=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-self-signed-cert.sh")
TLS_CERT=$(echo "${CERT_KEY}" | head -1)
TLS_KEY=$(echo "${CERT_KEY}" | tail -1)

shout "Apply Kyma config"
date

echo "Manual concatenating yamls"
"${KYMA_SCRIPTS_DIR}"/concat-yamls.sh "${INSTALLER_YAML}" "${INSTALLER_CONFIG}" "${INSTALLER_CR}" \
| sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" \
| sed -e "s/__DOMAIN__/${DOMAIN}/g" \
| sed -e "s/__REMOTE_ENV_IP__/${REMOTEENVS_IP_ADDRESS}/g" \
| sed -e "s/__TLS_CERT__/${TLS_CERT}/g" \
| sed -e "s/__TLS_KEY__/${TLS_KEY}/g" \
| sed -e "s/__EXTERNAL_PUBLIC_IP__/${GATEWAY_IP_ADDRESS}/g" \
| sed -e "s/__SKIP_SSL_VERIFY__/true/g" \
| sed -e "s/__VERSION__/0.0.1/g" \
| sed -e "s/__.*__//g" \
| kubectl apply -f-

"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-cluster-backup-config.sh"

shout "Trigger installation"
date
kubectl label installation/kyma-installation action=install
"${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout 30m

shout "Test Kyma"
date
"${KYMA_SCRIPTS_DIR}"/testing.sh

shout "End To End Test"
date
"${KYMA_SCRIPTS_DIR}"/e2e-testing.sh

shout "Success"