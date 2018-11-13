#!/usr/bin/env bash

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} /path/to/config.yaml"
    exit 1
}

CONFIG_PATH=$1
if [[ -z "${CONFIG_PATH}" ]]; then
    usage
fi

kubectl create configmap config --from-file=config.yaml="${CONFIG_PATH}" --dry-run -o yaml | kubectl replace configmap config -f -