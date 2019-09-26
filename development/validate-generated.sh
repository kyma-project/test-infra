#!/usr/bin/env bash
#
# This is a helper script for validating if test-infra generators were executed and results were committed
#
set -o nounset
set -o errexit
set -o pipefail

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${CURRENT_DIR}/../prow/scripts/library.sh" || { echo 'Cannot load library utilities.'; exit 1; }

shout "- Running test-infra jobs generator..."
go run ${CURRENT_DIR}/tools/cmd/rendertemplates/main.go --config ${CURRENT_DIR}/../templates/config.yaml

shout "- Checking for modified files..."

# The porcelain format is used because it guarantees not to change in a backwards-incompatible
# way between Git versions or based on user configuration.
# source: https://git-scm.com/docs/git-status#_porcelain_format_version_1
if [[ -n "$(git status --porcelain)" ]]; then
    echo "Detected modified files:"
    git status --porcelain

    echo "
    Run:
        go run development/tools/cmd/rendertemplates/main.go --config templates/config.yaml
    in the root of the repository and commit changes.
    For more info read: /docs/prow/templates.md
    "
    exit 1
fi

shout "- No issues detected. Have a nice day :-)"
