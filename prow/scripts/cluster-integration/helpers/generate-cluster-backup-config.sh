#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

discoverUnsetVar=false

for var in KYMA_RESOURCES_DIR TEST_INFRA_SOURCES_DIR BACKUP_RESTORE_BUCKET BACKUP_CREDENTIALS KYMA_SCRIPTS_DIR; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/library.sh

shout "Create a Secret for Velero"

CLOUD_PROVIDER="gcp"

VELERO_SECRET_TPL_PATH="${KYMA_RESOURCES_DIR}/backup-secret-${CLOUD_PROVIDER}.yaml.tpl"
VELERO_SECRET_OUTPUT_PATH=$(mktemp)
cp "${VELERO_SECRET_TPL_PATH}" "${VELERO_SECRET_OUTPUT_PATH}"


BASE64_CLOUD_PROVIDER=$(echo -n "${CLOUD_PROVIDER}" | base64 -w0)
BASE64_BUCKET=$(echo -n "${BACKUP_RESTORE_BUCKET}" | base64 -w0)
BASE64_CLOUD_CREDENTIALS_FILE_CONTENT_BASE64=$(base64 -w0 "${BACKUP_CREDENTIALS}")


bash "${KYMA_SCRIPTS_DIR}"/replace-placeholder.sh --path "${VELERO_SECRET_OUTPUT_PATH}" --placeholder "__CLOUD_PROVIDER__" --value "${BASE64_CLOUD_PROVIDER}"

bash "${KYMA_SCRIPTS_DIR}"/replace-placeholder.sh --path "${VELERO_SECRET_OUTPUT_PATH}" --placeholder "__BSL_BUCKET__" --value "${BASE64_BUCKET}"

bash "${KYMA_SCRIPTS_DIR}"/replace-placeholder.sh --path "${VELERO_SECRET_OUTPUT_PATH}" --placeholder "__CLOUD_CREDENTIALS_FILE_CONTENT_BASE64__" --value "${BASE64_CLOUD_CREDENTIALS_FILE_CONTENT_BASE64}"

echo -e "\nApplying secret for Velero"

sed -e "s/__.*__//g" "${VELERO_SECRET_OUTPUT_PATH}" | kubectl apply -f-

rm "${VELERO_SECRET_OUTPUT_PATH}"