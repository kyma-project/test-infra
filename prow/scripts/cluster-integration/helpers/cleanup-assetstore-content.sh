#!/usr/bin/env bash
set +e
for resource in assets docstopics buckets clusterassets clusterdocstopics clusterbuckets; do
    kubectl -n kyma-system delete --all ${resource}
done