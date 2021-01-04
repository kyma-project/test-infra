#!/usr/bin/env bash

#Description: Deletes docker image from registry
#
#Expected vars:
# - KYMA_INSTALLER_IMAGE: Image name with tag
# - CLOUDSDK_CORE_PROJECT: GCloud Project name, used for KYMA_INSTALLER_IMAGE validation
# - GOOGLE_APPLICATION_CREDENTIALS
# - TEST_INFRA_SOURCES_DIR
#
#Permissions: In order to run this script you need to use a service account with "Storage Admin" role

set -o errexit

# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"

requiredVars=(
    KYMA_INSTALLER_IMAGE
    CLOUDSDK_CORE_PROJECT
    GOOGLE_APPLICATION_CREDENTIALS
    TEST_INFRA_SOURCES_DIR
)

utils::check_required_vars "${requiredVars[@]}"

function cleanup() {
    # activate defauklt SA
    client_email=$(jq -r '.client_email' < "${GOOGLE_APPLICATION_CREDENTIALS}")
    log::info "Activating account $client_email"
    gcloud config set account "${client_email}" || exit 1
}

log::info "Authenticate as service account with write access to GCR"
trap cleanup EXIT
# authenticate Sa Gcr
authKey=$1
if [[ -n "${authKey}" ]]; then
    client_email=$(jq -r '.client_email' < "${authKey}")
    log::info "Authenticating in regsitry ${DOCKER_PUSH_REPOSITORY%%/*} as $client_email"
    docker login -u _json_key --password-stdin https://"${DOCKER_PUSH_REPOSITORY%%/*}" < "${authKey}" || exit 1
else
    log::info "could not authenticate to Docker Registry: authKey is empty" >&2
fi

gcloud container images delete "${KYMA_INSTALLER_IMAGE}"

