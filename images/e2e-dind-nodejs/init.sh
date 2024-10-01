#!/usr/local/bin/dumb-init /bin/bash
# shellcheck shell=bash

set -e

cleanup_dind() {
    if [[ "${DOCKER_IN_DOCKER_ENABLED:-false}" == "true" ]]; then
        echo "⏳  Cleaning up after docker"
        docker ps -aq | xargs -r docker rm -f || true
        docker image ls -aq | xargs -r docker image rm -r || true
        kill -TERM "$DOCKER_PID" || true
        echo "✅  Done cleaning up after docker"
    fi
}

# It's called using trap for INT TERM signal
# shellcheck disable=SC2317
early_exit_handler() {
    if [ -n "${WRAPPED_COMMAND_PID:-}" ]; then
        kill -TERM "$WRAPPED_COMMAND_PID" || true
    fi
    cleanup_dind
}

trap early_exit_handler INT TERM

LOG_DIR=${ARTIFACTS:-"/var/log"}

if [[ "${DOCKER_IN_DOCKER_ENABLED}" == "true" ]]; then
  echo "⏳  Starting Docker in Docker"
  dockerd --data-root=/docker-graph > "${LOG_DIR}/dockerd.log" 2>&1 &
  DOCKER_PID=$!
  echo "⏳  Waiting for Docker to be up..."
  while ! curl -s --unix-socket /var/run/docker.sock http/_ping >/dev/null 2>&1; do
    sleep 1
  done
  echo "✅  Docker is up! Docker info available in ${ARTIFACTS}/docker-info.log"
  docker info > "${ARTIFACTS}/docker-info.log"
fi

if [[ "$K3D_ENABLED" == "true" ]]; then
  ARGS=()
  echo -n "⏳ Provisioning k3d cluster"
  if [[ "$PROVISION_REGISTRY" == "true" ]]; then
    echo " with registry k3d-registry.localhost:5000"
    k3d registry create registry.localhost --port 5000
    ARGS+=( "--registry-use=k3d-registry.localhost:5000" )
  else
    echo
  fi
  k3d cluster create k3d "${ARGS[@]}"
  echo "✅  Cluster provisioned successfully!"
fi

# actually start bootstrap and the job
#set -o xtrace
"$@" &
WRAPPED_COMMAND_PID=$!
wait $WRAPPED_COMMAND_PID
EXIT_VALUE=$?
#set +o xtrace

# cleanup after job
if [[ "${DOCKER_IN_DOCKER_ENABLED}" == "true" ]]; then
    cleanup_dind
fi

# preserve exit value from job / bootstrap
exit ${EXIT_VALUE}
