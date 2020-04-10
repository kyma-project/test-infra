#!/usr/bin/env bash

set -eu

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "--------------------------------------------------------------------------------"
echo "Generating GitHub stats..."
echo "--------------------------------------------------------------------------------"

export APP_GITHUB_ACCESS_TOKEN=${BOT_GITHUB_TOKEN}

/prow-tools/cmd/githubstats -o kyma-project -r kyma
/prow-tools/cmd/githubstats -o kyma-project -r helm-broker
/prow-tools/cmd/githubstats -o kyma-project -r rafter
/prow-tools/cmd/githubstats -o kyma-project -r test-infra

/prow-tools/cmd/githubstats -o kyma-incubator -r compass
