#!/usr/bin/env bash

set -e
set -o pipefail

if [ -x /prow-tools/dnscleaner ];
then
  /prow-tools/dnscleaner --attempts=10 --ttl 60 "$@"
else
  cd "development/tools"
  go run "cmd/dnscleaner" --attempts=10 --ttl 60 "$@"
fi

echo "DNS Record deleted, but it can be visible for some time due to DNS caches"
