#!/usr/bin/env bash

#Description: Deletes a GKE cluster if exists along with DNS_RECORDS and STATIC IPs etc.
#
#Expected vars:
# - CLUSTER_NAME: name of the GKE cluster
# - TEST_INFRA_SOURCES_DIR: absolute path for test-infra/ directory 
# - TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS: absolute path for test-infra/prow/scripts/cluster-integration/helpers directory 
# - CLOUDSDK_COMPUTE_REGION: region where the GKE cluster is e.g. europe-west1-b
#
#Permissions: In order to run this script you need to use a service account with "Compute Admin,DNS Administrator, Kubernetes Engine Admin, Kubernetes Engine Cluster Admin, Service Account User, Storage Admin" role

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
function cleanup() {
	discoverUnsetVar=false

	for var in CLUSTER_NAME TEST_INFRA_SOURCES_DIR TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS CLOUDSDK_COMPUTE_REGION; do
		if [ -z "${!var}" ] ; then
			echo "ERROR: $var is not set"
			discoverUnsetVar=true
		fi
	done
	if [ "${discoverUnsetVar}" = true ] ; then
		exit 1
	fi
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

cleanup