#!/usr/bin/env bash

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} /path/to/plugins.yaml"
    exit 1
}

readonly PLUGINS_PATH=$1
if [[ -z "${PLUGINS_PATH}" ]]; then
    usage
fi

readonly UPLOADER="cmd/configuploader/main.go"
readonly CONFIG="${HOME}/.kube/config"

cd "${DEVELOPMENT_DIR}/tools" || exit 1
go run "${UPLOADER}" --kubeconfig "${CONFIG}" --plugins-config-path "${PLUGINS_PATH}"