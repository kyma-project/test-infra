#!/usr/bin/env bash

set -e
set -o pipefail

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

readonly TOOL_DIR=orphanremover
readonly OBJECT_NAME="LoadBalancer resources"

"${DEVELOPMENT_DIR}"/resources-cleanup.sh "${TOOL_DIR}" "${OBJECT_NAME}" "$@"

