#!/usr/bin/env bash

set -e
set -o pipefail

if [ -x /prow-tools/ipcleaner ];
then
  /prow-tools/ipcleaner "$@"
else
  cd "development/tools"
  go run "cmd/ipcleaner" "$@"
fi
