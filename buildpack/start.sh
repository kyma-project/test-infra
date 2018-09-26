#!/bin/bash

set -o errexit

service docker start
# the service can be started but the docker socket not ready, wait for ready
WAIT_N=0
MAX_WAIT=10
while true; do
    # docker ps -q should only work if the daemon is ready
    docker ps -q > /dev/null 2>&1 && break
    if [[ ${WAIT_N} -lt ${MAX_WAIT} ]]; then
        WAIT_N=$((WAIT_N+1))
        echo "Waiting for docker to be ready, sleeping for ${WAIT_N} seconds."
        sleep ${WAIT_N}
    else
        echo "Reached maximum attempts, not waiting any longer..."
        break
    fi
done

docker system prune --all --force --volumes

if [ ! -f /proc/sys/fs/binfmt_misc/status ]; then
    mount binfmt_misc -t binfmt_misc /proc/sys/fs/binfmt_misc
fi
echo -1 >/proc/sys/fs/binfmt_misc/status

printf '=%.0s' {1..80}; echo
echo "Done setting up docker in docker."

