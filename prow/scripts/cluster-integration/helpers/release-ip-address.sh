#!/usr/bin/env bash

set -e
set -o pipefail

readonly DEVELOPMENT_DIR="$( cd "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/../../../../development" && pwd )"

readonly TOOL_DIR=ipcleaner
readonly OBJECT_NAME="Long lasting cluster IP cleaner"

"${DEVELOPMENT_DIR}"/resources-cleanup.sh "${TOOL_DIR}" "${OBJECT_NAME}" "$@"
