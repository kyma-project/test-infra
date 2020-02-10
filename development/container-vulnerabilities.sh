#!/usr/bin/env bash

# Expected vars:
# - KYMA_PROJECT_DIR - directory of kyma-project sources
# - SAP_SLACK_BOT_TOKEN - Token for Slack bot for which the vulnerabilities reports will be sent

set -e
set -o pipefail

discoverUnsetVar=false

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

for var in SAP_SLACK_BOT_TOKEN; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

go run "${DEVELOPMENT_DIR}/tools/cmd/vulnerabilitycollector/main.go" "$@"
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "SUCCESS"
fi
