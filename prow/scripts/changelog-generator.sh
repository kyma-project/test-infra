#!/usr/bin/env bash
# DEPRECATED DO NOT USE
set -e

readonly KYMA_DIR="/home/prow/go/src/github.com/kyma-project/kyma"
readonly RELEASE_VERSION=$(cat "${KYMA_DIR}/VERSION")

# find latest tag from which the generator should started
# shellcheck disable=SC2046
TAG_LIST_STRING=$(git describe --tags $(git rev-list --tags) --always | grep -F . | grep -v "-")
IFS=" " read -r -a TAG_LIST <<< "${TAG_LIST_STRING}"
PENULTIMATE=${TAG_LIST[0]}

if [ "${PENULTIMATE}" = "" ]; then
    echo "PENULTIMATE tag does not exist, first commit of repository will be use."
    PENULTIMATE=$(git rev-list --max-parents=0 HEAD)
fi

#check if github token exist
if [[ "${BOT_GITHUB_TOKEN}" == "" ]]; then
    echo "Bot github token is empty, cannot create changelog file by lerna."
    exit 0
fi

#generate release changelog
docker run --rm -v "${KYMA_DIR}":/repository -w /repository -e FROM_TAG="${PENULTIMATE}" -e NEW_RELEASE_TITLE="${RELEASE_VERSION}" -e GITHUB_AUTH="${BOT_GITHUB_TOKEN}" -e CONFIG_FILE=.github/package.json eu.gcr.io/kyma-project/changelog-generator:0.2.0 sh /app/generate-release-changelog.sh;

#copy changelog file to KYMA_ARTIFACTS_BUCKET destination
cp "${KYMA_DIR}/.changelog/release-changelog.md" "${ARTIFACTS}/release-changelog.md"
gsutil cp "${KYMA_DIR}/.changelog/release-changelog.md" "${KYMA_ARTIFACTS_BUCKET}/${DOCKER_TAG}/release-changelog.md"
