#!/usr/bin/env bash

set -eu

echo "--------------------------------------------------------------------------------"
echo "Generating GitHub stats..."
echo "--------------------------------------------------------------------------------"

export APP_GITHUB_ACCESS_TOKEN=${BOT_GITHUB_TOKEN}

/prow-tools/githubstats -o kyma-project -r kyma
/prow-tools/githubstats -o kyma-project -r helm-broker
/prow-tools/githubstats -o kyma-project -r rafter
/prow-tools/githubstats -o kyma-project -r test-infra
/prow-tools/githubstats -o kyma-incubator -r compass
