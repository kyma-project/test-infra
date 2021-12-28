#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly TEST_INFRA_SOURCES_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"

KYMA_RESOURCES_DIR="/home/prow/go/src/github.com/kyma-project/kyma/resources"

# shellcheck source=prow/scripts/lib/docker.sh
source "$SCRIPT_DIR/lib/docker.sh"

docker::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"

pushd "${TEST_INFRA_SOURCES_DIR}"
echo "This tool generates component descriptor file"
go run ./development/image-url-helper \
    --resources-directory "$KYMA_RESOURCES_DIR" \
    components \
    --component-version "$(date +v%Y%m%d)-${PULL_PULL_SHA::8}" \
    --git-commit "${PULL_PULL_SHA}" \
    --git-branch "${PULL_BASE_REF}" \
    --output-dir "${ARTIFACTS}/cd" \
    --repo-context "${DOCKER_PUSH_REPOSITORY}"
echo "Compomnent descriptor was generated succesfully finished"
popd
