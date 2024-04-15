#!/usr/bin/env bash

set -e

echo "Binary existence checks"
docker run --rm \
  "$IMG" bash -c '
  set -e
  docker --version
  helm version
  kubectl version --client
  k3d version
  kind version
  jobguard -h
  env
  '