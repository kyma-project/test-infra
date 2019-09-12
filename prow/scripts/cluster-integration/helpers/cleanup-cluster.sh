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

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
DNS_NAME="a.build.kyma-project.io."

function cleanup() {
	
	shout "Running cleanup-cluster process"
	discoverUnsetVar=false

	for var in CLUSTER_NAME TEST_INFRA_SOURCES_DIR TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS CLOUDSDK_COMPUTE_REGION CLOUDSDK_DNS_ZONE_NAME; do
		if [ -z "${!var}" ] ; then
			echo "ERROR: $var is not set"
			discoverUnsetVar=true
		fi
	done

	if [[ "${PERFORMACE_CLUSTER_SETUP}" == "" ]]; then
		for var in GCLOUD_NETWORK_NAME GCLOUD_SUBNET_NAME; do
			if [ -z "${!var}" ] ; then
				echo "ERROR: $var is not set"
				discoverUnsetVar=true
			fi
		done
	fi

	if [ "${discoverUnsetVar}" = true ] ; then
		exit 1
	fi

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

    shout "Fetching OLD_TIMESTAMP from cluster to be deleted"
	readonly OLD_TIMESTAMP=$(gcloud container clusters describe "${CLUSTER_NAME}" --zone="${GCLOUD_COMPUTE_ZONE}" --project="${GCLOUD_PROJECT_NAME}" --format=json | jq --raw-output '.resourceLabels."created-at-readable"')

	shout "Delete cluster $CLUSTER_NAME"
	"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/deprovision-gke-cluster.sh
	TMP_STATUS=$?
	if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

	MSG=""
	if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
	shout "Cluster removal is finished: ${MSG}"
	date

	shout "Delete temporary Kyma-Installer Docker image with timestamp: ${OLD_TIMESTAMP}"
	date

	echo "KYMA_INSTALLER_IMAGE=${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/${STANDARIZED_NAME}/${REPO_OWNER}/${REPO_NAME}:${OLD_TIMESTAMP}"

	KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/${STANDARIZED_NAME}/${REPO_OWNER}/${REPO_NAME}:${OLD_TIMESTAMP}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-image.sh
	TMP_STATUS=$?
	if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

	set -e
}

function removeResources() {
    if [[ "${PERFORMACE_CLUSTER_SETUP}" == "" ]]; then
		set +e

		shout "Delete Cluster DNS Record"
		date
		GATEWAY_DNS_FULL_NAME="*.${CLUSTER_NAME}.${DNS_NAME}"
		# Get cluster IP address from DNS record.
		GATEWAY_IP_ADDRESS=$(gcloud dns record-sets list --zone "${CLOUDSDK_DNS_ZONE_NAME}" --name "${GATEWAY_DNS_FULL_NAME}" --format "value(rrdatas[0])")

        # Check if cluster IP was retrieved from DNS record. Remove cluster DNS record if IP address was retrieved.
		if [[ -n ${GATEWAY_IP_ADDRESS} ]]; then
		    shout "running /delete-dns-record.sh --project=${GCLOUD_PROJECT_NAME} --zone=${CLOUDSDK_DNS_ZONE_NAME} --name=${GATEWAY_DNS_FULL_NAME} --address=${GATEWAY_IP_ADDRESS} --dryRun=false"

		    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${GCLOUD_PROJECT_NAME}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${GATEWAY_DNS_FULL_NAME}" --address="${GATEWAY_IP_ADDRESS}" --dryRun=false
		    TMP_STATUS=$?
		    if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
		else
		    echo "DNS entry for ${GATEWAY_DNS_FULL_NAME} not found"
		fi

        shout "Delete Apiserver DNS Record"
		date
		APISERVER_DNS_FULL_NAME="apiserver.${CLUSTER_NAME}.${DNS_NAME}"
		# Get apiserver IP address from DNS record.
        APISERVER_IP_ADDRESS=$(gcloud dns record-sets list --zone "${CLOUDSDK_DNS_ZONE_NAME}" --name "${APISERVER_DNS_FULL_NAME}" --format="value(rrdatas[0])")

        # Check if apiserver IP was retrieved from DNS record. Remove apiserver DNS record if IP address was retrieved.
        if [[ -n ${APISERVER_IP_ADDRESS} ]]; then
            shout "running /delete-dns-record.sh --project=${GCLOUD_PROJECT_NAME} --zone=${CLOUDSDK_DNS_ZONE_NAME} --name=${APISERVER_DNS_FULL_NAME} --address=${APISERVER_IP_ADDRESS} --dryRun=false"

            "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${APISERVER_DNS_FULL_NAME}" --address="${APISERVER_IP_ADDRESS}" --dryRun=false
            TMP_STATUS=$?
            if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
        else
            echo "DNS entry for ${APISERVER_DNS_FULL_NAME} not found"
        fi

		shout "Release Cluster IP Address"
		date
		GATEWAY_IP_ADDRESS_NAME=${CLUSTER_NAME}

        # Check if cluster IP address reservation exist.
		if [[ -n $(gcloud compute addresses list --filter="name=${CLUSTER_NAME}" --format "value(ADDRESS)") ]]; then
		    # Get usage status of IP address reservation.
		    GATEWAY_IP_STATUS=$(gcloud compute addresses describe "${CLUSTER_NAME}" --region "${CLOUDSDK_COMPUTE_REGION}" --format "value(status)")
		    # Check if it's still in use. It shouldn't as we removed DNS records earlier.
		    if [[ ${GATEWAY_IP_STATUS} == "IN_USE" ]]; then
                echo "${GATEWAY_IP_ADDRESS_NAME} IP address has still status IN_USE. It should be unassigned earlier. Exiting"
                exit 1
            # Remove IP address reservation.
		    else
		        shout "running /release-ip-address.sh --project=${GCLOUD_PROJECT_NAME} --ipname=${GATEWAY_IP_ADDRESS_NAME} --region=${CLOUDSDK_COMPUTE_REGION} --dryRun=false"

                "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/release-ip-address.sh --project="${GCLOUD_PROJECT_NAME}" --ipname="${GATEWAY_IP_ADDRESS_NAME}" --region="${CLOUDSDK_COMPUTE_REGION}" --dryRun=false
		        TMP_STATUS=$?
		        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
		    fi
		else
	        echo "${GATEWAY_IP_ADDRESS_NAME} IP address not found"
		fi
    fi

	MSG=""
	if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
	shout "DNS, Gateway IP and Kyma installer image cleanup is finished: ${MSG}"
	date

	set -e
}

cleanup