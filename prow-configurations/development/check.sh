#!/usr/bin/env bash

echo "Checking plugin configuration from 'prow-configurations/plugins.yaml' and prow configuration from 'prow-configurations/config.yaml'"

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
vendoredChecker="${SCRIPT_DIR}/checker/vendor/k8s.io/test-infra/prow/cmd/checkconfig/main.go"
if [ ! -f ${vendoredChecker} ]; then
    echo "Vendoring 'k8s.io/test-infra/prow/cmd/checkconfig'"
    cd "${SCRIPT_DIR}/checker"
    dep ensure -v -vendor-only

fi

go run ${vendoredChecker} --plugin-config=${SCRIPT_DIR}/../plugins.yaml --config-path=${SCRIPT_DIR}/../config.yaml
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "OK"
fi