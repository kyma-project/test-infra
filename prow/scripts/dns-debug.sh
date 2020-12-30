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

export CLOUDSDK_CORE_PROJECT="sap-kyma-prow-workloads"
export CLOUDSDK_DNS_ZONE_NAME="build-kyma-workloads"
export CLOUDSDK_COMPUTE_REGION="europe-west3"
export DNS_FULL_NAME="dns-test.a.build.kyma-project.io."
export IP_ADDRESS_NAME="dns-test"
export IP_ADDRESS

SCRIPTS_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=prow/scripts/lib/gcloud.sh
source "${SCRIPTS_PATH}/lib/gcloud.sh"
#shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPTS_PATH}/lib/log.sh"


gcloud::authenticate
IP_ADDRESS=$("${SCRIPTS_PATH}"/cluster-integration/helpers/reserve-ip-address.sh)
"${SCRIPTS_PATH}"/cluster-integration/helpers/create-dns-record.sh
"${SCRIPTS_PATH}"/cluster-integration/helpers/delete-dns-record-gcloud.sh
"${SCRIPTS_PATH}"/cluster-integration/helpers/release-ip-address.sh --dryRun false --ipname ${IP_ADDRESS_NAME} --project ${CLOUDSDK_CORE_PROJECT} --region ${CLOUDSDK_COMPUTE_REGION}
