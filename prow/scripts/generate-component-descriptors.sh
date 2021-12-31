#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"

KYMA_RESOURCES_DIR="/home/prow/go/src/github.com/kyma-project/kyma/resources"

# shellcheck source=prow/scripts/lib/docker.sh
source "$SCRIPT_DIR/lib/docker.sh"

docker::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"

if [[ "$JOB_TYPE" == "presubmit" ]]; then
    # on presubmit use latest commit from the PR itself
    git_commit="${PULL_PULL_SHA}"
    skip_hashing="true"
else
    # use base commit for postsubmit jobs
    git_commit="${PULL_BASE_SHA}"
    skip_hashing="false"
fi

pushd "${TEST_INFRA_SOURCES_DIR}"
echo "This tool generates component descriptor file"
go run ./development/image-url-helper \
    --resources-directory "$KYMA_RESOURCES_DIR" \
    components \
    --component-version "$(date +v%Y%m%d)-${git_commit::8}" \
    --git-commit "${git_commit}" \
    --git-branch "${PULL_BASE_REF}" \
    --output-dir "${ARTIFACTS}/cd" \
    --repo-context "${DOCKER_PUSH_REPOSITORY}" \
    --skip-image-hashing="$skip_hashing"
echo "Compomnent descriptor was generated succesfully finished"
popd
