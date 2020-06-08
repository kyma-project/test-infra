#!/usr/bin/env bash

set -e
set -o pipefail

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} /path/to/plugins.yaml /path/to/config.yaml /path/to/jobs/dir"
    exit 1
}

readonly PLUGINS_PATH=$1
readonly CONFIG_PATH=$2
readonly JOBS_CONFIG_PATH=$3

if [[ -z "${PLUGINS_PATH}" ]] || [[ -z "${CONFIG_PATH}" ]] || [[ -z "${JOBS_CONFIG_PATH}" ]]; then
    usage
fi

echo "Checking plugin configuration from '${PLUGINS_PATH}' and prow configuration from '${CONFIG_PATH} and jobs configuration from '${JOBS_CONFIG_PATH}'"

/prow-tools/checkconfig --plugin-config="${PLUGINS_PATH}" --config-path="${CONFIG_PATH}" --job-config-path="${JOBS_CONFIG_PATH}"
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "OK"
fi

echo "Checking unique name of prow config jobs from '${JOBS_CONFIG_PATH}' directory"
/prow-tools/unique-jobs-name --config-path="${CONFIG_PATH}" --jobs-config-dir="${JOBS_CONFIG_PATH}"
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "OK"
fi
