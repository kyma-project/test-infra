#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

echo "Running jobs generator tool..."

go run cmd/tools/rendertemplates/main.go --config templates/config.yaml --templates templates/templates --data templates/data

echo "Looking for job definition and rendered job files inconsistency..."

CHANGES=$(git status --porcelain)
if [[ -n "${CHANGES}" ]]; then
  echo "Rendered job files does not match templates and the configuration:"
  echo "${CHANGES}"

  echo "
    Run:
        go run development/tools/cmd/rendertemplates/main.go --config templates/config.yaml --templates templates/templates --data templates/data
    in the root directory of the test-infra repository and commit changes.
    For more info read: /docs/prow/templates.md
    "
  exit 1
fi

echo "Rendered job files are up to date"
