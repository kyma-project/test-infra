#!/bin/sh

# shellcheck disable=SC2046
hadolint \
    --config .hadolint.yaml \
    --no-color \
    "$@"
