#!/usr/bin/env bash

set -e
set -o pipefail

EXCLUDED_BUCKETS='kyma-prow-secrets,kyma-prow-logs,kyma-prow-artifacts,kyma-development-artifacts,kyma-backup-restore,eu.artifacts.sap-kyma-prow.appspot.com'
BUCKET_REGEXP_NAME="^.+-([a-z0-9]+$)"

go run "${DEVELOPMENT_DIR}"/tools/cmd/gcscleaner/main.go \
      -bucketNameRegexp  "${BUCKET_REGEXP_NAME}"\
      -excludedBuckets "${EXCLUDED_BUCKETS}"\
      "$@"