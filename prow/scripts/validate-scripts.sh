#!/usr/bin/env bash
#
# This is a helper script for validating bash scripts inside the test-infra repository.
# It uses shellcheck as a validator.
set -e
set -o pipefail

export LC_ALL=C.UTF-8

# SHELLCHECK_OPTS="-e SC2204 -e SC2205 -e SC2200 -e SC2201 -e SC2198 -e SC2199 -e SC2196 -e SC2197 -e SC2195 -e SC2194 -e SC2193 -e SC2188 -e SC2189 -e SC2186 -e SC1109 -e SC1108 -e SC2185 -e SC2184 -e SC2183 -e SC2182 -e SC2181 -e SC1106"

echo "$SHELLCHECK_OPTS"

find "./development/" -type f -name "*.sh" -exec "shellcheck" -x {} +
find "./prow" -type f -name "*.sh" -exec "shellcheck" -x {} +

echo "No issues detected!"
