#!/usr/bin/env bash

set -e
set -o pipefail

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

EXCLUDED_BUCKETS='kyma-prow-secrets,kyma-prow-logs,kyma-prow-artifacts,kyma-development-artifacts,kyma-backup-restore,eu.artifacts.sap-kyma-prow.appspot.com'
BUCKET_REGEXP_NAME="^.+-([a-z0-9]+$)"
BUCKET_OBJECT_WORKERS_NUMBER=10

# [panic|fatal|error|warn|warning|info|debug|trace]
LOG_LEVEL=debug

if [ ! -d "${DEVELOPMENT_DIR}/tools/vendor" ]; then
    echo "Vendoring 'tools'"
    pushd "${DEVELOPMENT_DIR}/tools"
    dep ensure -v -vendor-only
    popd
fi

go run "${DEVELOPMENT_DIR}"/tools/cmd/gcscleaner/main.go \
      -bucketNameRegexp  "${BUCKET_REGEXP_NAME}"\
      -excludedBuckets "${EXCLUDED_BUCKETS}"\
      -workerNumber "${BUCKET_OBJECT_WORKERS_NUMBER}"\
      -logLevel "${LOG_LEVEL}"
      "$@"
