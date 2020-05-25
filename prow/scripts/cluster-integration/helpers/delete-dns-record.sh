#!/usr/bin/env bash

set -e
set -o pipefail

if [ -x /prow-tools/dnscleaner ];
then
  /prow-tools/dnscleaner "$@"
else
  cd "development/tools"
  go run "cmd/dnscleaner" "$@"
fi
echo "DNS Record deleted, but it can be visible for some time due to DNS caches"
