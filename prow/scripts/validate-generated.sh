#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"

log::info "Running jobs generator tool..."

cd "${TEST_INFRA_SOURCES_DIR}/development/tools"
# TODO use rendertemplates binary instead of building one
go run cmd/rendertemplates/main.go --config "${TEST_INFRA_SOURCES_DIR}"/templates/config.yaml

log::info "Looking for job definition and rendered job files inconsistency..."

CHANGES=$(git status --porcelain)
if [[ -n "${CHANGES}" ]]; then
  log::error "Rendered job files does not match templates and the configuration:"
  log::info "${CHANGES}"

  echo "
    Run:
        go run cmd/rendertemplates/main.go --config ../../templates/config.yaml
    in the development/tools directory of the repository and commit changes.
    For more info read: /docs/prow/templates.md
    "

  git diff
  exit 1
fi

log::info "Rendered job files are up to date"
