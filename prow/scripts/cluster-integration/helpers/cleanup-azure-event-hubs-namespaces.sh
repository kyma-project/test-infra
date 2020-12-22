#!/usr/bin/env bash
#########################################################################################################
# Cleanup Azure Eventhubs Namespaces if exceeded the provided TTL in hours.
#
# Expected vars:
#
# - AZURE_CREDENTIALS_FILE    - Azure credentials JSON file required for authentication.
# - AZURE_SUBSCRIPTION_ID     - Azure Subscription ID required to set the current active Subscription.
# - AZURE_SUBSCRIPTION_NAME   - Azure Subscription Name with the Eventhubs Namespaces to be cleaned-up.
# - TTL_HOURS                 - Time to live in hours before Azure Eventhubs Namespace is cleaned-up.
#########################################################################################################

set -o errexit
set -o pipefail

#########################################################################################################
# Global Variables
#########################################################################################################
VARIABLES=(
  AZURE_CREDENTIALS_FILE
  AZURE_SUBSCRIPTION_ID
  AZURE_SUBSCRIPTION_NAME
  TTL_HOURS
)

#########################################################################################################
# Constants
#########################################################################################################
readonly NOW="$(date +%s)"
readonly SECONDS_PER_HOUR=3600
readonly TEST_INFRA_PROJECT_DIR=/home/prow/go/src/github.com/kyma-project/test-infra

#########################################################################################################
# Source common library
#########################################################################################################
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_PROJECT_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/azure.sh
source "${TEST_INFRA_PROJECT_DIR}/prow/scripts/lib/azure.sh"
# shellcheck source=prow/scripts/lib/azure.sh
source "${TEST_INFRA_PROJECT_DIR}/prow/scripts/lib/utils.sh"

#########################################################################################################
# Ensure that all expected vars are set before running the script.
#
# GLOBALS:
#   VARIABLES
# ARGUMENTS:
#   None
# OUTPUTS:
#   Print vars which are not set.
# RETURN:
#   0 if success, non-zero on error.
#########################################################################################################
function ensure_vars_or_die() {
  log::banner "Ensure Env Vars"

  utils::check_required_vars "${VARIABLES[@]}"
}

#########################################################################################################
# Cleanup a single Azure Eventhubs Namespace if exceeded the maximum TTL in hours.
#
# GLOBALS:
#   NOW
#   TTL_HOURS
#   SECONDS_PER_HOUR
# ARGUMENTS:
#   A JSON String that represents a single Azure Eventhubs Namespace.
#   It should look like '{"createdAt":"any", "name":"any", "resourceGroup":"any"}'.
# OUTPUTS:
#   Print the result of deleting the Azure Eventhubs Namespace.
# RETURN:
#   0 if success, non-zero on error.
#########################################################################################################
function cleanup_eventhubs_namespace() {
  local name
  local created_at
  local elapsed_hours

  name=$(echo "${1}" | jq -r ".name")
  created_at=$(date -d "$(echo "${1}" | jq -r ".createdAt")" +%s)
  resourceGroup=$(echo "${1}" | jq -r ".resourceGroup")
  elapsed_hours=$(((NOW - created_at) / SECONDS_PER_HOUR))

  # delete Eventhubs Namespace if it is older than the TTL in hours
  if [[ ${elapsed_hours} -ge ${TTL_HOURS} ]]; then
    log::info "Delete Namespace [${name}] in ResourceGroup [${resourceGroup}] (${elapsed_hours}h old)"
    az eventhubs namespace delete -n "${name}" -g "${resourceGroup}"
  fi
}

#########################################################################################################
# Cleanup Azure Eventhubs Namespaces if exceeded the maximum TTL in hours.
#
# GLOBALS:
#   AZURE_SUBSCRIPTION_NAME
# ARGUMENTS:
#   None
# RETURN:
#   0 if success, non-zero on error.
#########################################################################################################
function cleanup() {
  log::banner "Cleanup Azure Eventhubs Namespaces older than ${TTL_HOURS}h"

  for ns in $(az eventhubs namespace list --subscription "${AZURE_SUBSCRIPTION_NAME}" --query '[].{createdAt:createdAt, name:name, resourceGroup:resourceGroup}' --output json | jq -c '.[]'); do
    cleanup_eventhubs_namespace "${ns}"
  done
}

#########################################################################################################
# Run the script steps in order.
#
# GLOBALS:
#   None
# ARGUMENTS:
#   None
# RETURN:
#   0 if success, non-zero on error.
#########################################################################################################
function main() {
  ensure_vars_or_die
  az::login "$AZURE_CREDENTIALS_FILE"
  az::set_subscription "$AZURE_SUBSCRIPTION_ID"
  cleanup
  log::success "Cleanup Azure Eventhubs Namespaces finished successfully"
}

main
