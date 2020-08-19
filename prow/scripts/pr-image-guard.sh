#!/usr/bin/env bash

set -e

readonly RESOURCES_PATH="/home/prow/go/src/github.com/kyma-project/kyma/resources"
readonly TESTS_PATH="/home/prow/go/src/github.com/kyma-project/kyma/tests"

# shellcheck disable=SC2126
RES=$(grep -e 'version:\s.*[Pp][Rr]-.*' -e 'image:.*:[Pp][Rr]-.*' -e 'tag:\s.*[Pp][Rr]-.*' -r "${RESOURCES_PATH}" -r "${TESTS_PATH}" -B 2 || true)
echo -n "$RES"
exit "$(echo -n "$RES" | wc -l | xargs)"