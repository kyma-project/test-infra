#!/usr/bin/env bash

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} /path/to/plugins.yaml"
    exit 1
}

PLUGINS_PATH=$1
if [[ -z "${PLUGINS_PATH}" ]]; then
    usage
fi

kubectl create configmap plugins --from-file=plugins.yaml="${PLUGINS_PATH}" --dry-run -o yaml | kubectl replace configmap plugins -f -