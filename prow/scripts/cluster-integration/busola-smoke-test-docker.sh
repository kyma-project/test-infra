#!/usr/bin/env bash

set -o errexit
set -o pipefail

CYPRESS_IMAGE="eu.gcr.io/kyma-project/external/cypress/included:8.7.0"

mkdir -p "$PWD/busola-tests/cypress/screenshots"

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
docker run -d --rm --net=host --pid=host --name busola europe-docker.pkg.dev/kyma-project/dev/busola:${DOCKER_TAG}

cp "$PWD/kubeconfig-kyma.yaml" "$PWD/busola-tests/fixtures/kubeconfig.yaml"

echo "STEP: Running Cypress tests inside Docker"
docker run --entrypoint /bin/bash --network=host -v "$PWD/busola-tests:/tests" -w /tests $CYPRESS_IMAGE -c "npm ci --no-optional; NO_COLOR=1 CYPRESS_DOMAIN=http://localhost:3001 cypress run --browser chrome -C /tests/cypress-smoke.json"
