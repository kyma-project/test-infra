#!/usr/bin/env bash

set -o errexit

SCRIPTS_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../.."

# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPTS_PATH}/lib/log.sh"

discoverUnsetVar=false

for var in CLOUDSDK_DNS_ZONE_NAME DNS_FULL_NAME IP_ADDRESS IP_ADDRESS; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done

if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi


set +e
set -x

aws route53 change-resource-record-sets --hosted-zone-id "${CLOUDSDK_DNS_ZONE_NAME}" \
--change-batch '{ "Comment": "Deleting a record set","Changes": [ { "Action": "DELETE", "ResourceRecordSet": { "Name":"'"${DNS_FULL_NAME}"'", "Type": "A", "TTL":60, "ResourceRecords": [ { "Value": "'"${IP_ADDRESS}"'" } ] } } ] }'

set -e
set +x
