#!/usr/local/bin/dumb-init /bin/bash

set -e

LOG_DIR=${ARTIFACTS:-"/var/log"}

if [[ "${DOCKER_IN_DOCKER_ENABLED}" == "true" ]]; then
  echo "[ * * * ] Starting Docker in Docker"
  dockerd --data-root=/docker-graph > "${LOG_DIR}/dockerd.log" 2>&1 &
  sleep 5 # sleep to wait for Docker daemon - idk if we can use some kind of condition
fi

if [[ "$K3D_ENABLED" == "true" ]]; then
  ARGS=()
  echo -n "[ * * * ] Provisioning k3d cluster"
  if [[ "$PROVISION_REGISTRY" == "true" ]]; then
    echo " with registry k3d-registry.localhost:5000"
    k3d registry create registry.localhost --port 5000
    ARGS+=( "--registry-use=k3d-registry.localhost:5000" )
  elif [[ -n "${K3D_SERVERS_MEMORY}" ]]; then
    echo " with servers memory ${K3D_SERVERS_MEMORY}"
    ARGS+=( "--servers-memory=${K3D_SERVERS_MEMORY}" )
  else
    echo
  fi
  k3d cluster create k3d "${ARGS[@]}"
fi
exec "$@"