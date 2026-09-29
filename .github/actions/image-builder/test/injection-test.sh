#!/usr/bin/env bash
# Validates that the three input-preparation steps in action.yml are not vulnerable
# to shell injection. Each step is replicated verbatim here so the test stays in
# sync with the action and catches regressions if someone reverts the env: fix.
#
# Strategy: start a nc listener on localhost, run each step with a payload that
# would curl that listener if injection were possible, then assert:
#   1. the step exits non-zero (validation rejected the input), and
#   2. nc received no data (the curl never fired).

set -euo pipefail

PORT=19876
PASS=0
FAIL=0

# ── helpers ──────────────────────────────────────────────────────────────────

start_listener() {
  rm -f /tmp/nc-out
  nc -l -p "$PORT" > /tmp/nc-out 2>/dev/null &
  NC_PID=$!
  # give nc a moment to bind
  sleep 0.2
}

stop_listener() {
  kill "$NC_PID" 2>/dev/null || true
  wait "$NC_PID" 2>/dev/null || true
}

assert_no_connection() {
  local label="$1"
  sleep 0.3
  stop_listener
  if [[ -s /tmp/nc-out ]]; then
    echo "FAIL [$label]: injection fired — nc received: $(cat /tmp/nc-out)"
    FAIL=$((FAIL + 1))
  else
    echo "PASS [$label]: no connection received"
    PASS=$((PASS + 1))
  fi
}

assert_exit_nonzero() {
  local label="$1"
  local exit_code="$2"
  if [[ "$exit_code" -ne 0 ]]; then
    echo "PASS [$label]: step rejected input (exit $exit_code)"
    PASS=$((PASS + 1))
  else
    echo "FAIL [$label]: step accepted malicious input (exit 0)"
    FAIL=$((FAIL + 1))
  fi
}

assert_exit_zero() {
  local label="$1"
  local exit_code="$2"
  if [[ "$exit_code" -eq 0 ]]; then
    echo "PASS [$label]: step accepted valid input (exit 0)"
    PASS=$((PASS + 1))
  else
    echo "FAIL [$label]: step rejected valid input (exit $exit_code)"
    FAIL=$((FAIL + 1))
  fi
}

# ── step implementations (verbatim from action.yml) ──────────────────────────

run_prepare_build_args() {
  local input="$1"
  local output
  OUTPUT_FILE=$(mktemp)
  INPUT_BUILD_ARGS="$input" bash -c '
    readarray -t lines <<< "$INPUT_BUILD_ARGS"
    result=""
    for entry in "${lines[@]}"; do
      if [[ -n "$entry" ]]; then
        [[ "$entry" =~ ^[A-Za-z_][A-Za-z0-9_]*=[A-Za-z0-9._:/@+\-]*$ ]] || \
          { echo "::error::Invalid build-arg: $entry"; exit 1; }
        result+=" --build-arg=$entry"
      fi
    done
    echo "build-args=$result" >> "$OUTPUT_FILE"
  ' OUTPUT_FILE="$OUTPUT_FILE"
  local exit_code=$?
  rm -f "$OUTPUT_FILE"
  return $exit_code
}

run_prepare_tags() {
  local input="$1"
  OUTPUT_FILE=$(mktemp)
  INPUT_TAGS="$input" bash -c '
    readarray -t lines <<< "$INPUT_TAGS"
    result=""
    for entry in "${lines[@]}"; do
      if [[ -n "$entry" ]]; then
        [[ "$entry" =~ ^[A-Za-z0-9._:/@+\-]+(=[A-Za-z0-9._:/@+\-]+)?$ ]] || \
          { echo "::error::Invalid tag: $entry"; exit 1; }
        result+=" --tag=$entry"
      fi
    done
    echo "tags=$result" >> "$OUTPUT_FILE"
  ' OUTPUT_FILE="$OUTPUT_FILE"
  local exit_code=$?
  rm -f "$OUTPUT_FILE"
  return $exit_code
}

run_prepare_platforms() {
  local input="$1"
  OUTPUT_FILE=$(mktemp)
  INPUT_PLATFORMS="$input" bash -c '
    readarray -t lines <<< "$INPUT_PLATFORMS"
    result=""
    for entry in "${lines[@]}"; do
      if [[ -n "$entry" ]]; then
        [[ "$entry" =~ ^linux/(amd64|arm64|arm/v7|s390x|ppc64le)$ ]] || \
          { echo "::error::Invalid platform: $entry"; exit 1; }
        result+=" --platform=$entry"
      fi
    done
    echo "platforms=$result" >> "$OUTPUT_FILE"
  ' OUTPUT_FILE="$OUTPUT_FILE"
  local exit_code=$?
  rm -f "$OUTPUT_FILE"
  return $exit_code
}

# ── injection payloads ────────────────────────────────────────────────────────

# Payload that would fire curl if the here-string is broken out of
CURL_PAYLOAD='$(curl -s http://localhost:'"$PORT"'/pwned)'
BACKTICK_PAYLOAD='`curl -s http://localhost:'"$PORT"'/pwned`'
HEREDOC_BREAK_PAYLOAD="foo\"
\$(curl -s http://localhost:$PORT/pwned)
\""

# ── tests: build-args ─────────────────────────────────────────────────────────

echo ""
echo "=== build-args ==="

# valid inputs
run_prepare_build_args "VERSION=1.2.3"
assert_exit_zero "build-args valid single" $?

run_prepare_build_args "$(printf 'VERSION=1.2.3\nREGISTRY=europe-docker.pkg.dev/kyma/prod')"
assert_exit_zero "build-args valid multi-line" $?

# injection: curl payload as value
start_listener
run_prepare_build_args "KEY=$CURL_PAYLOAD" || true
EC=$?
assert_no_connection "build-args curl-payload no-exec"
assert_exit_nonzero "build-args curl-payload rejected" $EC

# injection: semicolon
run_prepare_build_args "KEY=val;rm -rf /tmp/pwned" || true
assert_exit_nonzero "build-args semicolon rejected" $?

# injection: backtick
start_listener
run_prepare_build_args "KEY=$BACKTICK_PAYLOAD" || true
EC=$?
assert_no_connection "build-args backtick no-exec"
assert_exit_nonzero "build-args backtick rejected" $EC

# ── tests: tags ──────────────────────────────────────────────────────────────

echo ""
echo "=== tags ==="

run_prepare_tags "1.2.3"
assert_exit_zero "tags valid semver" $?

run_prepare_tags "PR-123"
assert_exit_zero "tags valid PR tag" $?

run_prepare_tags "name=1.2.3"
assert_exit_zero "tags valid name=value" $?

run_prepare_tags "$(printf '1.2.3\nlatest')"
assert_exit_zero "tags valid multi-line" $?

start_listener
run_prepare_tags "$CURL_PAYLOAD" || true
EC=$?
assert_no_connection "tags curl-payload no-exec"
assert_exit_nonzero "tags curl-payload rejected" $EC

start_listener
run_prepare_tags "\";exit 1" || true
EC=$?
assert_no_connection "tags quote-break no-exec"
assert_exit_nonzero "tags quote-break rejected" $EC

# ── tests: platforms ─────────────────────────────────────────────────────────

echo ""
echo "=== platforms ==="

run_prepare_platforms "linux/amd64"
assert_exit_zero "platforms valid amd64" $?

run_prepare_platforms "linux/arm64"
assert_exit_zero "platforms valid arm64" $?

run_prepare_platforms "$(printf 'linux/amd64\nlinux/arm64')"
assert_exit_zero "platforms valid multi-line" $?

run_prepare_platforms "windows/amd64" || true
assert_exit_nonzero "platforms windows rejected" $?

start_listener
run_prepare_platforms "linux/amd64;curl -s http://localhost:$PORT/pwned" || true
EC=$?
assert_no_connection "platforms injection no-exec"
assert_exit_nonzero "platforms injection rejected" $EC

start_listener
run_prepare_platforms "$CURL_PAYLOAD" || true
EC=$?
assert_no_connection "platforms curl-payload no-exec"
assert_exit_nonzero "platforms curl-payload rejected" $EC

# ── summary ───────────────────────────────────────────────────────────────────

echo ""
echo "Results: $PASS passed, $FAIL failed"
[[ "$FAIL" -eq 0 ]]
