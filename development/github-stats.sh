#!/usr/bin/env bash

set -eu

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "--------------------------------------------------------------------------------"
echo "Generating GitHub stats..."
echo "--------------------------------------------------------------------------------"

if [ ! -d "${DEVELOPMENT_DIR}/tools/vendor" ]; then
    echo "Vendoring 'tools'"
    pushd "${DEVELOPMENT_DIR}/tools"
    dep ensure -v -vendor-only
    popd
fi

export APP_GITHUB_ACCESS_TOKEN=${BOT_GITHUB_TOKEN}

go run "${DEVELOPMENT_DIR}/tools/cmd/githubstats/main.go" -o kyma-project -r kyma
go run "${DEVELOPMENT_DIR}/tools/cmd/githubstats/main.go" -o kyma-project -r helm-broker
go run "${DEVELOPMENT_DIR}/tools/cmd/githubstats/main.go" -o kyma-project -r rafter
go run "${DEVELOPMENT_DIR}/tools/cmd/githubstats/main.go" -o kyma-project -r test-infra

go run "${DEVELOPMENT_DIR}/tools/cmd/githubstats/main.go" -o kyma-incubator -r compass
