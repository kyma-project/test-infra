#!/usr/bin/env bash

#Description: Adds new type "A" DNS entry for given subdomain and IP Address
#
#Expected vars:
# - CLOUDSDK_CORE_PROJECT: name of a GCP project where new DNS record is created.
# - CLOUDSDK_DNS_ZONE_NAME: Name of an existing DNS zone in the project (NOT its DNS name!)
# - CLOUDSDK_COMPUTE_REGION: Region for the IP Address (e.g. europe-west3)
# - DNS_FULL_NAME: DNS name
# - IP_ADDRESS_NAME: Name for the IP Address object (NOT an actual IP Address)
# - IP_ADDRESS: v4 IP Address for the DNS record.
#
#Permissions: In order to run this script you need to use a service account with "DNS Administrator" role
#Permissions: In order to run this script you need to use a service account with "Compute Network Admin" role

export IP_ADDRESS
export IP_ADDRESS_NAME
export DNS_FULL_NAME
export CLOUDSDK_CORE_PROJECT="sap-kyma-prow-workloads"
export CLOUDSDK_DNS_ZONE_NAME="build-kyma-workloads"
export CLOUDSDK_COMPUTE_REGION="europe-west3"
DNS_FULL_NAME="dns-test-$(date +%s).a.build.kyma-project.io."
IP_ADDRESS_NAME="dns-test-$(date +%s)"

SCRIPTS_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=prow/scripts/lib/gcloud.sh
source "${SCRIPTS_PATH}/lib/gcloud.sh"
#shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPTS_PATH}/lib/log.sh"


gcloud::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"
IP_ADDRESS=$(gcloud::reserve_ip_address "${IP_ADDRESS_NAME}")
gcloud::create_dns_record "${IP_ADDRESS}" "${DNS_FULL_NAME}"
log::banner "Debug output"
cat "${ARTIFACTS}/dns-debug.txt"
gcloud::delete_dns_record "${IP_ADDRESS}" "${DNS_FULL_NAME}"
gcloud::delete_ip_address "${IP_ADDRESS_NAME}"
