#!/usr/bin/env bash

#Description: Creates a GKE cluster
#
#Expected vars:
# - TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS: absolute path for test-infra/prow/scripts/cluster-integration/helpers directory
# - STANDARIZED_NAME:
# - DNS_SUBDOMAIN: name of the GCP managed zone
# - DNS_DOMAIN: name of the cluster

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

function createCluster() {
	discoverUnsetVar=false

	for var in CLUSTER_NAME STANDARIZED_NAME TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS; do
		if [ -z "${!var}" ] ; then
			echo "ERROR: $var is not set"
			discoverUnsetVar=true
		fi
	done

	if [[ "${PERFORMACE_CLUSTER_SETUP}" == "" ]]; then
		for var in GCLOUD_NETWORK_NAME GCLOUD_SUBNET_NAME DNS_SUBDOMAIN DNS_DOMAIN; do
			if [ -z "${!var}" ] ; then
				echo "ERROR: $var is not set"
				discoverUnsetVar=true
			fi
		done
	fi

	if [ "${discoverUnsetVar}" = true ] ; then
		exit 1
	fi

	if [[ "${PERFORMACE_CLUSTER_SETUP}" == "" ]]; then

		shout "Reserve IP Address for Ingressgateway"
		date
		GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"
		GATEWAY_IP_ADDRESS=$(IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/reserve-ip-address.sh)
		echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

		shout "Create DNS Record for Ingressgateway IP"
		date
		GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
		IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-dns-record.sh

		NETWORK_EXISTS=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/network-exists.sh")
		if [ "$NETWORK_EXISTS" -gt 0 ]; then
			shout "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
			date
			"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-network-with-subnet.sh"
		else
			shout "Network ${GCLOUD_NETWORK_NAME} exists"
		fi

        export GATEWAY_IP_ADDRESS
	fi

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

createCluster