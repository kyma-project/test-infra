#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"
TESTDIR="${LIBDIR}/../cli-tests"

#
# Get the path to the file containing the test steps.
# @param Name of the test-stragey (will be mapped to test file).
#
clitests::getTestFile() {
    local testSuite=$1

    echo "${TESTDIR}/testsuite-${testSuite}.sh"
}

#
# Check if a test strategy exists.
# @param Name of the test strategy.
#
clitests::testSuiteExists() {
    local testSuite=$1

    clitests::assertVarNotEmpty testSuite

    test -f "$(clitests::getTestFile $testSuite)"
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
        shoutFail "The variable '$var' was not defined or is empty"
    else
        shoutFail "$msg"
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
    local cliVersion=$4

    clitests::assertVarNotEmpty testSuite
    clitests::assertVarNotEmpty gcpZone
    clitests::assertVarNotEmpty gcpHost
    clitests::assertVarNotEmpty cliVersion

    shout "Executing test suite '$testSuite'"
    export ZONE=$gcpZone
    export HOST=$gcpHost
    export SOURCE=$cliVersion

    local testFile
    testFile=$(clitests::getTestFile "$testSuite")
    # shellcheck disable=SC1090
    source "$testFile"
}