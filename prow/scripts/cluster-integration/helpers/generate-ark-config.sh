#!/usr/bin/env bash


shout "Create a Secret for Ark"

export SA_NAME=sa-gcs-plank
export SA_DISPLAY_NAME=sa-gcs-plank
export SECRET_FILE=sa-gcs-plank
export ROLE=roles/storage.objectAdmin

gcloud iam service-accounts create $SA_NAME --display-name $SA_DISPLAY_NAME

ARK_SECRET_TPL_PATH="${KYMA_RESOURCES_DIR}/ark-secret.yaml.tpl"
ARK_SECRET_OUTPUT_PATH=$(mktemp)
cp "${ARK_SECRET_TPL_PATH}" "${ARK_SECRET_OUTPUT_PATH}"

CLOUD_PROVIDER="gcp"

BASE64_CLOUD_PROVIDER=$(echo -n "${CLOUD_PROVIDER}" | base64 | tr -d '\n')
BASE64_BUCKET=$(echo -n "${KYMA_ARK_BUCKET}" | base64 | tr -d '\n')

gcloud iam service-accounts keys create $SECRET_FILE --iam-account=$SA_NAME@$GCLOUD_PROJECT_NAME.iam.gserviceaccount.com

BASE64_CLOUD_CREDENTIALS_FILE_CONTENT_BASE64=$(gcloud iam service-accounts keys list --iam-account=$SA_NAME@$GCLOUD_PROJECT_NAME.iam.gserviceaccount.com --limit=1 --format='csv[no-heading](KEY_ID)')


bash "${KYMA_SCRIPTS_DIR}"/replace-placeholder.sh --path "${ARK_SECRET_OUTPUT_PATH}" --placeholder "__CLOUD_PROVIDER__" --value "${BASE64_CLOUD_PROVIDER}"

bash "${KYMA_SCRIPTS_DIR}"/replace-placeholder.sh --path "${ARK_SECRET_OUTPUT_PATH}" --placeholder "__BSL_BUCKET__" --value "${BASE64_BUCKET}"

bash "${KYMA_SCRIPTS_DIR}"/replace-placeholder.sh --path "${ARK_SECRET_OUTPUT_PATH}" --placeholder "__CLOUD_CREDENTIALS_FILE_CONTENT_BASE64__" --value "${BASE64_CLOUD_CREDENTIALS_FILE_CONTENT_BASE64}"

echo -e "\nApplying secret for Ark"
kubectl apply -f "${ARK_SECRET_OUTPUT_PATH}"

rm "${ARK_SECRET_OUTPUT_PATH}"