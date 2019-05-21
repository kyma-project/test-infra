#!/usr/bin/env bash

#Description: Installs Kyma in a given GKE cluster
#
#Expected vars:
# - REMOTEENVS_IP_ADDRESS: static IP for remote env
# - GATEWAY_IP_ADDRESS: static IP for gateway
# - DOCKER_PUSH_REPOSITORY: name of the docker registry where images are pushed
# - KYMA_SOURCES_DIR: absolute path for kyma sources directory
# - DOCKER_PUSH_DIRECTORY: directory for docker images where it should be pushed
# - GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path
# - STANDARIZED_NAME: a variation of cluster name
# - REPO_OWNER: Kyma repository owner
# - REPO_NAME: name of the Kyma repository
# - CURRENT_TIMESTAMP: Current timestamp which is computed as $(date +%Y%m%d) 
# - DOMAIN: Combination of gcloud managed-zones and cluster name "${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
function installKyma() {

	kymaUnsetVar=false

	for var in REMOTEENVS_IP_ADDRESS GATEWAY_IP_ADDRESS DOCKER_PUSH_REPOSITORY KYMA_SOURCES_DIR DOCKER_PUSH_DIRECTORY GOOGLE_APPLICATION_CREDENTIALS STANDARIZED_NAME REPO_OWNER REPO_NAME CURRENT_TIMESTAMP DOMAIN; do
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
	INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"
	PROMTAIL_CONFIG_NAME=promtail-k8s-1-14.yaml
	# shellcheck disable=SC1090
    source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/generate-and-export-letsencrypt-TLS-cert.sh

	shout "Apply Kyma config"
	date

	"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "istio-overrides" \
        --data "gateways.istio-ingressgateway.loadBalancerIP:${GATEWAY_IP_ADDRESS}" \
        --label "component=istio"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "knative-serving-overrides" \
        --data "knative-serving.domainName=${DOMAIN}" \
        --label "component=knative-serving"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "installation-config-overrides" \
        --data "global.domainName=${DOMAIN}" \
        --data "global.loadBalancerIP=${GATEWAY_IP_ADDRESS}" \
        --data "nginx-ingress.controller.service.loadBalancerIP=${REMOTEENVS_IP_ADDRESS}"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "cluster-certificate-overrides" \
        --data "global.tlsCrt=${TLS_CERT}" \
        --data "global.tlsKey=${TLS_KEY}"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "core-test-ui-acceptance-overrides" \
        --data "test.acceptance.ui.logging.enabled=true" \
        --label "component=core"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "intallation-logging-overrides" \
        --data "global.logging.promtail.config.name=${PROMTAIL_CONFIG_NAME}" \
        --label "component=logging"

	cat "${INSTALLER_YAML}" \
		| sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" \
		| kubectl apply -f-

	waitUntilInstallerApiAvailable

	shout "Trigger installation"
	date

    sed -e "s/__VERSION__/0.0.1/g" "${INSTALLER_CR}"  | sed -e "s/__.*__//g" | kubectl apply -f-
	kubectl label installation/kyma-installation action=install --overwrite
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

installKyma
