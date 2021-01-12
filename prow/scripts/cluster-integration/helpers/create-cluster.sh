#!/usr/bin/env bash

#Description: Creates a GKE cluster
#
#Expected vars:
# - TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS: absolute path for test-infra/prow/scripts/cluster-integration/helpers directory
# - STANDARIZED_NAME:
# - DNS_SUBDOMAIN: name of the GCP managed zone
# - DNS_DOMAIN: name of the cluster

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcloud.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcloud.sh"


function createCluster() {
	requiredVars=(
		CLUSTER_NAME
		STANDARIZED_NAME
		TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS
	)

	utils::check_required_vars "${requiredVars[@]}"

	if [[ -z "${PERFORMACE_CLUSTER_SETUP}" ]]; then
		requiredVars+=(
			GCLOUD_NETWORK_NAME
			GCLOUD_SUBNET_NAME
			DNS_SUBDOMAIN DNS_DOMAIN
		)
	fi

	utils::check_required_vars "${requiredVars[@]}"

	if [[ "${PERFORMACE_CLUSTER_SETUP}" == "" ]]; then

		log::info "Reserve IP Address for Ingressgateway"
		GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"
		GATEWAY_IP_ADDRESS=$(IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/reserve-ip-address.sh)
		echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

		log::info "Create DNS Record for Ingressgateway IP"
		GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
		gcloud::create_dns_record "${GATEWAY_IP_ADDRESS}" "${GATEWAY_DNS_FULL_NAME}"

		NETWORK_EXISTS=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/network-exists.sh")
		if [ "$NETWORK_EXISTS" -gt 0 ]; then
			log::info "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
			"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-network-with-subnet.sh"
		else
			log::info "Network ${GCLOUD_NETWORK_NAME} exists"
		fi

        export GATEWAY_IP_ADDRESS
	fi

	log::info "Provision cluster: \"${CLUSTER_NAME}\""

	if [ -z "$MACHINE_TYPE" ]; then
		export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
	fi
	if [ -z "${CLUSTER_VERSION}" ]; then
		export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
	fi
	env ADDITIONAL_LABELS="created-at=${CURRENT_TIMESTAMP}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/provision-gke-cluster.sh

}

createCluster