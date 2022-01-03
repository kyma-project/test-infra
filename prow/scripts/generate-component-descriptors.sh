#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"

KYMA_RESOURCES_DIR="/home/prow/go/src/github.com/kyma-project/kyma/resources"

# shellcheck source=prow/scripts/lib/docker.sh
source "$SCRIPT_DIR/lib/docker.sh"

docker::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"

params=()
params+=("--component-version $(date +v%Y%m%d-%s)-${git_commit::8}")
params+=("--git-commit ${git_commit}")
params+=("--git-branch ${PULL_BASE_REF}")
params+=("--output-dir ${ARTIFACTS}/cd")
params+=("--skip-image-hashing=$skip_hashing")
if [[ "$JOB_TYPE" == "presubmit" ]]; then
    # on presubmit use latest commit from the PR itself
    git_commit="${PULL_PULL_SHA}"
    skip_hashing="true"
else
    # use base commit for postsubmit jobs
    git_commit="${PULL_BASE_SHA}"
    skip_hashing="false"
    params+=("--repo-context ${DOCKER_PUSH_REPOSITORY}")
fi

pushd "${TEST_INFRA_SOURCES_DIR}"
echo "This tool generates component descriptor file"
set -x
go run ./development/image-url-helper \
    --resources-directory "$KYMA_RESOURCES_DIR" \
    components \
    "${params[@]}"
set +x
echo "Compomnent descriptor was generated succesfully finished"
popd
