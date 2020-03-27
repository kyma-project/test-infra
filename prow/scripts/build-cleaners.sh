#!/bin/bash

set -eu

if [ -z "${KYMA_CLEANERS_BUCKET}" ]; then
  echo "KYMA_CLEANERS_BUCKET is not set!"
  exit 1
fi

readonly TOOLS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}/../../development/tools" )" && pwd )"
mkdir -p bin

for D in `find cmd -mindepth 1 -maxdepth 1 -type d`
do
    name=$(basename "${D}")
    echo "building ${name}..."
    cd "${TOOLS_DIR}"/"${D}"
    go build -o "${TOOLS_DIR}"/bin/"${name}" -ldflags="-s -w" main.go
    # shellcheck disable=SC2086
    upx -q "${TOOLS_DIR}"/bin/${name}
    chmod a+x "${TOOLS_DIR}"/bin/${name}
done

echo "copying new binaries on a bucket..."
gsutil cp "bin/*" "$KYMA_CLEANERS_BUCKET/"