#!/bin/sh

# shellcheck disable=SC2046
hadolint \
    "$(echo "$IGNORED_RULES" | sed -e 's/^\| / --ignore/' )" \
    --no-color \
    "$@"
