#!/usr/bin/env bash

source /home/prow/go/src/github.com/kyma-project/master_test-infra/prow/scripts/log.sh

if [ -d /home/prow/go/src/github.com/kyma-project/test-infra/vpath ]; then
  log::error "Directory vpath is present. Remove it to merge PR."
  exit 1
else
  log:success "Directory vpath is not present."
  exit 0
fi