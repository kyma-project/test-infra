#!/usr/bin/env bash

set -e
set -o pipefail

readonly DEVELOPMENT_DIR="$( cd "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/../../../../development" && pwd )"

readonly TOOL_DIR=dnscleaner
readonly OBJECT_NAME="Long lasting cluster DNS cleaner"

"${DEVELOPMENT_DIR}"/resources-cleanup.sh "${TOOL_DIR}" "${OBJECT_NAME}" --attempts 10 "$@"
echo "DNS Record deleted, but it can be visible for some time due to DNS caches"
