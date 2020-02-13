#!/usr/bin/env bash

# Expected vars:
# - SAP_SLACK_BOT_TOKEN - Token for Slack bot for which the vulnerabilities reports will be sent

set -e
set -o pipefail

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -z "$SAP_SLACK_BOT_TOKEN" ] ; then
    echo "ERROR: \$SAP_SLACK_BOT_TOKEN is not set"
    exit 1
fi

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ ! -d "${DEVELOPMENT_DIR}/tools/vendor" ]; then
    echo "Vendoring 'tools'"
    pushd "${DEVELOPMENT_DIR}/tools"
    dep ensure -v -vendor-only
    popd
fi

go run "${DEVELOPMENT_DIR}/tools/cmd/vulnerabilitycollector/main.go" "$@"
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "SUCCESS"
fi
