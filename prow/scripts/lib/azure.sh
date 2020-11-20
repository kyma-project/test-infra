#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}"/log.sh

function az::verify_binary {
  if ! [[ -x $(command -v az) ]]; then
    log::error "'az' command not found in \$PATH. Exiting..."
    exit 1
  else
    az version
  fi
}

# az::login logs in to the azure service using provided credentials file in the function argument.
# Function accepts JSON file formatted below:
# {
#   "tenant": "tenant_id",
#   "app_id": "subscription_app_id",
#   "secret": "subscription_secret"
# }
function az::login {
  local AZURE_SUBSCRIPTION_TENANT
  local AZURE_SUBSCRIPTION_APP_ID
  local AZURE_SUBSCRIPTION_SECRET

  # Check the provided credentials in the argument.
  # Use arguments to avoid exporting sensitive values.
  if [[ -z "$1" ]]; then
    log::error "Azure credentials file not provided. please provide azure credentials filepath in the argument. Exiting..."
    exit 1
  elif ! [[ -f "$1" ]]; then
    log:error "Azure credentials file not found. Make sure it is present under the provided filepath. Exiting..."
    exit 1
  fi
  AZURE_CREDENTIALS_FILE="$1"
  read -r AZURE_SUBSCRIPTION_TENANT AZURE_SUBSCRIPTION_APP_ID AZURE_SUBSCRIPTION_SECRET < <(jq -r '.tenant, .app_id, .secret' "$AZURE_CREDENTIALS_FILE")

  # login
  az login --service-principal -u "${AZURE_SUBSCRIPTION_APP_ID}" -p "${AZURE_SUBSCRIPTION_SECRET}" --tenant "${AZURE_SUBSCRIPTION_TENANT}"
}

# This check will trigger everytime the file is sourced.
# This should allow easy checking for the related requirements without copying loads of code.
# If more checks will be needed we should add another function there.
az::verify_binary
