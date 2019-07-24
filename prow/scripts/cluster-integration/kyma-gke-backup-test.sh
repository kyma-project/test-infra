#!/usr/bin/env bash

set -o errexit

discoverUnsetVar=false

for var in REPO_OWNER REPO_NAME DOCKER_PUSH_REPOSITORY KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS KYMA_ARTIFACTS_BUCKET KYMA_BACKUP_RESTORE_BUCKET KYMA_BACKUP_CREDENTIALS CLOUDSDK_COMPUTE_ZONE; do
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

# Enforce lowercase
readonly REPO_OWNER=$(echo "${REPO_OWNER}" | tr '[:upper:]' '[:lower:]')
export REPO_OWNER
readonly REPO_NAME=$(echo "${REPO_NAME}" | tr '[:upper:]' '[:lower:]')
export REPO_NAME

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c10)

if [[ "$BUILD_TYPE" == "pr" ]]; then
    # In case of PR, operate on PR number
    readonly COMMON_NAME_PREFIX="gke-backup-pr"
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
    KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-backup-test/${REPO_OWNER}/${REPO_NAME}:PR-${PULL_NUMBER}"
    export KYMA_INSTALLER_IMAGE
elif [[ "$BUILD_TYPE" == "release" ]]; then
    readonly COMMON_NAME_PREFIX="gke-backup-rel"
    readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
    readonly RELEASE_VERSION=$(cat "${SCRIPT_DIR}/../../RELEASE_VERSION")
    shout "Read the release version from RELEASE_VERSION file, got: ${RELEASE_VERSION}"
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
else
    # Otherwise (master), operate on triggering commit id
    readonly COMMON_NAME_PREFIX="gke-backup-commit"
    readonly COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
    KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-backup-test/${REPO_OWNER}/${REPO_NAME}:COMMIT-${COMMIT_ID}"
    export KYMA_INSTALLER_IMAGE
fi

### Cluster name must be less than 40 characters!
export CLUSTER_NAME="${COMMON_NAME}"

readonly DNS_SUBDOMAIN="${CLUSTER_NAME}"
export GCLOUD_NETWORK_NAME="gke-backup-test-net"
export GCLOUD_SUBNET_NAME="gke-backup-test-subnet"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

removeCluster() {
    shout "Deprovision cluster: \"${CLUSTER_NAME}\""
    date

    #save disk names while the cluster still exists to remove them later
    DISKS=$(kubectl get pvc --all-namespaces -o jsonpath="{.items[*].spec.volumeName}" | xargs -n1 echo)
    export DISKS

    #Delete cluster
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/deprovision-gke-cluster.sh
}

function cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        shout "AN ERROR OCCURED! Take a look at preceding log entries."
        echo
    fi

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    removeCluster
    
    shout "Delete Gateway DNS Record"
	date
	GATEWAY_IP_ADDRESS=$(gcloud compute addresses describe "${CLUSTER_NAME}" --format json --region "${CLOUDSDK_COMPUTE_REGION}" | jq '.address' | tr -d '"')
	IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh
	TMP_STATUS=$?
	if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

    if [ ! -z "${GATEWAY_IP_ADDRESS_NAME}" ]; then
        shout "Release Gateway IP Address"
        date
        IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/release-ip-address.sh
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    fi

    shout "Delete temporary Kyma-Installer Docker image"
    date
    KYMA_INSTALLER_IMAGE="${KYMA_INSTALLER_IMAGE}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-image.sh
    TMP_STATUS=$?
    if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    shout "Job is finished ${MSG}"
    date
    set -e

    exit "${EXIT_STATUS}"
}

trap cleanup EXIT INT

#Local variables
KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
export KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

shout "Authenticate"
date
init
DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"

GATEWAY_IP_ADDRESS_NAME="${CLUSTER_NAME}"
GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"

export KYMA_INSTALLER_IMAGE
shout "Build Kyma-Installer Docker image"
date
KYMA_INSTALLER_IMAGE="${KYMA_INSTALLER_IMAGE}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-image.sh


shout "Create new cluster"
date

shout "Reserve IP Address for Ingressgateway"
date
GATEWAY_IP_ADDRESS=$(IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/reserve-ip-address.sh")
echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"


shout "Create DNS Record for Ingressgateway IP"
date
IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"


NETWORK_EXISTS=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/network-exists.sh")
if [ "$NETWORK_EXISTS" -gt 0 ]; then
    shout "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
    date
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-network-with-subnet.sh"
else
    shout "Network ${GCLOUD_NETWORK_NAME} exists"
fi

function provisionCluster() {
    shout "Provision cluster: \"${CLUSTER_NAME}\""
    date

    if [ -z "$MACHINE_TYPE" ]; then
        export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
    fi
    if [ -z "${CLUSTER_VERSION}" ]; then
        export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
    fi

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/provision-gke-cluster.sh"
}

function installKyma() {
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

    # shellcheck disable=SC2002
    sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" "${INSTALLER_YAML}" \
        | kubectl apply -f-

    # Generate backup config
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-cluster-backup-config.sh"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "cluster-certificate-overrides" \
        --data "global.tlsCrt=${TLS_CERT}" \
        --data "global.tlsKey=${TLS_KEY}"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "istio-overrides" \
        --data "gateways.istio-ingressgateway.loadBalancerIP=${GATEWAY_IP_ADDRESS}" \
        --label "component=istio"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "installation-config-overrides" \
        --data "global.domainName=${DOMAIN}" \
        --data "global.loadBalancerIP=${GATEWAY_IP_ADDRESS}"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "core-test-ui-acceptance-overrides" \
        --data "test.acceptance.ui.logging.enabled=true" \
        --label "component=core"

    sed -e "s/__VERSION__/0.0.1/g" "${INSTALLER_CR}" \
        | sed -e "s/__.*__//g" \
        | kubectl apply -f-

    shout "Installation triggered"
    date

    "${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout 30m
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-helm-certs.sh"

    shout "Success! Kyma installed"
}

BACKUP_FILE="${KYMA_SOURCES_DIR}"/docs/backup/assets/backup.yaml
BACKUP_NAME=$(cat /proc/sys/kernel/random/uuid)

function takeBackup() {

    shout "Take backup"
    date

    sed -i "s/name: kyma-backup/name: ${BACKUP_NAME}/g" "${BACKUP_FILE}"
    kubectl apply -f "${BACKUP_FILE}"
    sleep 30

    attempts=3
    retryTimeInSec="45"
    for ((i=1; i<=attempts; i++)); do
        STATUS=$(kubectl get backup "${BACKUP_NAME}" -n kyma-backup -o jsonpath='{.status.phase}')
        if [ "${STATUS}" == "Completed" ]; then
            shout "Backup completed"
            break
        elif [ "${STATUS}" == "Failed" ] || [ "${STATUS}" == "FailedValidation" ]; then
            shout "Backup failed"
            exit 1
        fi
        
        if [[ "${i}" -lt "${attempts}" ]]; then
            echo "Unable to get backup status, let's wait ${retryTimeInSec} seconds and retry. Attempts ${i} of ${attempts}."
        else
            echo "Unable to get backup status after ${attempts} attempts, giving up."
            exit 1
        fi
        sleep ${retryTimeInSec}
    done
}

function restoreKyma() {

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    shout "Install Velero CLI"
    date

    wget -q https://github.com/heptio/velero/releases/download/v1.0.0/velero-v1.0.0-linux-amd64.tar.gz && \
    tar -xvf velero-v1.0.0-linux-amd64.tar.gz && \
    mv velero-v1.0.0-linux-amd64/velero /usr/local/bin && \
    rm -rf velero-v1.0.0-linux-amd64 velero-v1.0.0-linux-amd64.tar.gz 

    CLOUD_PROVIDER="gcp"

    shout "Install Velero Server"
    date
    velero install --bucket "$BACKUP_RESTORE_BUCKET" --provider "$CLOUD_PROVIDER" --secret-file "$BACKUP_CREDENTIALS" --restore-only --wait

    sleep 15

    shout "Check if the backup ${BACKUP_NAME} exists"
    date
    attempts=3
    for ((i=1; i<=attempts; i++)); do
        result=$(velero get backup "${BACKUP_NAME}")
        if [[ "$result" == *"NAME"* ]]; then
            echo "Backup ${BACKUP_NAME} exists"
            break
        elif [[ "${i}" == "${attempts}" ]]; then
            echo "ERROR: backup ${BACKUP_NAME} not found"
            exit 1
        fi
        echo "Sleep for 15 seconds"
        sleep 15
    done
    

    shout "Restore Kyma CRDs, Services and Endpoints"
    date
    velero restore create --from-backup "${BACKUP_NAME}" --include-resources customresourcedefinitions.apiextensions.k8s.io,services,endpoints --include-cluster-resources --wait

    sleep 30

    shout "Restore the rest of Kyma"
    date

    attempts=3
    for ((i=1; i<=attempts; i++)); do
        
        velero restore create --from-backup "${BACKUP_NAME}" --exclude-resources customresourcedefinitions.apiextensions.k8s.io,services,endpoints, --include-cluster-resources --restore-volumes --wait

        sleep 60

        echo "Check if VirtualServices are restored"
        
        result=$(kubectl get virtualservices -n kyma-system)
        if [[ "$result" == *"NAME"* ]]; then
            echo "VirtualServices are restored"
            break
        elif [[ "${i}" == "${attempts}" ]]; then
            echo "ERROR: restoring VirtualServices failed"
            exit 1
        fi

        echo "Sleep for 30 seconds"
        sleep 30
    done

    set -e
}

provisionCluster
installKyma

shout "Run tests before backup"
date
cd "${KYMA_SCRIPTS_DIR}"
set +e
ACTION="testBeforeBackup" ./e2e-testing.sh
TEST_STATUS=$?
if [ ${TEST_STATUS} -ne 0 ]
then
    shout "Tests before backup failed"
    exit 1
fi

takeBackup
removeCluster

### Restore phase starts here

provisionCluster
restoreKyma

shout "Run tests after restore"
date
cd "${KYMA_SCRIPTS_DIR}"
set +e
ACTION="testAfterRestore" ./e2e-testing.sh
TEST_STATUS=$?
if [ ${TEST_STATUS} -ne 0 ]
then
    shout "Tests after restore failed"
    exit 1
fi

shout "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
