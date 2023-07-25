#!/usr/bin/env bash
#
# This script is meant to serve backwards compatibility during migrating all build jobs to gcbuild
# without the need to implement workload identity beforehand.

if [ ! -z "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
  gcloud auth activate-service-account --key-file="$GOOGLE_APPLICATION_CREDENTIALS"
fi

/gcbuild "$@"