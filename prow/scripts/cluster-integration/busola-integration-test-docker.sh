#!/usr/bin/env bash

set -o errexit
set -o pipefail

CYPRESS_IMAGE="eu.gcr.io/kyma-project/external/cypress/included:8.6.0"

install_busola(){
    # shellcheck disable=SC2086
    docker run --rm --net=host --pid=host --name busola eu.gcr.io/kyma-project/busola:$DOCKER_TAG
}

cp "$PWD/kubeconfig-kyma.yaml" "$PWD/busola-tests/fixtures/kubeconfig.yaml"
mkdir -p "$PWD/busola-tests/cypress/screenshots"

echo "STEP: Running Cypress tests inside Docker"
docker run --entrypoint /bin/bash --network=host -v "$PWD/busola-tests:/tests" -w /tests $CYPRESS_IMAGE -c "npm ci --no-optional; NO_COLOR=1 CYPRESS_DOMAIN=localhost:3001 cypress run --browser chrome"

