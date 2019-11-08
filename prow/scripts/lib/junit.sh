#!/usr/bin/env bash

JUNIT_SUITE_START_TIME=0
JUNIT_SUITE_NAME=""

JUNIT_TOTAL_TESTS_COUNT=0
JUNIT_FAILED_TESTS_COUNT=0
JUNIT_SKIPPED_TESTS_COUNT=0

JUNIT_TEST_NAME=""
JUNIT_TEST_START_TIME=0

function junit::suite_init {
    JUNIT_SUITE_START_TIME=$(date +%s)
    JUNIT_SUITE_NAME="${1}"
    readonly JUNIT_SUITE_NAME JUNIT_SUITE_START_TIME
    rm -f "$(junit::file_path_prefix)"_*
}

function junit::suite_save {
    local -r duration=$(($(date +%s)-JUNIT_SUITE_START_TIME))
    echo "<testsuite failures=\"${JUNIT_FAILED_TESTS_COUNT}\" name=\"${JUNIT_SUITE_NAME}\" skipped=\"${JUNIT_SKIPPED_TESTS_COUNT}\" tests=\"${JUNIT_TOTAL_TESTS_COUNT}\" time=\"${duration}\">
$(cat "$(junit::suite_filename)")
</testsuite>" > "$(junit::suite_filename)"
}

function junit::file_path_prefix {
    echo "${ARTIFACTS_DIR}/junit_${JUNIT_SUITE_NAME}"
}

function junit::suite_filename {
    echo "$(junit::file_path_prefix)_suite.xml"
}

function junit::test_output_filename {
    echo "$(junit::file_path_prefix)_${JUNIT_TEST_NAME}_output.log"
}

function junit::test_start {
    JUNIT_TOTAL_TESTS_COUNT=$((++JUNIT_TOTAL_TESTS_COUNT))
    JUNIT_TEST_NAME="${1}"
    JUNIT_TEST_START_TIME=$(date +%s)
    rm -f "$(junit::test_output_filename)"
    echo "=== RUN: ${JUNIT_TEST_NAME} 🤔"
}

function junit::test_output {
    while read -r data; do
        echo "${data}" | tee -a "$(junit::test_output_filename)"
    done
}

function junit::test_pass {
    local -r duration=$(($(date +%s)-JUNIT_TEST_START_TIME))
    echo "--- PASS: ${JUNIT_TEST_NAME} (${duration}s) 😍"
    echo "        <testcase name=\"${JUNIT_TEST_NAME}\" time=\"${duration}\"></testcase>" >> "$(junit::suite_filename)"
}

function junit::test_fail {
    local -r duration=$(($(date +%s)-JUNIT_TEST_START_TIME))
    JUNIT_FAILED_TESTS_COUNT=$((++JUNIT_FAILED_TESTS_COUNT))
    echo "--- FAIL: ${JUNIT_TEST_NAME} (${duration}s) 💩"
    echo "        <testcase name=\"${JUNIT_TEST_NAME}\" time=\"${duration}\"><failure>$(< "$(junit::test_output_filename)" tr -cd '\11\12\15\40-\176' | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g; s/"/\&quot;/g; s/'"'"'/\&#39;/g')</failure></testcase>" >> "$(junit::suite_filename)"
    return 1
}

function junit::test_skip {
    local -r duration=$(($(date +%s)-JUNIT_TEST_START_TIME))
    JUNIT_SKIPPED_TESTS_COUNT=$((++JUNIT_SKIPPED_TESTS_COUNT))
    echo "--- SKIP: ${JUNIT_TEST_NAME} (${duration}s) 🙈"
    echo "        <testcase name=\"${JUNIT_TEST_NAME}\" time=\"${duration}\"><skipped>${1}</skipped></testcase>" >> "$(junit::suite_filename)"
}
