#!/bin/bash

set -eu


readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
mkdir -p bin

for D in `find cmd -mindepth 1 -maxdepth 1 -type d`
do
    name=$(basename ${D})
    echo ${name}
    cd ${CURRENT_DIR}/${D}
    go build -o ${CURRENT_DIR}/bin/${name}  -ldflags="-s -w"  main.go
    upx -q ${CURRENT_DIR}/bin/${name}
done

cd $CURRENT_DIR
