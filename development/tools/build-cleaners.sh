#!/bin/bash

set -eu

readonly TOOLS_DIR="${PWD}/development/tools"
cd "${TOOLS_DIR}"

mkdir -p ./bin

for D in "${TOOLS_DIR}"/cmd/*;
do
  if [ -d "$D" ];
  then
    name=$(basename "${D}")
    echo "building ${name}..."
    cd "${TOOLS_DIR}/cmd/${name}"
    if [ -a "nobuild.lock" ]; then
      echo "No go files. Skipped..."
      continue
    fi
    CGO_ENABLED=0 go build -o "${TOOLS_DIR}/bin/${name}" -ldflags="-s -w" main.go
    chmod a+x "${TOOLS_DIR}/bin/${name}"
  fi
done
