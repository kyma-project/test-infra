#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

kubectl create configmap plugins --from-file=plugins.yaml=${SCRIPT_DIR}/plugins.yaml --dry-run -o yaml | kubectl replace configmap plugins -f -