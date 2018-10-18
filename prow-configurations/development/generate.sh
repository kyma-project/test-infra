#!/usr/bin/env bash

set -o errexit

# This script combines template file (plugins.yaml.tpl) with actual values provided as a JSON file (environment variable INPUT_JSON or read from command line)
# and generates output to plugins.yaml

if [ "$INPUT_JSON" == "" ]; then
    echo -n "Provide path to JSON file with values for plugins template: "
    read parametrizedJson
else
    parametrizedJson="$INPUT_JSON"
fi

if [ -z "$parametrizedJson" ]
then
    echo "ERROR: Path to JSON file is required"
    exit 1
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"


go run ${SCRIPT_DIR}/generator/main.go -template=${SCRIPT_DIR}/../plugins.yaml.tpl -out=${SCRIPT_DIR}/../plugins.yaml -input=${parametrizedJson}

echo "Content of generated file, plugins.yaml:"
cat ${SCRIPT_DIR}/../plugins.yaml