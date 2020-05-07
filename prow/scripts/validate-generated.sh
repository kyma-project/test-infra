#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"

# shellcheck disable=SC1090
# shellcheck disable=SC2086
source "${SCRIPT_DIR}/library.sh"

shout " - Running jobs generator tool..."

cd "${TEST_INFRA_SOURCES_DIR}/development/tools"
# TODO use rendertemplates binary instead of building one
go run cmd/rendertemplates/main.go --config "${TEST_INFRA_SOURCES_DIR}"/templates/config.yaml

shout " - Looking for job definition and rendered job files inconsistency..."

CHANGES=$(git status --porcelain)
if [[ -n "${CHANGES}" ]]; then
  echo "ERROR: Rendered job files does not match templates and the configuration:"
  echo "${CHANGES}"

  echo "
    Run:
        go run cmd/rendertemplates/main.go --config ${TEST_INFRA_SOURCES_DIR}/templates/config.yaml
    in the development/tools directory of the repository and commit changes.
    For more info read: /docs/prow/templates.md
    "
  exit 1
fi

shout " - Rendered job files are up to date"
