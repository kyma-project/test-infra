#!/bin/sh
IGNORED_RULES=($IGNORED_RULES)
hadolint \
    "${IGNORED_RULES[@]/#/--ignore }" \
    --no-color \
    "$@"
