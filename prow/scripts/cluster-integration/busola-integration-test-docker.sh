#!/usr/bin/env bash

set -o errexit
set -o pipefail

CYPRESS_IMAGE="eu.gcr.io/kyma-project/external/cypress/included:8.6.0"

mkdir -p "$PWD/busola-tests/cypress/screenshots"

DOCKER_TAG=test
if [ -n "${PULL_NUMBER}" ]; then
  echo "Building from PR"
  DOCKER_TAG="PR-${PULL_NUMBER}"
else
  # Build artifacts using short SHA for all branches postsubmits
  echo "Building as usual"
  DOCKER_TAG="${PULL_BASE_SHA::8}"
fi

#export DOCKER_TAG
echo DOCKER_TAG "${DOCKER_TAG}"

# shellcheck disable=SC2086
docker run -d --rm --net=host --pid=host --name busola eu.gcr.io/kyma-project/busola:${DOCKER_TAG}

cp "$PWD/kubeconfig-kyma.yaml" "$PWD/busola-tests/fixtures/kubeconfig.yaml"

echo "STEP: Running Cypress tests inside Docker"
docker run --entrypoint /bin/bash --network=host -v "$PWD/busola-tests:/tests" -w /tests $CYPRESS_IMAGE -c "npm ci --no-optional; NO_COLOR=1 CYPRESS_DOMAIN=http://localhost:3001 cypress run --browser chrome"
