#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"

KYMA_RESOURCES_DIR="/home/prow/go/src/github.com/kyma-project/kyma/resources"

# shellcheck source=prow/scripts/lib/docker.sh
source "$SCRIPT_DIR/lib/docker.sh"

docker::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"

params=()

params+=("--git-branch=${PULL_BASE_REF}")
params+=("--output-dir=${ARTIFACTS}/cd")
if [[ "$JOB_TYPE" == "presubmit" ]]; then
    # on presubmit use latest commit from the PR itself
    git_commit="${PULL_PULL_SHA}"
    params+=("--skip-image-hashing=true")
else
    # use base commit for postsubmit jobs
    git_commit="${PULL_BASE_SHA}"
    params+=("--repo-context=${DOCKER_PUSH_REPOSITORY}")
fi

params+=("--component-version=$(date +v%Y%m%d-%H%M%S)-${git_commit::8}")
params+=("--git-commit=${git_commit}")

pushd "${TEST_INFRA_SOURCES_DIR}"
log::info "This tool generates component descriptor file"
/prow-tools/image-url-helper \
    --resources-directory "$KYMA_RESOURCES_DIR" \
    components \
    "${params[@]}"

log::info "Compomnent descriptor was generated succesfully finished"
popd
