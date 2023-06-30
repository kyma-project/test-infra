#!/usr/bin/env bash

set -e

# WORKAROUND
#TODO (@Ressetkk): Use bundled image with docker-credential-gcr and docker
if ! command -v docker-credential-gcr; then
  curl -fsSLo docker-credential-gcr.tar.gz "https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v2.1.10/docker-credential-gcr_linux_amd64-2.1.10.tar.gz" && \
  tar xzf docker-credential-gcr.tar.gz \
  && chmod +x docker-credential-gcr && mv docker-credential-gcr /usr/bin/
fi

docker-credential-gcr configure-docker --registries=europe-docker.pkg.dev

REGISTRY="europe-docker.pkg.dev/kyma-project/prod/testimages"
TAG="$(date +v%Y%m%d)-${PULL_BASE_SHA::8}"

# Add new image here
images=(
e2e-dind-nodejs
e2e-dind-k3d
e2e-gcloud
buildpack-go
e2e-garden
)

docker buildx create --driver docker-container --use --name builder

toPush=()
for v in "${images[@]}"; do
  echo "building $v..."
  docker buildx build \
    --builder "builder" \
    --load \
    -t "$REGISTRY/$v:latest" \
    -t "$REGISTRY/$v:$TAG" \
    "./$v"
  toPush+=("$REGISTRY/$v")
done

if [[ "$1" == "push" ]]; then
  for v in "${toPush[@]}"; do
    echo "pushing $v"
    docker push -a "$v"
  done
fi