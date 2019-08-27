#!/usr/bin/env bash

set -e

readonly RESOURCES_PATH="/home/prow/go/src/github.com/kyma-project/kyma/resources"

# shellcheck disable=SC2126
RES=$(grep -e 'version:\s*[Pp][Rr]-.*' -e 'image:.*:[Pp][Rr]-.*' -r "${RESOURCES_PATH}" | wc -l | xargs)

exit "${RES}"