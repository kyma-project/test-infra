#!/usr/bin/env bash

#Description: Creates a GKE cluster
#
#Expected vars:
# - TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS: absolute path for test-infra/prow/scripts/cluster-integration/helpers directory 
# - STANDARIZED_NAME:
# - TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS: absolute path for test-infra/prow/scripts/cluster-integration/helpers directory
# - DNS_SUBDOMAIN: name of the GCP managed zone
# - DNS_DOMAIN: name of the cluster

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

function createCluster() {
	discoverUnsetVar=false

	for var in STANDARIZED_NAME TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS DNS_SUBDOMAIN DNS_DOMAIN; do
		if [ -z "${!var}" ] ; then
			echo "ERROR: $var is not set"
			discoverUnsetVar=true
		fi
	done
	if [ "${discoverUnsetVar}" = true ] ; then
		exit 1
	fi
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

    export REMOTEENVS_IP_ADDRESS
    export GATEWAY_IP_ADDRESS
}

createCluster