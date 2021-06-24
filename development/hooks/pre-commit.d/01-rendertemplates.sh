#!/bin/bash

if ! [ `command -v go`]; then
  echo -e "ERROR: golang is not available. Please install the latest version of Go from https://golang.org/dl/!"
  exit 1
fi


pushd "$(git rev-parse --show-toplevel)"
  if $(git diff --quiet -- "templates"); then
    go run development/tools/cmd/rendertemplates -config templates/config.yaml
  fi
popd

git add prow/jobs
