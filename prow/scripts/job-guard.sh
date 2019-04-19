#!/bin/bash

set -e

goPath=${GOPATH:-/home/prow/go}

cd "${goPath}"/src/github.com/kyma-project/test-infra/development/tools
make resolve
go run ./cmd/jobguard/main.go