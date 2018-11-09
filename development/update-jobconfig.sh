#!/usr/bin/env bash

DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} /path/to/job-config.yaml"
    exit 1
}

JOB_CONFIG_PATH=$1
if [[ -z "${JOB_CONFIG_PATH}" ]]; then
    usage
fi

kubectl create configmap job-config --from-file=$(basename ${JOB_CONFIG_PATH})=${JOB_CONFIG_PATH} --dry-run -o yaml | kubectl replace configmap job_config -f -
