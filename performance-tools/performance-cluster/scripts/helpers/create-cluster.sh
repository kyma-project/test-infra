#!/usr/bin/env bash

#Description: Creates a GKE cluster
#
#Expected vars:
# - TEST_INFRA_PERFORMANCE_TOOLS_CLUSTER_SCRIPTS: absolute path for test-infra/performance-tools/performance-cluster/scripts/helpers directory
# - STANDARIZED_NAME:
# - DNS_SUBDOMAIN: name of the GCP managed zone
# - DNS_DOMAIN: name of the cluster

source "${CURRENT_PATH}/scripts/library.sh"

function createCluster() {
	discoverUnsetVar=false

	for var in CLUSTER_NAME STANDARIZED_NAME TEST_INFRA_PERFORMANCE_TOOLS_CLUSTER_SCRIPTS; do
		if [ -z "${!var}" ] ; then
			echo "ERROR: $var is not set"
			discoverUnsetVar=true
		fi
	done
	if [ "${discoverUnsetVar}" = true ] ; then
		exit 1
	fi


	shout "Provision cluster: \"${CLUSTER_NAME}\""
	date
	
	if [ -z "$MACHINE_TYPE" ]; then
		export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
	fi
	if [ -z "${CLUSTER_VERSION}" ]; then
		export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
	fi
	env ADDITIONAL_LABELS="created-at=${CURRENT_TIMESTAMP}" "${TEST_INFRA_PERFORMANCE_TOOLS_CLUSTER_SCRIPTS}"/provision-gke-cluster.sh

}

createCluster