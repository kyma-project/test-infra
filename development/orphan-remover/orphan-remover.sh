#!/bin/bash

while [[ ORPHANS=$(gcloud compute instance-groups list --filter="INSTANCES=0" | tail -n +2 | wc -l) -gt 0 ]]; do
	echo "---> There are ${ORPHANS} to delete"

	RESOURCE_ID=$(gcloud compute instance-groups list --filter="INSTANCES=0" | tail -n 1 | awk '{print $1}' | cut -d '-' -f4)
	RESOURCE_ZONE=$(gcloud compute instance-groups list --filter="INSTANCES=0" | tail -n 1 | awk '{print $2}')
	RESOURCE_REGION=$(echo "${RESOURCE_ZONE}" | cut -d '-' -f1 -f2)

	RESOURCE_BACKEND_IDS=$(gcloud compute backend-services list --filter="name~${RESOURCE_ID}" | tail -n +2 | awk '{print $1}' | cut -d '-' -f3)

	echo "---> Deleting orphaned resources of ${RESOURCE_ID} in region: ${RESOURCE_REGION}/${RESOURCE_ZONE}"

	echo "---> Deleting forwarding-rules"
	gcloud compute forwarding-rules delete "k8s-fw-default-ing--${RESOURCE_ID}" --global -q
	echo "---> Deleting target-http-proxies"
	gcloud compute target-http-proxies delete "k8s-tp-default-ing--${RESOURCE_ID}" -q
	echo "---> Deleting url-maps"
	gcloud compute url-maps delete "k8s-um-default-ing--${RESOURCE_ID}" -q
	for ID in ${RESOURCE_BACKEND_IDS}; do
		echo "---> Deleting backend-services for ${ID}"
		gcloud compute backend-services delete "k8s-be-${ID}--${RESOURCE_ID}" --global -q
		echo "---> Deleting health-checks for ${ID}"
		gcloud compute health-checks delete "k8s-be-${ID}--${RESOURCE_ID}" -q
	done
	echo "---> Deleting instance-groups"
	gcloud compute instance-groups unmanaged delete "k8s-ig--${RESOURCE_ID}" --zone "${RESOURCE_ZONE}" -q
done