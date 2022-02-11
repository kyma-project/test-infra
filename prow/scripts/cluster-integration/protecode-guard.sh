#!/usr/bin/env bash

set -o errexit

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"


# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"

TMP_DIR=$(mktemp -d)

CLOUD_FUNCTION_URL="https://europe-west3-kyma-project.cloudfunctions.net/get-scan-result"
RESPONSE_FILE=${TMP_DIR}/response.json

function on_exit() {
    exit_status=$?
    if [ ${parse_status} != 0 ]; then
        echo "Last response: $RESULT"
    fi
}
trap on_exit exit

getScanResult(){
    local tag=$1
    curl -s "${CLOUD_FUNCTION_URL}?tag=${tag}" -H "Content-Type:application/json" > "${RESPONSE_FILE}"

    SUCCESS=$(jq '.success' "${RESPONSE_FILE}")
    if [[ "$SUCCESS" == "true" ]]; then
        echo "success"
    else
        echo "failure"
    fi
}

if [ -n "${PULL_NUMBER}" ]; then
  log::info "Execute Job Guard"
  "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

PR_NAME="PR-${PULL_NUMBER}"

echo "Protecode scan result for ${PR_NAME}:"

counter=1
limit=30
while [ $counter -le $limit ]
do
    log::banner "Attempt ${counter} of ${limit}"
    ((counter++))

    RESULT=$(getScanResult "${PR_NAME}")
    jq '.' "${RESPONSE_FILE}"

    if [[ "${RESULT}" == "success" ]]; then
        log::success "All images are green!"
        exit 0
    else
        log::warn "Some images contain security vulnerabilities"
        log::warn "For more details please check json output"

        # check if all images were already scanned
        images_in_queue=$(jq '.items | .[] | .scan | select(.status == "")' "$RESPONSE_FILE")
        if [[ -z "$images_in_queue" ]]; then
            # all images were scanned
            exit 1
        fi
    fi

    sleep 15
done

log::error "Timeout reached - job failed"
exit 1
