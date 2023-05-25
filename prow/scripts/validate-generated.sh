#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"

log::info "Running jobs generator tool..."

cd "${TEST_INFRA_SOURCES_DIR}"
# TODO use rendertemplates binary instead of building one
DUPLICATES=$(go run development/tools/cmd/rendertemplates/main.go --config "${TEST_INFRA_SOURCES_DIR}"/templates/config.yaml --templates "${TEST_INFRA_SOURCES_DIR}"/templates/templates --data "${TEST_INFRA_SOURCES_DIR}"/templates/data)

log::info "Looking for duplicate target files..."

if [[ -n "${DUPLICATES}" ]]; then
  log::error "Rendered jobs has duplicated target files:"
  log::info "${DUPLICATES}"

  exit 1
fi

log::info "Looking for job definition and rendered job files inconsistency..."

CHANGES=$(git status --porcelain)
if [[ -n "${CHANGES}" ]]; then
  log::error "Rendered job files does not match templates and the configuration:"
  log::info "${CHANGES}"

  echo "
    Run:
        go run development/tools/cmd/rendertemplates/main.go --config templates/config.yaml --templates templates/templates --data templates/data
    in the root directory of the test-infra repository and commit changes.
    For more info read: /docs/prow/templates.md
    "
  exit 1
fi

log::info "Rendered job files are up to date"
