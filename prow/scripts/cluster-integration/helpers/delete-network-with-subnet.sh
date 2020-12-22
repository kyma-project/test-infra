#!/usr/bin/env bash

# Description: Deletes the GCP network for cluster

# Expected vars:
# - GCLOUD_NETWORK_NAME - name for the new GCP network
# - GCLOUD_SUBNET_NAME - name for the subnet of the network
# - GCLOUD_PROJECT_NAME - name of GCP project

set -o errexit

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

requiredVars=(
    GCLOUD_NETWORK_NAME
    GCLOUD_SUBNET_NAME
    GCLOUD_PROJECT_NAME
)

utils::checkRequiredVars "${requiredVars[@]}"

gcloud compute networks subnets delete "${GCLOUD_SUBNET_NAME}" \
 --quiet

gcloud compute networks delete "${GCLOUD_NETWORK_NAME}" \
 --project="${GCLOUD_PROJECT_NAME}" \
 --quiet
