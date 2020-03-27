#!/bin/bash

set -eu

if [ -z "${KYMA_CLEANERS_BUCKET}" ]; then
  echo "KYMA_CLEANERS_BUCKET is not set!"
  exit 1
fi

# shellcheck disable=SC2046
readonly TOOLS_DIR="${PWD}/development/tools"
cd "$TOOLS_DIR"

mkdir -p bin

for D in "$TOOLS_DIR"/cmd/*;
do
  if [ -d "$D" ];
  then
    name=$(basename "${D}")
    echo "building ${name}..."
    cd "$TOOLS_DIR"/cmd/"$name"
    go build -o "${TOOLS_DIR}"/bin/"${name}" -ldflags="-s -w" main.go
    # shellcheck disable=SC2086
    upx -q "${TOOLS_DIR}"/bin/${name}
    chmod a+x "${TOOLS_DIR}"/bin/"${name}"
  fi
done

echo "copying new binaries on a bucket..."
gsutil cp "$TOOLS_DIR/bin/*" "$KYMA_CLEANERS_BUCKET/"