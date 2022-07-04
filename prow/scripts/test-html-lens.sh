#!/usr/bin/env bash



golangci-lint  run ./... --out-format html > "${ARTIFACTS}/report-golint.html"

curl -L -o "${ARTIFACTS}/report-kyma.html" "https://kyma-project.io"
