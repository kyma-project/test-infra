#!/usr/bin/env bash

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# shellcheck disable=SC1091
# shellcheck source=/dev/null
source "$SCRIPT_DIR/lib/log.sh"

# Print test-infra commit on which image was built.
if [ -n "${IMAGE_COMMIT:+set}" ]; then
  echo "IMAGE_COMMIT: $IMAGE_COMMIT"
fi

if [ -d /home/prow/go/src/github.com/kyma-project/test-infra/vpath ]; then
  log::error "Directory vpath is present. Remove it to merge PR."
  exit 1
else
  log:success "Directory vpath is not present."
  exit 0
fi