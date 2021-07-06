#!/usr/bin/env bash

#Description: Deletes a GKE cluster if exists along with DNS_RECORDS and STATIC IPs etc.
#
#Expected vars:
# - CLUSTER_NAME: name of the GKE cluster
# - TEST_INFRA_SOURCES_DIR: absolute path for test-infra/ directory
# - TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS: absolute path for test-infra/prow/scripts/cluster-integration/helpers directory
# - CLOUDSDK_COMPUTE_REGION: region where the GKE cluster is e.g. europe-west1-b
# - CLOUDSDK_DNS_ZONE_NAME: dns zone
#
#Permissions: In order to run this script you need to use a service account with "Compute Admin,DNS Administrator, Kubernetes Engine Admin, Kubernetes Engine Cluster Admin, Service Account User, Storage Admin" role

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/gcloud.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcloud.sh"
DNS_NAME="a.build.kyma-project.io."

function cleanup() {
	
	log::info "Running cleanup-cluster process"

	requiredVars=(
		CLUSTER_NAME
		TEST_INFRA_SOURCES_DIR
		TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS
		CLOUDSDK_COMPUTE_REGION
		CLOUDSDK_DNS_ZONE_NAME
		GCLOUD_NETWORK_NAME
		GCLOUD_SUBNET_NAME
	)

	utils::check_required_vars "${requiredVars[@]}"

	#Exporting variables used in subshells.
	export CLOUDSDK_DNS_ZONE_NAME
	export CLUSTER_NAME
	export CLOUDSDK_COMPUTE_REGION

	CLUSTER_EXISTS=$(gcloud container clusters list --filter="name~^${CLUSTER_NAME}" --format json | jq '.[].name' | tr -d '"' | wc -l)
	if [[ "$CLUSTER_EXISTS" -gt 0 ]]; then
		echo "Cleaning up: $CLUSTER_NAME"
		removeCluster
		echo "Cluster: $CLUSTER_NAME cleanup completed, moving to NET and DNS resources cleanup"
		removeResources
	else
		echo "Cluster: $CLUSTER_NAME not found, cleaning up NET and DNS resources"
		removeResources
	fi

}

function removeCluster() {
	#Turn off exit-on-error so that next step is executed even if previous one fails.
	set +e

	if [ -z "${SKIP_IMAGE_REMOVAL}" ] || [ "${SKIP_IMAGE_REMOVAL}" == "false" ]; then
		log::info "Fetching OLD_TIMESTAMP from cluster to be deleted"
		# Check if removing regionl cluster.
		if [ "${PROVISION_REGIONAL_CLUSTER}" ] && [ "${CLOUDSDK_COMPUTE_REGION}" ]; then
			#Pass gke region name instead zone name.
			readonly OLD_TIMESTAMP=$(gcloud container clusters describe "${CLUSTER_NAME}" --zone="${CLOUDSDK_COMPUTE_REGION}" --project="${GCLOUD_PROJECT_NAME}" --format=json | jq --raw-output '.resourceLabels."created-at-readable"')
		else
			readonly OLD_TIMESTAMP=$(gcloud container clusters describe "${CLUSTER_NAME}" --zone="${GCLOUD_COMPUTE_ZONE}" --project="${GCLOUD_PROJECT_NAME}" --format=json | jq --raw-output '.resourceLabels."created-at-readable"')
		fi
	fi

	log::info "Delete cluster $CLUSTER_NAME"
	gcloud::deprovision_gke_cluster "$CLUSTER_NAME"
	TMP_STATUS=$?
	if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

	MSG=""
	if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
	log::info "Cluster removal is finished: ${MSG}"

	if [ -z "${SKIP_IMAGE_REMOVAL}" ] || [ "${SKIP_IMAGE_REMOVAL}" == "false" ]; then
		log::info "Delete temporary Kyma-Installer Docker image with timestamp: ${OLD_TIMESTAMP}"

		echo "KYMA_INSTALLER_IMAGE=${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/${STANDARIZED_NAME}/${REPO_OWNER}/${REPO_NAME}:${OLD_TIMESTAMP}"

		gcloud::delete_docker_image "${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/${STANDARIZED_NAME}/${REPO_OWNER}/${REPO_NAME}:${OLD_TIMESTAMP}"
		TMP_STATUS=$?
		if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
	fi

	set -e
}

function removeResources() {
	set +e

	log::info "Delete Cluster DNS Record"
	GATEWAY_DNS_FULL_NAME="*.${CLUSTER_NAME}.${DNS_NAME}"
	# Get cluster IP address from DNS record.
	GATEWAY_IP_ADDRESS=$(gcloud dns record-sets list --zone "${CLOUDSDK_DNS_ZONE_NAME}" --name "${GATEWAY_DNS_FULL_NAME}" --format "value(rrdatas[0])")

	# Check if cluster IP was retrieved from DNS record. Remove cluster DNS record if IP address was retrieved.
	if [[ -n ${GATEWAY_IP_ADDRESS} ]]; then
		gcp::delete_dns_record \
			-a "$GATEWAY_IP_ADDRESS" \
			-h "*" \
			-s "$COMMON_NAME" \
			-p "$CLOUDSDK_CORE_PROJECT" \
			-z "$CLOUDSDK_DNS_ZONE_NAME"
	else
		echo "DNS entry for ${GATEWAY_DNS_FULL_NAME} not found"
	fi

	log::info "Delete Apiserver DNS Record"
	APISERVER_DNS_FULL_NAME="apiserver.${CLUSTER_NAME}.${DNS_NAME}"
	# Get apiserver IP address from DNS record.
	APISERVER_IP_ADDRESS=$(gcloud dns record-sets list --zone "${CLOUDSDK_DNS_ZONE_NAME}" --name "${APISERVER_DNS_FULL_NAME}" --format="value(rrdatas[0])")

	# Check if apiserver IP was retrieved from DNS record. Remove apiserver DNS record if IP address was retrieved.
	if [[ -n ${APISERVER_IP_ADDRESS} ]]; then
		gcp::delete_dns_record \
			-a "$APISERVER_IP_ADDRESS" \
			-h "apiserver" \
			-s "$COMMON_NAME" \
			-p "$CLOUDSDK_CORE_PROJECT" \
			-z "$CLOUDSDK_DNS_ZONE_NAME"
	else
		echo "DNS entry for ${APISERVER_DNS_FULL_NAME} not found"
	fi

	log::info "Release Cluster IP Address"
	GATEWAY_IP_ADDRESS_NAME=${CLUSTER_NAME}

	# Check if cluster IP address reservation exist.
	if [[ -n $(gcloud compute addresses list --filter="name=${CLUSTER_NAME}" --format "value(ADDRESS)") ]]; then
		# Get usage status of IP address reservation.
		GATEWAY_IP_STATUS=$(gcloud compute addresses describe "${CLUSTER_NAME}" --region "${CLOUDSDK_COMPUTE_REGION}" --format "value(status)")
		# Check if it's still in use. It shouldn't as we removed DNS records earlier.
		if [[ ${GATEWAY_IP_STATUS} == "IN_USE" ]]; then
			SECONDS=0
			END_TIME=$((SECONDS+600)) #600 seconds == 10 minutes
			echo "Waiting 600 seconds to unassigne cluster IP address."
			while [ ${SECONDS} -lt ${END_TIME} ];do
				sleep 10
				echo "Checking if cluster IP is unassigned."
				GATEWAY_IP_STATUS=$(gcloud compute addresses describe "${CLUSTER_NAME}" --region "${CLOUDSDK_COMPUTE_REGION}" --format "value(status)")
				if [[ ${GATEWAY_IP_STATUS} != "IN_USE" ]]; then
					echo "Cluster IP address sucessfully unassigned."
					break
				fi
			done
		fi
		if [[ ${GATEWAY_IP_STATUS} == "IN_USE" ]]; then
			echo "${GATEWAY_IP_ADDRESS_NAME} IP address has still status IN_USE. It should be unassigned earlier. Exiting"
			exit 1
		# Remove IP address reservation.
		else
			gcp::delete_ip_address \
    			-n "${GATEWAY_IP_ADDRESS_NAME}" \
				-p "$CLOUDSDK_CORE_PROJECT" \
				-R "$CLOUDSDK_COMPUTE_REGION"
			TMP_STATUS=$?
			if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
		fi
	else
		echo "${GATEWAY_IP_ADDRESS_NAME} IP address not found"
	fi

	MSG=""
	if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
	log::info "DNS, Gateway IP and Kyma installer image cleanup is finished: ${MSG}"

	set -e
}

cleanup
