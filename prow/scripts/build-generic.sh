#!/usr/bin/env bash

set -e
set -o pipefail

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=prow/scripts/lib/gcloud.sh
source "${SCRIPT_DIR}/lib/gcloud.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${SCRIPT_DIR}/lib/docker.sh"

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} /path/to/component [Makefile targets]"
    exit 1
}

# get first argument and assume it's a path to the sources.
readonly SOURCES_DIR=$1; shift
if [[ ! -d "${SOURCES_DIR}" ]]; then
  echo -e "Error: Directory \"$SOURCES_DIR\" does not exist."
  usage
fi

gcloud::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"
docker::start

mkdir -p /tmp/artifacts/

echo "K8S_NODE_NAME ${K8S_NODE_NAME}"
echo "HOME: ${HOME}"
echo "CACHE_DIR: ${CACHE_DIR}"

ls -la "${CACHE_DIR}"
touch "${CACHE_DIR}/test.txt"
ls -la "${CACHE_DIR}"

#ls -la "${CACHE_DIR}/go-build" || true
rm -rf "${CACHE_DIR}/go-build"

go version

if [ -n "${PULL_NUMBER}" ]; then
  echo "Building from PR"
  DOCKER_TAG="PR-${PULL_NUMBER}"
else
  # Build artifacts using short SHA for all branches postsubmits
  echo "Building as usual"
  DOCKER_TAG="${PULL_BASE_SHA::8}"
fi

export DOCKER_TAG
echo DOCKER_TAG "${DOCKER_TAG}"

# Adding script argument checking allows to define custom build targets because `ci-release` is not in several Makefiles.
if [ -n "$1" ]; then
  make -C "${SOURCES_DIR}" "$@" 2>&1 | while read -r line ; do printf "[%04d] | %s\n" $SECONDS "$line"; done;
else
  make -C "${SOURCES_DIR}" release 2>&1 | while read -r line ; do printf "[%04d] | %s\n" $SECONDS "$line"; done;
fi
