#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

usage () {
    echo "Usage: \$ ${BASH_SOURCE[1]} /path/to/component"
    exit 1
}

readonly SOURCES_DIR=$1

if [[ -z "${SOURCES_DIR}" ]]; then
    usage
fi

init

if [ -n "${PULL_NUMBER}" ]; then
  echo "Building from PR"
  DOCKER_TAG="PR-${PULL_NUMBER}"
elif [[ "${PULL_BASE_REF}" =~ release-.* ]]; then
  echo "Building from release ${PULL_BASE_REF}"
  DOCKER_TAG=$(cat "${SCRIPT_DIR}/../../RELEASE_VERSION")
  echo "Reading docker tag from RELEASE_VERSION file, got: ${DOCKER_TAG}"

  if [[ "${REPO_OWNER}" == "kyma-project" && "${REPO_NAME}" == "kyma" ]]; then
    NEXT_RELEASE=$(cat "${SCRIPT_DIR}/../RELEASE_VERSION")
    echo "Checking if ${NEXT_RELEASE} was already published on github..."
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" https://api.github.com/repos/kyma-project/kyma/releases/tags/"${NEXT_RELEASE}")
    if [[ $RESPONSE != 404* ]]; then
        echo "The ${NEXT_RELEASE} is already published on github. Stopping."
        exit 1
    fi
  fi

else
  echo "Building as usual"
  DOCKER_TAG=$(echo "${PULL_BASE_SHA}" | cut -c1-8)
  set +e
  DOCKER_POST_PR_TAG="PR-$(git -C "${SOURCES_DIR}" log --format=%B -n 1 "${PULL_BASE_SHA}" | grep -o '(#[[:digit:]]\+)$' | grep -o '[[:digit:]]\+')"
  set -e
fi

readonly DOCKER_TAG
export DOCKER_TAG
echo DOCKER_TAG "${DOCKER_TAG}"

if [ -n "${DOCKER_POST_PR_TAG}" ]; then
    readonly DOCKER_POST_PR_TAG
    export DOCKER_POST_PR_TAG
    echo DOCKER_POST_PR_TAG "${DOCKER_TAG}"
fi

make -C "${SOURCES_DIR}" release