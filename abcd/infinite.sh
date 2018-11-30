#!/usr/bin/env bash

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPT_DIR}/../prow/scripts/library.sh"
init

for i in {1..600}
do
    echo "Building ${i}"
    docker build -t abcd .
    docker rmi abcd:latest
done

for i in {1..300}
do
    echo "Sleeping ${i}"
    sleep 1
done 