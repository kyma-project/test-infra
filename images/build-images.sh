#!/usr/bin/env bash

set -e

# WORKAROUND
#TODO (@Ressetkk): Use bundled image with docker-credential-gcr and docker
if [[ $CI == "true" ]]; then
  if ! command -v docker-credential-gcr; then
    curl -fsSLo docker-credential-gcr.tar.gz "https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v2.1.10/docker-credential-gcr_linux_amd64-2.1.10.tar.gz" && \
    tar xzf docker-credential-gcr.tar.gz \
    && chmod +x docker-credential-gcr && mv docker-credential-gcr /usr/bin/
  fi

  docker-credential-gcr configure-docker --registries=europe-docker.pkg.dev
fi
REGISTRY="europe-docker.pkg.dev/kyma-project/prod/testimages"
TAG="$(date +v%Y%m%d)-${PULL_BASE_SHA::8}"

toPush=()
for v in $(find . -type d -exec test -e '{}'/Dockerfile \; -print | cut -c3-) ; do
  name=$(echo "$v" | sed "s/\//-/g")
  echo "building $name..."
  docker buildx build \
    --load \
    -t "$REGISTRY/$name:latest" \
    -t "$REGISTRY/$name:$TAG" \
    "./$v"
  toPush+=("$REGISTRY/$name")
done

if [[ "$1" == "push" ]]; then
  for v in "${toPush[@]}"; do
    echo "pushing $v"
    docker push -a "$v"
  done
fi