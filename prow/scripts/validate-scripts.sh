#!/usr/bin/env bash
#
# This is a helper script for validating bash scripts inside the test-infra repository.
# It uses shellcheck as a validator.
set -e
set -o pipefail

export LC_ALL=C.UTF-8
find "./development/" -type f -name "*.sh" -exec "shellcheck" -x {} +
find "./prow" -type f -name "*.sh" -exec "shellcheck" -x {} +

echo "No issues detected!"
