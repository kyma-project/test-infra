#!/usr/bin/env bash

set -e

readonly RESOURCES_PATH="/Users/i355395/go/src/github.com/kyma-project/kyma/resources"

# shellcheck disable=SC2126
RES=$(grep -e 'version:\s*[Pp][Rr]-.*' -e 'image:.*:[Pp][Rr]-.*' -r "${RESOURCES_PATH}" -B 2 || true)
echo -n "$RES"
exit "$(echo -n "$RES" | wc -l | xargs)"