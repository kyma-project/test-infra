#!/usr/bin/env bash

set -eu

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"

log::banner "Generating GitHub stats..."

export APP_GITHUB_ACCESS_TOKEN=${BOT_GITHUB_TOKEN}

/prow-tools/githubstats -o kyma-project -r kyma
/prow-tools/githubstats -o kyma-project -r helm-broker
/prow-tools/githubstats -o kyma-project -r rafter
/prow-tools/githubstats -o kyma-project -r test-infra
/prow-tools/githubstats -o kyma-incubator -r compass
