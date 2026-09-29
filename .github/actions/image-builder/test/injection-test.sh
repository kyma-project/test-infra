#!/usr/bin/env bash
# Regression tests for the input validation in action.yml prepare-* steps.
#
# Each step is replicated verbatim so this test stays in sync with the action
# and catches any revert of the env: isolation or allowlist validation.
#
# Injection tests start a nc listener and confirm that malicious payloads are
# blocked by the allowlist before the shell ever evaluates them.
#
# Requires: bash 4+, nc (netcat)

set -euo pipefail

PORT=19876
PASS=0
FAIL=0

# ── helpers ──────────────────────────────────────────────────────────────────

start_listener() {
  rm -f /tmp/nc-out
  nc -l -p "$PORT" > /tmp/nc-out 2>/dev/null &
  NC_PID=$!
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

assert_exit_zero() {
  local label="$1" exit_code="$2"
  if [[ "$exit_code" -eq 0 ]]; then
    echo "PASS [$label]: accepted valid input"
    PASS=$((PASS + 1))
  else
    echo "FAIL [$label]: rejected valid input (exit $exit_code)"
    FAIL=$((FAIL + 1))
  fi
}

assert_exit_nonzero() {
  local label="$1" exit_code="$2"
  if [[ "$exit_code" -ne 0 ]]; then
    echo "PASS [$label]: rejected invalid input (exit $exit_code)"
    PASS=$((PASS + 1))
  else
    echo "FAIL [$label]: accepted invalid input (exit 0)"
    FAIL=$((FAIL + 1))
  fi
}

# ── step implementations (verbatim from action.yml) ──────────────────────────

run_prepare_build_args() {
  local input="$1"
  local out
  out=$(mktemp)
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
    echo "build-args=$result" >> "$OUT_FILE"
  ' OUT_FILE="$out"
  local ec=$?
  rm -f "$out"
  return $ec
}

run_prepare_tags() {
  local input="$1"
  local out
  out=$(mktemp)
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
    echo "tags=$result" >> "$OUT_FILE"
  ' OUT_FILE="$out"
  local ec=$?
  rm -f "$out"
  return $ec
}

run_prepare_platforms() {
  local input="$1"
  local out
  out=$(mktemp)
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
    echo "platforms=$result" >> "$OUT_FILE"
  ' OUT_FILE="$out"
  local ec=$?
  rm -f "$out"
  return $ec
}

# ── build-args ────────────────────────────────────────────────────────────────

echo ""
echo "=== build-args: valid inputs ==="

run_prepare_build_args "VERSION=1.2.3"
assert_exit_zero "single arg" $?

run_prepare_build_args "$(printf 'VERSION=1.2.3\nREGISTRY=europe-docker.pkg.dev/kyma/prod')"
assert_exit_zero "multi-line args" $?

run_prepare_build_args ""
assert_exit_zero "empty input" $?

echo ""
echo "=== build-args: injection payloads (must be blocked before execution) ==="

start_listener
run_prepare_build_args "KEY=$(printf '$(curl -s http://localhost:%s/pwned)' "$PORT")" || true; ec=$?
assert_no_connection "build-args subshell \$()"
assert_exit_nonzero "build-args subshell \$() rejected" $ec

start_listener
run_prepare_build_args "KEY=$(printf '`curl -s http://localhost:%s/pwned`' "$PORT")" || true; ec=$?
assert_no_connection "build-args backtick"
assert_exit_nonzero "build-args backtick rejected" $ec

run_prepare_build_args "KEY=val;rm -rf /tmp/pwned" || true
assert_exit_nonzero "build-args semicolon rejected" $?

run_prepare_build_args "KEY=val|evil" || true
assert_exit_nonzero "build-args pipe rejected" $?

run_prepare_build_args "KEY=val space" || true
assert_exit_nonzero "build-args space rejected" $?

# ── tags ──────────────────────────────────────────────────────────────────────

echo ""
echo "=== tags: valid inputs ==="

run_prepare_tags "1.2.3"
assert_exit_zero "semver tag" $?

run_prepare_tags "PR-123"
assert_exit_zero "PR tag" $?

run_prepare_tags "name=1.2.3"
assert_exit_zero "name=value tag" $?

run_prepare_tags "$(printf '1.2.3\nlatest')"
assert_exit_zero "multi-line tags" $?

run_prepare_tags ""
assert_exit_zero "empty input" $?

echo ""
echo "=== tags: injection payloads (must be blocked before execution) ==="

start_listener
run_prepare_tags "$(printf '$(curl -s http://localhost:%s/pwned)' "$PORT")" || true; ec=$?
assert_no_connection "tags subshell \$()"
assert_exit_nonzero "tags subshell \$() rejected" $ec

start_listener
run_prepare_tags "$(printf '`curl -s http://localhost:%s/pwned`' "$PORT")" || true; ec=$?
assert_no_connection "tags backtick"
assert_exit_nonzero "tags backtick rejected" $ec

run_prepare_tags '";exit 1' || true
assert_exit_nonzero "tags quote-break rejected" $?

run_prepare_tags "tag;evil" || true
assert_exit_nonzero "tags semicolon rejected" $?

run_prepare_tags "tag|evil" || true
assert_exit_nonzero "tags pipe rejected" $?

# ── platforms ─────────────────────────────────────────────────────────────────

echo ""
echo "=== platforms: valid inputs ==="

run_prepare_platforms "linux/amd64"
assert_exit_zero "linux/amd64" $?

run_prepare_platforms "linux/arm64"
assert_exit_zero "linux/arm64" $?

run_prepare_platforms "$(printf 'linux/amd64\nlinux/arm64')"
assert_exit_zero "multi-line platforms" $?

echo ""
echo "=== platforms: injection payloads (must be blocked before execution) ==="

start_listener
run_prepare_platforms "$(printf '$(curl -s http://localhost:%s/pwned)' "$PORT")" || true; ec=$?
assert_no_connection "platforms subshell \$()"
assert_exit_nonzero "platforms subshell \$() rejected" $ec

start_listener
run_prepare_platforms "$(printf 'linux/amd64;curl -s http://localhost:%s/pwned' "$PORT")" || true; ec=$?
assert_no_connection "platforms semicolon injection"
assert_exit_nonzero "platforms semicolon rejected" $ec

run_prepare_platforms "windows/amd64" || true
assert_exit_nonzero "platforms unknown os rejected" $?

run_prepare_platforms "linux/amd64 extra" || true
assert_exit_nonzero "platforms trailing content rejected" $?

# ── summary ───────────────────────────────────────────────────────────────────

echo ""
echo "Results: $PASS passed, $FAIL failed"
[[ "$FAIL" -eq 0 ]]
