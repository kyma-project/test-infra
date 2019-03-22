#!/usr/bin/env bash

#Description: Installs Kyma in a given GKE cluster
#
#Expected vars:
# - REMOTEENVS_IP_ADDRESS: static IP for remote env
# - GATEWAY_IP_ADDRESS: static IP for gateway
# - DOCKER_PUSH_REPOSITORY: name of the docker registry where images are pushed
# - KYMA_SOURCES_DIR: absolute path for kyma sources directory
# - DOCKER_PUSH_DIRECTORY: directory for docker images where it should be pushed
# - STANDARIZED_NAME: a variation of cluster name
# - REPO_OWNER: Kyma repository owner
# - REPO_NAME: name of the Kyma repository
# - CURRENT_TIMESTAMP: Current timestamp which is computed as $(date +%Y%m%d) 
# - DOMAIN: Combination of gcloud managed-zones and cluster name "${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
function installKyma() {

	kymaUnsetVar=false

	for var in REMOTEENVS_IP_ADDRESS GATEWAY_IP_ADDRESS DOCKER_PUSH_REPOSITORY KYMA_SOURCES_DIR DOCKER_PUSH_DIRECTORY STANDARIZED_NAME REPO_OWNER REPO_NAME CURRENT_TIMESTAMP DOMAIN; do
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
	# shellcheck disable=SC2153
    KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"
	INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
	INSTALLER_CONFIG="${KYMA_RESOURCES_DIR}/installer-config-cluster.yaml.tpl"
	INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"
	
	# shellcheck disable=SC1090
    source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/generate-and-export-letsencrypt-TLS-cert.sh

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

function getHelmCerts() {
    RETRY_COUNT=3
    RETRY_TIME_SEC=5

    for (( i = 0; i < RETRY_COUNT; i++ )); do
        mkdir -p "$(helm home)"

        echo "---> Get Helm secrets and put then into $(helm home)"
        kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 --decode > "$(helm home)/ca.pem"
        kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 --decode > "$(helm home)/cert.pem"
        kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 --decode > "$(helm home)/key.pem"

        if [[ "${i}" -lt "${RETRY_COUNT}" ]]; then
            echo "---> Unable to get Helm Certs. Waiting for ${RETRY_TIME_SEC}. Attempt ${i} of ${RETRY_COUNT}"
        else
            echo "---> Unable to get Helm Certs after ${RETRY_COUNT} attempts. Exitting"
            exit 1
        fi

        sleep "${RETRY_TIME_SEC}"
    done
}

installKyma
getHelmCerts
