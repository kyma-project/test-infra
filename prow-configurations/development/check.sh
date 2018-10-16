#!/usr/bin/env bash

echo "Checking plugin configuration from 'prow-configurations/plugins.yaml' and prow configuration from 'prow-configurations/config.yaml'"

go run ./checker/vendor/k8s.io/test-infra/prow/cmd/checkconfig/main.go --plugin-config=./../plugins.yaml --config-path=./../config.yaml
status=$?

if [ ${status} -ne 0 ]
then
    echo "ERROR"
    exit 1
else
    echo "OK"
fi