#!/usr/bin/env bash

JUNIT_SUITE_START_TIME=0
JUNIT_SUITE_NAME=""

JUNIT_TOTAL_TESTS_COUNT=0
JUNIT_FAILED_TESTS_COUNT=0
JUNIT_SKIPPED_TESTS_COUNT=0

JUNIT_TEST_NAME=""
JUNIT_TEST_START_TIME=0

# junit::suite_init initializes the JUnit test suite
#
# Arguments:
#   $1 - Suite name
function junit::suite_init {
    JUNIT_SUITE_START_TIME=$(date +%s)
    JUNIT_SUITE_NAME="${1}"
    readonly JUNIT_SUITE_NAME JUNIT_SUITE_START_TIME
    rm -f "$(junit::file_path_prefix)"_*
}

# junit::suite_save saves the test suite to the file
function junit::suite_save {
    local -r duration=$(($(date +%s)-JUNIT_SUITE_START_TIME))
    echo "<testsuite failures=\"${JUNIT_FAILED_TESTS_COUNT}\" name=\"${JUNIT_SUITE_NAME}\" skipped=\"${JUNIT_SKIPPED_TESTS_COUNT}\" tests=\"${JUNIT_TOTAL_TESTS_COUNT}\" time=\"${duration}\">
$(cat "$(junit::suite_filename)")
</testsuite>" > "$(junit::suite_filename)"
}

# (private) junit::file_path_prefix returns path prefix to the JUnit file
function junit::file_path_prefix {
    echo "${ARTIFACTS_DIR}/junit_${JUNIT_SUITE_NAME}"
}

# (private) junit::file_path_prefix returns path to the JUnit suite results
function junit::suite_filename {
    echo "$(junit::file_path_prefix)_suite.xml"
}

# (private) junit::file_path_prefix returns path to the JUnit step output
function junit::test_output_filename {
    echo "$(junit::file_path_prefix)_${JUNIT_TEST_NAME}_output.log"
}

# junit::test_start initializes the JUnit test
#
# Arguments:
#   $1 - Test case name
function junit::test_start {
    JUNIT_TOTAL_TESTS_COUNT=$((++JUNIT_TOTAL_TESTS_COUNT))
    JUNIT_TEST_NAME="${1}"
    JUNIT_TEST_START_TIME=$(date +%s)
    rm -f "$(junit::test_output_filename)"
    echo "=== RUN: ${JUNIT_TEST_NAME} 🤔"
}

# junit::test_output writes JUnit test output to the file
function junit::test_output {
    while read -r data; do
        echo "${data}" | tee -a "$(junit::test_output_filename)"
    done
}

# junit::test_pass marks current JUnit test as passed
function junit::test_pass {
    if [[ -z ${JUNIT_TEST_NAME} ]]; then
        return 0
    fi
    local -r duration=$(($(date +%s)-JUNIT_TEST_START_TIME))
    echo "--- PASS: ${JUNIT_TEST_NAME} (${duration}s) 😍"
    echo "        <testcase name=\"${JUNIT_TEST_NAME}\" time=\"${duration}\"></testcase>" >> "$(junit::suite_filename)"
    JUNIT_TEST_NAME=""
}

# junit::test_pass marks current JUnit test as failed
function junit::test_fail {
    if [[ -z ${JUNIT_TEST_NAME} ]]; then
        return 0
    fi
    local -r duration=$(($(date +%s)-JUNIT_TEST_START_TIME))
    JUNIT_FAILED_TESTS_COUNT=$((++JUNIT_FAILED_TESTS_COUNT))
    echo "--- FAIL: ${JUNIT_TEST_NAME} (${duration}s) 💩"
    echo "        <testcase name=\"${JUNIT_TEST_NAME}\" time=\"${duration}\"><failure><![CDATA[$(< "$(junit::test_output_filename)")]]></failure></testcase>" >> "$(junit::suite_filename)"
    JUNIT_TEST_NAME=""
    return 1
}

# junit::test_pass marks current JUnit test as skipped
function junit::test_skip {
    if [[ -z ${JUNIT_TEST_NAME} ]]; then
        return 0
    fi
    local -r duration=$(($(date +%s)-JUNIT_TEST_START_TIME))
    JUNIT_SKIPPED_TESTS_COUNT=$((++JUNIT_SKIPPED_TESTS_COUNT))
    echo "--- SKIP: ${JUNIT_TEST_NAME} (${duration}s) 🙈"
    echo "        <testcase name=\"${JUNIT_TEST_NAME}\" time=\"${duration}\"><skipped><![CDATA[${1}]]></skipped></testcase>" >> "$(junit::suite_filename)"
    JUNIT_TEST_NAME=""
}
