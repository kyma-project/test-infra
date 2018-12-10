#!/bin/bash

# EU_ZONES=$(gcloud compute regions list --format json --filter="name~europe" | jq '.[].name' )

# for ZONE in ${EU_ZONES}; do
# 	echo "-> ZONE: ${ZONE}"

# done

IP_ADDRESSES=$(gcloud compute addresses list --filter="status=RESERVED" --filter="name~gkeint" --format json | jq '.[].address')
echo "---> IPS: ${RESERVED_IP_ADDRESSES}"
for IP in ${IP_ADDRESSES}; do
	
done

DNS_DOMAINS=$(gcloud dns record-sets list --zone build-kyma --filter="name~gkeint" --format json | jq '.[].name')