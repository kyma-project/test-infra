#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"
TESTDIR="${LIBDIR}/../cli-tests"
# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"

#
# Get the path to the file containing the test steps.
# @param Name of the test file.
#
clitests::getTestFile() {
    local testFile=$1

    echo "${TESTDIR}/${testFile}.sh"
}

#
# Check if a test strategy exists.
# @param Name of the test strategy.
#
clitests::testSuiteExists() {
    local testSuite=$1

    clitests::assertVarNotEmpty testSuite

    local testFile
    testFile=$(clitests::getTestFile "$testSuite")
    test -f "$testFile"
    return $?
}

#
# Ensure that a variable is defined.
# @param Name of the variable (with $).
# @param Optional: Custom error message shown if variable is empty.
#
clitests::assertVarNotEmpty() {
  local var=$1
  local msg=$2

  if [ -z "${!var}" ]; then
    if [ -z "$msg" ]; then
        log::error "The variable '$var' was not defined or is empty"
    else
        log::error "$msg"
    fi
    exit 1
  fi
}


#
# Execute a test.
# @param Name of the test strategy.
# @param GCP Zone (where the VM is running).
# @param Name of the VM.
# @param The Kyma CLI version.
#
clitests::execute() {
    local testSuite=$1
    local gcpZone=$2
    local gcpHost=$3
    # optional:
    local cliVersion=$4

    clitests::assertVarNotEmpty testSuite
    clitests::assertVarNotEmpty gcpZone
    clitests::assertVarNotEmpty gcpHost

    log::info "Executing test suite '$testSuite'"
    export ZONE=$gcpZone
    export HOST=$gcpHost
    export SOURCE=$cliVersion

    local testFile
    testFile=$(clitests::getTestFile "$testSuite")
    # shellcheck disable=SC1090
    source "$testFile"
}

#
# Run a command remote on GCP host.
# @param Command to execute.
# @param Optional: GCP Zone (default is $ZONE)
# @param Optional: GCP Host (default is $HOST)
#
clitests::assertRemoteCommand() {
    local cmd="$1"
    local assert="$2"
    local jqFilter="$3"
    local zone="${4:-$ZONE}"
    local host="${5:-$HOST}"

    # config values
    local interval=15
    local retries=10

    date
    local output
    for loopCount in $(seq 1 $retries); do
        output=$(gcloud compute ssh --ssh-key-file="${SSH_KEY_FILE_PATH:-/root/.ssh/user/google_compute_engine}" --quiet --zone="${zone}" "${host}" -- "$cmd")
        cmdExitCode=$?

        # check return code and apply assertion (if defined)
        if [ $cmdExitCode -eq 0 ]; then
            if [ -n "$assert" ]; then
                # apply JQ filter (if defined)
                if [ -n "$jqFilter" ]; then
                    local jqOutput
                    jqOutput=$(echo "$output" | jq -r "$jqFilter")
                    jqExitCode=$?

                    # show JQ failures
                    if [ $jqExitCode -eq 0 ]; then
                        output="$jqOutput"
                    else
                        echo "JQFilter '${jqFilter}' failed with exit code '${jqExitCode}' (JSON input: ${output})"
                    fi
                fi
                # assert output
                if [ "$output" = "$assert" ]; then
                    break
                else
                    echo "Assertion failed: expected '$assert' but got '$output'"
                fi
            else
                # no assertion required
                break
            fi
        fi

        # abort if max-retires are reached
        if [ "$loopCount" -ge $retries ]; then
            echo "Failed after $loopCount attempts."
            exit 1
        fi

        echo "Retrying in $interval seconds.."
        sleep $interval
    done;
}
