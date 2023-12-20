#!/bin/bash

echo "This jobguard location is deprecated. Please call jobguard from new location github.com/kyma-project/test-infra/development/jobguard"

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

cd "${ROOT_PATH}/../../../jobguard/scripts" || exit 1

ROOT_PATH="$(pwd)"

"${ROOT_PATH}/run.sh"
