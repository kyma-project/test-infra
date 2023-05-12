#!/usr/local/bin/dumb-init /bin/bash

set -e

LOG_DIR=${ARTIFACTS:-"/var/log"}

if [[ -n "${DOCKER_IN_DOCKER_ENABLED}" ]]; then
  echo "[* * *] Starting Docker in Docker"
  dockerd --data-root=/docker-graph > "${LOG_DIR}/dockerd.log" 2>&1 &
fi

exec "$@"