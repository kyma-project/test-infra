#!/bin/bash

# builder.sh rebuilds all images in prow/images directory in specific order.
# This ensures that the test-infra images used in the pipelines are consistent between each other and always contain latest changes from main.
# TODO (@Ressetkk) rewrite this script in Go using buildah as a library...
set -e

required=( yq jq buildah )
for p in "${required[@]}"; do
  if ! command -v "$p"; then
    echo -e "$p not found. Exiting..."
  fi
done

function usage() {
  echo 'Chain build test-infra images.

Options:
-c path     Path to build config YAML file.
--debug     Enable debug output.
--dry-run   Enable dry-run mode.'
}

function debug() {
  if [ "$DEBUG" == "true" ]; then
    echo -e "$(date "+%d/%m/%Y %X") DEBUG: $*"
  fi
}

function info() {
  echo -e "$(date "+%d/%m/%Y %X") INFO: $*"
}

function error() {
  echo -e "$(date "+%d/%m/%Y %X") ERROR: $*"
}

function run() {
  debug "running command: $*"
  randomstr=$(tr -dc '[:alnum:]' < /dev/urandom | dd bs=4 count=2 2>/dev/null)
  cmdLogFile="${buildName:-$randomstr}.log"
  cmdLogPath="/tmp/$cmdLogFile"
  if [ "$DRY_RUN" != "true" ]; then
    "$@" &> "$cmdLogPath" && (cp "$cmdLogPath" "$ARTIFACTS/$cmdLogFile"; info "Image $buildName built successfully!") || ( error "An error occured during command execution"; cat "$cmdLogPath"; exit 1 )
  fi

}

ARTIFACTS="${ARTIFACTS:-/tmp}"

while [ $# -gt 0 ]; do
  case $1 in
  "-c") # path to the config
    BUILD_CONFIG_PATH="$2"
    CONFIG="$(yq e -o=j $BUILD_CONFIG_PATH)"
    shift
    shift
    ;;
  "--debug")
    DEBUG="true"
    shift
    ;;
  "--dry-run")
    DRY_RUN="true"
    shift
    ;;
  *)
    echo -e "Unknown opition \"$1\""
    usage
    exit 1
  esac
done

debug "$DEBUG"
debug "$BUILD_CONFIG_PATH"

if [ "$JOB_TYPE" == "presubmit" ]; then
  BASE_TAG="PR-$PULL_NUMBER"
else
  BASE_TAG="$(date +v%Y%m%d)-$(git rev-parse --short HEAD)"
fi

IFS=''
while read -r s; do
  steps+=("$s")
done < <(echo $CONFIG | jq -c '.steps[]' -)

for step in ${steps[@]}; do
  stepName=$(echo "$step" | jq -c '.name' -)
  info "Executing step: $stepName"
  # run through all images in step and build them based on properties
  echo "$step" | jq -r '.images[] | .name, .context, .remotePrefix, .dockerfile' - | (
    while read buildName; do
      read context
      read remotePrefix
      read dockerfile

      if [ "$context" == "null" ]; then
        echo -e "context cannot be empty! Exiting..."
        exit 1
      fi
      # TODO multiple dockerfiles
      if [ "$dockerfile" == "null" ]; then
        dockerfile="Dockerfile"
      fi

      info "Building $buildName..."
      debug "bn=$buildName ctx=$context re=$remotePrefix df=$dockerfile"
      TAGS=( "-t=$buildName" )
      if [ "$remotePrefix" != "null" ]; then
        # TODO tagging based on a list of tags instead of hardcode it
        remoteName="$remotePrefix/$buildName"
        TAGS+=( "-t=$remoteName:$BASE_TAG" )
        REMOTE_PUSH+=( "$remoteName" )
      fi
      run buildah bud -f="$dockerfile" "${TAGS[@]}" "$context"
    done
  )
done

  # push to GCR
  if ! [ -z ${GOOGLE_APPLICATION_CREDENTIALS+x} ]; then
    run "cat $GOOGLE_APPLICATION_CREDENTIALS | buildah login -u _json_key --password-stdin https://eu.gcr.io"
    for r in ${REMOTE_PUSH[@]}; do
      run buildah push "$r"
    done
  fi
