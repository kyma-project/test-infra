#!/usr/bin/env bash

set -e

echo "Binary existence checks"
docker run --rm --privileged \
  -v /sys/fs/cgroup:/sys/fs/cgroup \
  -v /lib/modules:/lib/modules:ro \
  -e DOCKER_IN_DOCKER_ENABLED=true \
  "$IMG" bash -c '
  set -e
  cat $ARTIFACTS/docker-info.log
  docker ps -a
  helm version
  kubectl version --client
  k3d version
  kind version
  jobguard -h
  env
  '