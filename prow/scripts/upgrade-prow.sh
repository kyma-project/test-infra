#!/bin/bash

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"

authenticate

TAG=$(curl https://raw.githubusercontent.com/kubernetes/test-infra/master/prow/cluster/starter.yaml | grep image | grep deck | awk '{print $2}' | cut -d ':' -f2)
DEPLOY_LIST="deck hook horologium plank sinker tide"

rollback() {
	for DEPLOY in ${DEPLOY_LIST}; do
    kubectl rollout undo deployment "${DEPLOY}"
done
}

trap rollback ERR

for DEPLOY in ${DEPLOY_LIST}; do
    kubectl set image deploy "${DEPLOY}" "${DEPLOY}=gcr.io/k8s-prow/${DEPLOY}:${TAG}" --record
done
