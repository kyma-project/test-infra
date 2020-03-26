#!/usr/bin/env bash

set -e
set -o pipefail

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

readonly TOOL_DIR=dnscollector
readonly OBJECT_NAME="DNSs and IPs"

"${DEVELOPMENT_DIR}"/resources-cleanup-prebuilt.sh "${TOOL_DIR}" "${OBJECT_NAME}" "$@"
