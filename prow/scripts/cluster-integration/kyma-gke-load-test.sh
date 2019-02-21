#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.
discoverUnsetVar=false

for var in INPUT_CLUSTER_NAME DOCKER_PUSH_REPOSITORY DOCKER_PUSH_DIRECTORY KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_COMPUTE_ZONE CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS SLACK_CLIENT_TOKEN LT_TIMEOUT LT_REQS_PER_ROUTINE LOAD_TEST_SLACK_CLIENT_CHANNEL_ID; do
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

export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"
export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"

readonly REPO_OWNER="kyma-project"
readonly REPO_NAME="kyma"
readonly CURRENT_TIMESTAMP=$(date +%Y%m%d)

readonly STANDARIZED_NAME=$(echo "${INPUT_CLUSTER_NAME}" | tr "[:upper:]" "[:lower:]")
readonly DNS_SUBDOMAIN="${STANDARIZED_NAME}"

export CLUSTER_NAME="${STANDARIZED_NAME}"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

function removeCluster() {
	#Turn off exit-on-error so that next step is executed even if previous one fails.
	set +e

    # CLUSTER_NAME variable is used in other scripts so we need to change it for a while
    ORIGINAL_CLUSTER_NAME=${CLUSTER_NAME}
	CLUSTER_NAME=$1

	EXIT_STATUS=$?

    shout "Fetching OLD_TIMESTAMP from cluster to be deleted"
	readonly OLD_TIMESTAMP=$(gcloud container clusters describe "${CLUSTER_NAME}" --zone="${GCLOUD_COMPUTE_ZONE}" --project="${GCLOUD_PROJECT_NAME}" --format=json | jq --raw-output '.resourceLabels."created-at"')

	shout "Delete cluster $CLUSTER_NAME"
	"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/deprovision-gke-cluster.sh
	TMP_STATUS=$?
	if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

	shout "Delete Gateway DNS Record"
	date
	GATEWAY_IP_ADDRESS=$(gcloud compute addresses describe "${CLUSTER_NAME}" --format json --region "${CLOUDSDK_COMPUTE_REGION}" | jq '.address' | tr -d '"')
	GATEWAY_DNS_FULL_NAME="*.${CLUSTER_NAME}.${DNS_NAME}"
	IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh
	TMP_STATUS=$?
	if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

	shout "Release Gateway IP Address"
	date
	GATEWAY_IP_ADDRESS_NAME=${CLUSTER_NAME}
	IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/release-ip-address.sh
	TMP_STATUS=$?
	if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

	shout "Delete Remote Environments DNS Record"
	date
	REMOTEENVS_IP_ADDRESS=$(gcloud compute addresses describe "remoteenvs-${CLUSTER_NAME}" --format json --region "${CLOUDSDK_COMPUTE_REGION}" | jq '.address' | tr -d '"')
	REMOTEENVS_DNS_FULL_NAME="gateway.${CLUSTER_NAME}.${DNS_NAME}"
	IP_ADDRESS=${REMOTEENVS_IP_ADDRESS} DNS_FULL_NAME=${REMOTEENVS_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh
	TMP_STATUS=$?
	if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

	shout "Release Remote Environments IP Address"
	date
	REMOTEENVS_IP_ADDRESS_NAME="remoteenvs-${CLUSTER_NAME}"
	IP_ADDRESS_NAME=${REMOTEENVS_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/release-ip-address.sh
	TMP_STATUS=$?
	if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

	shout "Delete temporary Kyma-Installer Docker image"
	date


    KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/${STANDARIZED_NAME}/${REPO_OWNER}/${REPO_NAME}:${OLD_TIMESTAMP}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-image.sh
	TMP_STATUS=$?
	if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

	MSG=""
	if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
	shout "Job is finished ${MSG}"
	date

    # Revert previous value for CLUSTER_NAME variable
    CLUSTER_NAME=${ORIGINAL_CLUSTER_NAME}
	set -e
}

function createCluster() {
	shout "Reserve IP Address for Ingressgateway"
	date
	GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"
	GATEWAY_IP_ADDRESS=$(IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/reserve-ip-address.sh)
	echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

	shout "Create DNS Record for Ingressgateway IP"
	date
	GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
	IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-dns-record.sh

	shout "Reserve IP Address for Remote Environments"
	date
	REMOTEENVS_IP_ADDRESS_NAME="remoteenvs-${STANDARIZED_NAME}"
	REMOTEENVS_IP_ADDRESS=$(IP_ADDRESS_NAME=${REMOTEENVS_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/reserve-ip-address.sh)
	echo "Created IP Address for Remote Environments: ${REMOTEENVS_IP_ADDRESS}"

	shout "Create DNS Record for Remote Environments IP"
	date
	REMOTEENVS_DNS_FULL_NAME="gateway.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
	IP_ADDRESS=${REMOTEENVS_IP_ADDRESS} DNS_FULL_NAME=${REMOTEENVS_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-dns-record.sh

	shout "Provision cluster: \"${CLUSTER_NAME}\""
	date
	
	if [ -z "$MACHINE_TYPE" ]; then
		export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
	fi
	if [ -z "${CLUSTER_VERSION}" ]; then
		export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
	fi
	env ADDITIONAL_LABELS="created-at=${CURRENT_TIMESTAMP}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/provision-gke-cluster.sh
}

function waitUntilInstallerApiAvailable() {
    shout "Waiting for Installer API"

	attempts=5
    for ((i=1; i<=attempts; i++)); do
        numberOfLines=$(kubectl api-versions | grep -c "installer.kyma-project.io")
        if [[ "$numberOfLines" == "1" ]]; then
            echo "API found"
            break
        elif [[ "${i}" == "${attempts}" ]]; then
            echo "ERROR: API not found, exit"
            exit 1
        fi

        echo "Sleep for 3 seconds"
        sleep 3
    done
}

function generateAndExportLetsEncryptCert() {
	shout "Generate lets encrypt certificate"
	date

    mkdir letsencrypt
    cp /etc/credentials/sa-gke-kyma-integration/service-account.json letsencrypt
    docker run  --name certbot \
        --rm  \
        -v "$(pwd)/letsencrypt:/etc/letsencrypt"    \
        certbot/dns-google \
        certonly \
        -m "kyma.bot@sap.com" \
        --agree-tos \
        --no-eff-email \
        --dns-google \
        --dns-google-credentials /etc/letsencrypt/service-account.json \
        --server https://acme-v02.api.letsencrypt.org/directory \
        --dns-google-propagation-seconds=600 \
        -d "*.${DOMAIN}"

    TLS_CERT=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/fullchain.pem | tr -d '\n')
    export TLS_CERT
    TLS_KEY=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/privkey.pem   | tr -d '\n')
    export TLS_KEY

}

function installKyma() {

	kymaUnsetVar=false

	for var in REMOTEENVS_IP_ADDRESS GATEWAY_IP_ADDRESS ; do
    	if [ -z "${!var}" ] ; then
        	echo "ERROR: $var is not set"
        	kymaUnsetVar=true
    	fi
	done
	if [ "${kymaUnsetVar}" = true ] ; then
    	exit 1
	fi

	KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/${STANDARIZED_NAME}/${REPO_OWNER}/${REPO_NAME}:${CURRENT_TIMESTAMP}"

	shout "Build Kyma-Installer Docker image"
	date
	KYMA_INSTALLER_IMAGE="${KYMA_INSTALLER_IMAGE}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-image.sh

	KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"
	INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
	INSTALLER_CONFIG="${KYMA_RESOURCES_DIR}/installer-config-cluster.yaml.tpl"
	INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"
	

	DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
	export DOMAIN
    generateAndExportLetsEncryptCert

	shout "Apply Kyma config"
	date

	"${KYMA_SCRIPTS_DIR}"/concat-yamls.sh "${INSTALLER_YAML}" "${INSTALLER_CONFIG}" \
		| sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" \
		| sed -e "s/__DOMAIN__/${DOMAIN}/g" \
		| sed -e "s/__REMOTE_ENV_IP__/${REMOTEENVS_IP_ADDRESS}/g" \
		| sed -e "s#__TLS_CERT__#${TLS_CERT}#g" \
		| sed -e "s#__TLS_KEY__#${TLS_KEY}#g" \
		| sed -e "s/__EXTERNAL_PUBLIC_IP__/${GATEWAY_IP_ADDRESS}/g" \
		| sed -e "s/__SKIP_SSL_VERIFY__/true/g" \
		| sed -e "s/__VERSION__/0.0.1/g" \
		| sed -e "s/__.*__//g" \
		| kubectl apply -f-

	waitUntilInstallerApiAvailable

	shout "Trigger installation"
	date

    sed -e "s/__VERSION__/0.0.1/g" "${INSTALLER_CR}"  | sed -e "s/__.*__//g" | kubectl apply -f-
	kubectl label installation/kyma-installation action=install
	"${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout 30m
}

function waitUntilHPATestIsDone {
    local succeeded='false'
    while true; do
        hpa_test_status=$(kubectl get po -n kyma-system load-test -ojsonpath="{.status.phase}")
        if [ "${hpa_test_status}" != "Running" ]; then
            if [ "${hpa_test_status}" == "Succeeded" ]; then
                succeeded='true'
            fi
            break
        fi
        sleep 5
        echo "HPA tests are in progress..."
    done

    if [ $succeeded == 'true' ]; then
        shout "Load test successfully completed!"
    else
        kubectl describe pod load-test -n kyma-system
        shout "Load test failed!"
        shout "Please check logs by executing: kubectl logs -n kyma-system load-test"
        exit 1
    fi
}

function installLoadTest() {
	LT_FOLDER=${KYMA_SOURCES_DIR}/tools/load-test
	LT_FOLDER_CHART=${LT_FOLDER}'/deploy/chart/load-test'

	shout "Executing load tests..."

	shout "Installing helm chart..."
	helm install --set slackClientToken="${SLACK_CLIENT_TOKEN}" \
				--set slackClientChannelId="${LOAD_TEST_SLACK_CLIENT_CHANNEL_ID}" \
				--set loadTestExecutionTimeout="${LT_TIMEOUT}" \
				--set reqsPerRoutine="${LT_REQS_PER_ROUTINE}" \
				"${LT_FOLDER_CHART}" \
				--namespace=kyma-system \
				--name=load-test \
				--timeout=700 \
				--wait 

	waitUntilHPATestIsDone
	
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

shout "Authenticate"
date
init

DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"
export DNS_DOMAIN

shout "Cleanup"
date
cleanup

shout "Create new cluster"
date
createCluster

shout "Install tiller"
date

shout "Account is:"
gcloud config get-value account

kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
"${KYMA_SCRIPTS_DIR}"/install-tiller.sh

shout "Install kyma"
date
installKyma

shout "Install load-test"
date
installLoadTest

shout "Cleanup after load test"
date
cleanup
