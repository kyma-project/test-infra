#!/usr/bin/env bash

set -o errexit

#shellcheck source=prow/scripts/lib/azure.sh
source "prow/scripts/lib/azure.sh"
#shellcheck source=prow/scripts/lib/log.sh
source "prow/scripts/lib/log.sh"
# shellcheck disable=SC1090
source "prow/scripts/lib/utils.sh"

log::info "Validating environment"

requiredVars=(
    AZURE_RS_GROUP
    AZURE_SUBSCRIPTION_ID
    AZURE_CREDENTIALS_FILE
)

utils::checkRequiredVars ${requiredVars[@]}

az::login "$AZURE_CREDENTIALS_FILE"
az::set_subscription "$AZURE_SUBSCRIPTION_ID"

log::info "Removing orphaned Storage Accounts from ${AZURE_RS_GROUP} Resource Group"
while read -r account; do
    if [[ -z "${account}" ]]; then
        # Nothing to delete
        continue
    fi

    echo "....Removing ${account}"
    az storage account delete \
        --name "${account}" \
        --resource-group "${AZURE_RS_GROUP}" \
        --yes
done <<< "$(az storage account list \
    --query "[?tags.\"created-by\"=='prow' && sum([to_number(tags.\"created-at\"),to_number(tags.\"ttl\")]) < to_number('$(date +%s)')].name" \
    --resource-group "${AZURE_RS_GROUP}" \
    --output tsv)"

log::success "Finished"
