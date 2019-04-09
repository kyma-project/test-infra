#!/usr/bin/env bash

set -e
set -o pipefail

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

readonly TOOL_DIR=firewallcleaner
readonly OBJECT_NAME="Firewall rules"

"${DEVELOPMENT_DIR}"/resources-cleanup.sh "${TOOL_DIR}" "${OBJECT_NAME}" "$@"

