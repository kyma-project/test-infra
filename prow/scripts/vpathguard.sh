#!/usr/bin/env bash

# shellcheck disable=SC1091
source /home/prow/go/src/github.com/kyma-project/master_test-infra/prow/scripts/log.sh

# Print test-infra commit on which image was built.
if [ -n "${IMAGE_COMMIT:+set}" ]; then
  echo "$IMAGE_COMMIT"
fi

if [ -d /home/prow/go/src/github.com/kyma-project/test-infra/vpath ]; then
  log::error "Directory vpath is present. Remove it to merge PR."
  exit 1
else
  log:success "Directory vpath is not present."
  exit 0
fi