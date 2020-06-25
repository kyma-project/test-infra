#!/usr/bin/env bash

set -e
set -o pipefail

EXCLUDED_BUCKETS='kyma-prowjobs-secrets,kyma-prow-secrets,kyma-prow-logs,kyma-prow-artifacts,kyma-development-artifacts,kyma-backup-restore,eu.artifacts.sap-kyma-prow.appspot.com,kyma-prow-access-storage-logs,kyma-prow-workloads-logs'
BUCKET_REGEXP_NAME="^.+-([a-z0-9]+$)"
BUCKET_OBJECT_WORKERS_NUMBER=10

# [panic|fatal|error|warn|warning|info|debug|trace]
LOG_LEVEL=info

/prow-tools/gcscleaner -bucketNameRegexp "${BUCKET_REGEXP_NAME}"\
      -excludedBuckets "${EXCLUDED_BUCKETS}"\
      -workerNumber "${BUCKET_OBJECT_WORKERS_NUMBER}"\
      -logLevel "${LOG_LEVEL}" "$@"
