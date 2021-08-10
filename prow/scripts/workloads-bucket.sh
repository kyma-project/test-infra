#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=prow/scripts/lib/docker.sh
source "${SCRIPT_DIR}/lib/docker.sh"

docker::start

docker pull alpine:edge
docker tag alpine:edge eu.gcr.io/sap-kyma-prow-workloads/alpine_edge
docker push eu.gcr.io/sap-kyma-prow-workloads/alpine_edge
