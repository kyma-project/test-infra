#!/usr/bin/env bash

set -e
set -o pipefail

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

EXCLUDED_BUCKETS='kyma-prowjobs-secrets,kyma-prow-secrets,kyma-prow-logs,kyma-prow-artifacts,kyma-development-artifacts,kyma-backup-restore,eu.artifacts.sap-kyma-prow.appspot.com'
BUCKET_REGEXP_NAME="^.+-([a-z0-9]+$)"
BUCKET_OBJECT_WORKERS_NUMBER=10

# [panic|fatal|error|warn|warning|info|debug|trace]
LOG_LEVEL=info

"${DEVELOPMENT_DIR}/resources-cleanup.sh" "gcscleaner" "assetstore GCP buckets" \
      -bucketNameRegexp  "${BUCKET_REGEXP_NAME}"\
      -excludedBuckets "${EXCLUDED_BUCKETS}"\
      -workerNumber "${BUCKET_OBJECT_WORKERS_NUMBER}"\
      -logLevel "${LOG_LEVEL}" "$@"
