#!/usr/bin/env bash
#
# This script generates plugins.yaml and config.yaml from template files (plugins.yaml.tpl and config.yaml.tpl respectively)
# by applying actual values from file defined by $INPUT_JSON or provided via command line.

set -o errexit
readonly INPUT_JSON

if [ "$INPUT_JSON" == "" ]; then
    echo -n "Provide path to JSON file with actual values for templates: "
    read  parametrizedJson
else
    parametrizedJson="$INPUT_JSON"
fi

if [ -z "$parametrizedJson" ]
then
    echo "ERROR: Path to JSON file is required"
    exit 1
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"


go run "${SCRIPT_DIR}/generator/main.go" -template="${SCRIPT_DIR}/../plugins.yaml.tpl" -out="${SCRIPT_DIR}/../plugins.yaml" -input="${parametrizedJson}"

echo "Content of generated file, plugins.yaml:"
cat "${SCRIPT_DIR}/../plugins.yaml"

echo

go run "${SCRIPT_DIR}/generator/main.go" -template="${SCRIPT_DIR}/../config.yaml.tpl" -out="${SCRIPT_DIR}/../config.yaml" -input="${parametrizedJson}"
echo "Content of generated file, config.yaml:"
cat "${SCRIPT_DIR}/../config.yaml"
