#!/usr/bin/env bash

set -o errexit
set -o pipefail

CYPRESS_IMAGE="eu.gcr.io/kyma-project/external/cypress/included:8.7.0"

mkdir -p "$PWD/kyma-dashboard-tests/cypress/screenshots"

function load_env() {
  ENV_FILE=".env"
  if [ -f "${ENV_FILE}" ]; then
    # shellcheck disable=SC2046
    export $(xargs < "${ENV_FILE}")
  fi
}

load_env
echo DOCKER_TAG "${DOCKER_TAG}"
# shellcheck disable=SC2086
docker run -d --rm --net=host --pid=host --name busola eu.gcr.io/kyma-project/kyma-dashboard-local-dev:${DOCKER_TAG}

cp "$PWD/kubeconfig-kyma.yaml" "$PWD/kyma-dashboard-tests/fixtures/kubeconfig.yaml"

echo "STEP: Running Cypress tests inside Docker"
docker run --entrypoint /bin/bash --network=host -v "$PWD/kyma-dashboard-tests:/tests" -w /tests $CYPRESS_IMAGE -c "npm ci --no-optional; NO_COLOR=1 CYPRESS_DOMAIN=http://localhost:3001 cypress run --browser chrome -C /tests/cypress.json"
