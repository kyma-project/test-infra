#!/bin/bash

# Expected vars:
# - INFLUXDB_FQDN: FQDN for influxDB
# - TESTS_DIR: Directory where all the tests are located

set -o pipefail

source $LIBS_DIR

# Setup
export K6_USER="$(cat /var/k6-details/k6admin)"
export K6_PASSWORD="$(cat /var/k6-details/k6admin_pass)"
export INFLUXDB="$(cat /var/k6-details/k6database)"


function runAll() {
    shout "Running the complete suite"
    for f in $(find "${TESTS_PATH}" -maxdepth 2 -type f -name *.js);
    do
        shout "Running File $f"
        $K6_CMD $f
    done
}

function runOne() {
    shout "Single file Mode"
    shout "Running following File $1"
    k6 run $1
}

if [[ "${1}" == "" ]]; then
    shoutFail "Please pass either 'all' or 'path to the test scrit' to run!!"
    exit 1
fi

if [[ "${TESTS_PATH}" == "" ]]; then
	shoutFail "TESTS_PATH not set!! Exiting"
	+  exit 1
fi

if [[ "${INFLUXDB_FQDN}" == "" ]]; then
    shoutFail "INFLUXDB not set exiting"
    exit 1
fi

K6_CMD="k6 run --out influxdb=http://${K6_USER}:${K6_PASSWORD}@${INFLUXDB_FQDN}/${INFLUXDB}"


if [[ "${1}" == "all" ]]; then
    runAll
else 
    runOne $1
fi  
