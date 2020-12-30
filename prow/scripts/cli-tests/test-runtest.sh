#!/usr/bin/env bash

shout "Running a simple test on Kyma"
clitests::assertRemoteCommand "sudo kyma test run dex-connection"

echo "Check if the test succeeds"
clitests::assertRemoteCommand \
    "sudo kyma test status -o json" \
    'Succeeded' \
    '.status.results[0].status'
