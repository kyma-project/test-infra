#!/usr/bin/env bash

DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} /path/to/plugins.yaml /path/to/config.yaml"
    exit 1
}

PLUGINS_PATH=$1
CONFIG_PATH=$2

if [[ -z "${PLUGINS_PATH}" ]] || [[ -z "${CONFIG_PATH}" ]]; then
    usage
fi

echo "Checking plugin configuration from '${PLUGINS_PATH}' and prow configuration from '${CONFIG_PATH}'"

vendoredChecker="${DEVELOPMENT_DIR}/checker/vendor/k8s.io/test-infra/prow/cmd/checkconfig/main.go"
if [ ! -f ${vendoredChecker} ]; then
    echo "Vendoring 'k8s.io/test-infra/prow/cmd/checkconfig'"
    cd "${DEVELOPMENT_DIR}/checker"
    dep ensure -v -vendor-only
fi

go run ${vendoredChecker} --plugin-config=${PLUGINS_PATH} --config-path=${CONFIG_PATH}
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "OK"
fi
