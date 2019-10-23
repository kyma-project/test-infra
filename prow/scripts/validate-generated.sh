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

cd "${TEST_INFRA_SOURCES_DIR}/development/tools/"
make resolve

cd "${TEST_INFRA_SOURCES_DIR}"
go run development/tools/cmd/rendertemplates/main.go --config templates/config.yaml

shout " - Looking for job definition and rendered job files inconsistency..."

CHANGES=$(git status --porcelain)
if [[ -n "${CHANGES}" ]]; then
  echo "ERROR: Rendered job files does not match templates and the configuration:"
  echo "${CHANGES}"

  echo "
    Run:
        go run development/tools/cmd/rendertemplates/main.go --config templates/config.yaml
    in the root of the repository and commit changes.
    For more info read: /docs/prow/templates.md
    "
  exit 1
fi

shout " - Rendered job files are up to date"
