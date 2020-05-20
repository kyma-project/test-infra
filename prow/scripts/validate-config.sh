#!/usr/bin/env bash

set -e
set -o pipefail

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} /path/to/plugins.yaml /path/to/config.yaml /path/to/jobs/dir"
    exit 1
}

readonly BASE_DIR=$(pwd)
readonly PLUGINS_PATH=$1
readonly CONFIG_PATH=$2
readonly JOBS_CONFIG_PATH=$3

if [[ -z "${PLUGINS_PATH}" ]] || [[ -z "${CONFIG_PATH}" ]] || [[ -z "${JOBS_CONFIG_PATH}" ]]; then
    usage
fi

echo "Checking plugin configuration from '${PLUGINS_PATH}' and prow configuration from '${CONFIG_PATH} and jobs configuration from '${JOBS_CONFIG_PATH}'"

cd "development/checker"
go get k8s.io/test-infra/prow/cmd/checkconfig@v0.0.0-20200320172837-fbc86f22b087
"${GOPATH}/bin/checkconfig" --plugin-config="${BASE_DIR}/${PLUGINS_PATH}" --config-path="${BASE_DIR}/${CONFIG_PATH}" --job-config-path="${BASE_DIR}/${JOBS_CONFIG_PATH}"
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "OK"
fi

echo "Checking unique name of prow config jobs from '${JOBS_CONFIG_PATH}' directory"
go run "unique-jobs-name/main.go" --config-path="${BASE_DIR}/${CONFIG_PATH}" --jobs-config-dir="${BASE_DIR}/${JOBS_CONFIG_PATH}"
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "OK"
fi
