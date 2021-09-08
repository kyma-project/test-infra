#!/bin/bash

# builder.sh rebuilds all images in prow/images directory in specific order.
# TODO (@Ressetkk) rewrite this script in Go using buildah as a library...

required=( yq jq )
for p in "${required[@]}"; do
  if ! $(command -v "$p" &> /dev/null); then
    echo -e "$p not found. Exiting..."
  fi
done

function usage() {
  echo 'Chain build test-infra images.

Options:
-c path   Path to build config YAML file.
-d        Enable debug output.'
}

function debug() {
  if [ "$DEBUG" == "true" ]; then
    echo -e "$(date "+%d/%m/%Y %X") DEBUG: $@"
  fi
}

function info() {
  echo -e "$(date "+%d/%m/%Y %X") INFO: $@"
}

while [ $# -gt 0 ]; do
  case $1 in
  "-c") # path to the config
    BUILD_CONFIG_PATH="$2"
    CONFIG="$(yq e -o=j $BUILD_CONFIG_PATH)"
    shift
    shift
    ;;
  "-d")
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

BASE_TAG="$(date +v%Y%m%d)-$(git rev-parse --short HEAD)"


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
      fi
      debug "buildah bud -f=\"$dockerfile\" ${TAGS[@]} \"$context\""
      buildah bud -f "$dockerfile" -t "${TAGS[@]}" "$context"

      # TODO authentication and push to the remote
    done
  )
done
