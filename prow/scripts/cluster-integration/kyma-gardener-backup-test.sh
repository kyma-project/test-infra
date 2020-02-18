#!/usr/bin/env bash

#Description: Kyma backup plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to test the backup scenario for Kyma on a real Gardener cluster.
#
#
#Expected vars:
#
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - REGION - Gardener compute region
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME Name of the azure secret configured in the gardener project to access the cloud provider
# - MACHINE_TYPE (optional): Machine type
# - CLUSTER_VERSION (optional): Kubernetes version
#

set -o errexit

discoverUnsetVar=false

for var in KYMA_PROJECT_DIR REGION GARDENER_KYMA_PROW_KUBECONFIG GARDENER_KYMA_PROW_PROJECT_NAME GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
export KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
readonly COMMON_NAME_PREFIX="grdnr"
readonly STANDARIZED_NAME=$(echo "${COMMON_NAME_PREFIX}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
readonly DNS_SUBDOMAIN="${STANDARIZED_NAME}"

### Cluster name must be less than 20 characters!
export CLUSTER_NAME="${STANDARIZED_NAME}"

export RESOURCE_GROUP="shoot--${GARDENER_KYMA_PROW_PROJECT_NAME}--${CLUSTER_NAME}"
export AZURE_BACKUP_RESOURCE_GROUP="Velero_Backups"

BACKUP_NAME=$(cat /proc/sys/kernel/random/uuid)

### For generate-cluster-backup-config.sh
export BACKUP_RESTORE_BUCKET="velero"
export CLOUD_PROVIDER="azure"
AZURE_STORAGE_ACCOUNT_ID="velero${RANDOM_NAME_SUFFIX}"
export AZURE_STORAGE_ACCOUNT_ID
export PROVIDER_PLUGIN_IMAGE="velero/velero-plugin-for-microsoft-azure:v1.0.0"
export API_TIMEOUT="3m0s"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/testing-helpers.sh"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers/kyma-cli.sh"

removeCluster() {
  shout "Deprovision cluster: \"${CLUSTER_NAME}\""
  date
  # Export envvars for the script
  export GARDENER_CLUSTER_NAME=${CLUSTER_NAME}
  export GARDENER_PROJECT_NAME=${GARDENER_KYMA_PROW_PROJECT_NAME}
  export GARDENER_CREDENTIALS=${GARDENER_KYMA_PROW_KUBECONFIG}
  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/deprovision-gardener-cluster.sh
}

#!Put cleanup code in this function! Function is executed at exit from the script and on interuption.
cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        shout "AN ERROR OCCURED! Take a look at preceding log entries."
        echo
    fi

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    shout "Remove DNS Record for Ingressgateway"
    GATEWAY_DNS_FULL_NAME="*.${DOMAIN}."
    GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"

    GATEWAY_IP_ADDRESS=$(az network public-ip show -g "${RESOURCE_GROUP}" -n "${GATEWAY_IP_ADDRESS_NAME}" --query ipAddress -o tsv)
    TMP_STATUS=$?
    if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    if [[ -n ${GATEWAY_IP_ADDRESS} ]];then
        echo "Fetched Azure Gateway IP: ${GATEWAY_IP_ADDRESS}"
        # only try to delete the dns record if the ip address has been found
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${GATEWAY_DNS_FULL_NAME=}" --address="${GATEWAY_IP_ADDRESS}" --dryRun=false
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    else
        echo "Could not fetch Azure Gateway IP: GATEWAY_IP_ADDRESS variable is empty. Something went wrong. Failing"
    fi
    TMP_STATUS=$?
    if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

    shout "Remove DNS Record for Apiserver Proxy IP"
    APISERVER_DNS_FULL_NAME="apiserver.${DOMAIN}."
    APISERVER_IP_ADDRESS=$(gcloud dns record-sets list --zone "${CLOUDSDK_DNS_ZONE_NAME}" --name "${APISERVER_DNS_FULL_NAME}" --format="value(rrdatas[0])")
    if [[ -n ${APISERVER_IP_ADDRESS} ]]; then
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${APISERVER_DNS_FULL_NAME}" --address="${APISERVER_IP_ADDRESS}" --dryRun=false
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    fi

    rm -rf "${TMP_DIR}"

    if [ -n "${CLEANUP_CLUSTER}" ]; then
      removeCluster
    fi

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    shout "Job is finished ${MSG}"
    date
    set -e

    exit "${EXIT_STATUS}"
}

trap cleanup EXIT INT

provisionCluster() {
  shout "Provision cluster: \"${CLUSTER_NAME}\""
  date

  GARDENER_CLUSTER_VERSION="1.16.7"

  if [ -z "$MACHINE_TYPE" ]; then
        export MACHINE_TYPE="Standard_D4_v3"
  fi

  CLEANUP_CLUSTER="true"
  (
  set -x
  kyma provision gardener \
          --target-provider azure --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" \
          --name "${CLUSTER_NAME}" --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
          --region "${REGION}" -z "1" -t "${MACHINE_TYPE}" --disk-size 50 --disk-type=Standard_LRS --extra vnetcidr="10.250.0.0/16" \
          --nodes 4 \
          --kube-version="${GARDENER_CLUSTER_VERSION}"
  )
}

function azureAuthenticating() {
	shout "Authenticating to azure"
	date

	az login --service-principal -u "${AZURE_SUBSCRIPTION_APP_ID}" -p "${AZURE_SUBSCRIPTION_SECRET}" --tenant "${AZURE_SUBSCRIPTION_TENANT}"
	az account set --subscription "${AZURE_SUBSCRIPTION_ID}"
}

function createPublicIPandDNS() {
	# IP address and DNS for Ingressgateway
	shout "Reserve IP Address for Ingressgateway"
	date

	GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"
	az network public-ip create -g "${RESOURCE_GROUP}" -n "${GATEWAY_IP_ADDRESS_NAME}" -l "${REGION}" --allocation-method static

	GATEWAY_IP_ADDRESS=$(az network public-ip show -g "${RESOURCE_GROUP}" -n "${GATEWAY_IP_ADDRESS_NAME}" --query ipAddress -o tsv)
	echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

	shout "Create DNS Record for Ingressgateway IP"
	date

  DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --project="${CLOUDSDK_CORE_PROJECT}" --format="value(dnsName)")"
  export DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"

	GATEWAY_DNS_FULL_NAME="*.${DOMAIN}."
	IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-dns-record.sh

  shout "Generate self-signed certificate"
  date

  CERT_KEY=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-self-signed-cert.sh")
  TLS_CERT=$(echo "${CERT_KEY}" | head -1)
  TLS_KEY=$(echo "${CERT_KEY}" | tail -1)
  export TLS_CERT
  export TLS_KEY
}

installKyma() {
  shout "Create backup resource group ${AZURE_BACKUP_RESOURCE_GROUP}"
  az group create -n "${AZURE_BACKUP_RESOURCE_GROUP}" --location westeurope

  shout "Create storage account ${AZURE_STORAGE_ACCOUNT_ID}"
  az storage account create \
      --name "${AZURE_STORAGE_ACCOUNT_ID}" \
      --resource-group "${AZURE_BACKUP_RESOURCE_GROUP}" \
      --sku Standard_GRS \
      --encryption-services blob \
      --https-only true \
      --kind BlobStorage \
      --access-tier Hot

  shout "Create storage container ${BACKUP_RESTORE_BUCKET}"
  az storage container create -n "${BACKUP_RESTORE_BUCKET}" --public-access off --account-name "${AZURE_STORAGE_ACCOUNT_ID}"

cat << EOF  > ./credentials-velero
AZURE_SUBSCRIPTION_ID=${AZURE_SUBSCRIPTION_ID}
AZURE_TENANT_ID=${AZURE_SUBSCRIPTION_TENANT}
AZURE_CLIENT_ID=${AZURE_SUBSCRIPTION_APP_ID}
AZURE_CLIENT_SECRET=${AZURE_SUBSCRIPTION_SECRET}
AZURE_RESOURCE_GROUP=${RESOURCE_GROUP}
AZURE_CLOUD_NAME=AzurePublicCloud
EOF

  export BACKUP_CREDENTIALS="credentials-velero"

  kubectl create namespace "kyma-installer"

  # Generate backup config
  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-cluster-backup-config.sh"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "istio-overrides" \
  --data "gateways.istio-ingressgateway.loadBalancerIP=${GATEWAY_IP_ADDRESS}" \
  --label "component=istio"

  shout "Installing Kyma"
  date

  kyma install \
    --ci \
    --source "latest" \
    --domain "${DOMAIN}" \
    --tlsCert "${TLS_CERT}" \
    --tlsKey "${TLS_KEY}" \
    --timeout 90m

  if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
    shout "Create DNS Record for Apiserver proxy IP"
    date
    APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    APISERVER_DNS_FULL_NAME="apiserver.${DOMAIN}."
    IP_ADDRESS=${APISERVER_IP_ADDRESS} DNS_FULL_NAME=${APISERVER_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
  fi

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-helm-certs.sh"
}

function takeBackup() {
    shout "Take backup"
    date

    BACKUP_FILE="${KYMA_SOURCES_DIR}"/docs/backup/assets/backup.yaml

    sed -i "s/name: kyma-backup/name: ${BACKUP_NAME}/g" "${BACKUP_FILE}"
    kubectl apply -f "${BACKUP_FILE}"
    sleep 45

    attempts=3
    retryTimeInSec="30"
    for ((i=1; i<=attempts; i++)); do
        STATUS=$(kubectl get backup "${BACKUP_NAME}" -n kyma-system -o jsonpath='{.status.phase}')
        if [ "${STATUS}" == "Completed" ]; then
            shout "Backup completed"
            break
        elif [ "${STATUS}" == "Failed" ] || [ "${STATUS}" == "FailedValidation" ]; then
            shout "Backup ${BACKUP_NAME} failed with the status: ${STATUS}"
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

    wget -q https://github.com/vmware-tanzu/velero/releases/download/v1.3.1/velero-v1.3.1-linux-amd64.tar.gz && \
    tar -xvf velero-v1.3.1-linux-amd64.tar.gz && \
    mv velero-v1.3.1-linux-amd64/velero /usr/local/bin && \
    rm -rf velero-v1.3.1-linux-amd64 velero-v1.3.1-linux-amd64.tar.gz

    E2E_TESTING_ENV_FILE="${KYMA_SCRIPTS_DIR}/e2e-testing.env"
    if [[ -f "${E2E_TESTING_ENV_FILE}" ]]; then
	    # shellcheck disable=SC1090
    	source "${E2E_TESTING_ENV_FILE}"
    fi

    VELERO_PLUGIN_IMAGES="${PROVIDER_PLUGIN_IMAGE},${ADDITIONAL_VELERO_PLUGIN_IMAGES:-eu.gcr.io/kyma-project/backup-plugins:c08e6274}"

    shout "Install Velero Server"
    date
    velero install \
        --bucket "${BACKUP_RESTORE_BUCKET}" \
        --provider "${CLOUD_PROVIDER}" \
        --secret-file "${BACKUP_CREDENTIALS}" \
        --plugins "${VELERO_PLUGIN_IMAGES}" \
        --backup-location-config resourceGroup="${AZURE_BACKUP_RESOURCE_GROUP}",storageAccount="${AZURE_STORAGE_ACCOUNT_ID}" \
        --snapshot-location-config apiTimeout="${API_TIMEOUT}",resourceGroup="${AZURE_BACKUP_RESOURCE_GROUP}" \
        --restore-only \
        --wait

    sleep 15

    shout "Check if the backup ${BACKUP_NAME} exists"
    date
    attempts=3
    for ((i=1; i<=attempts; i++)); do
        result=$(velero get backup "${BACKUP_NAME}")
        if [[ "${result}" == *"NAME"* ]]; then
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
    velero restore create --from-backup "${BACKUP_NAME}" --include-resources customresourcedefinitions.apiextensions.k8s.io,services,endpoints --wait

    sleep 30

    shout "Restore the rest of Kyma"
    date

    attempts=3
    for ((i=1; i<=attempts; i++)); do
        
        velero restore create --from-backup "${BACKUP_NAME}" --exclude-resources customresourcedefinitions.apiextensions.k8s.io,services,endpoints --restore-volumes --wait

        sleep 60

        echo "Check if VirtualServices are restored"
        
        result=$(kubectl get virtualservices -n kyma-system)
        if [[ "${result}" == *"NAME"* ]]; then
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

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

init
azureAuthenticating

export INSTALL_DIR=${TMP_DIR}
install::kyma_cli

provisionCluster
createPublicIPandDNS #Try with gardener automatic assigned domains
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
