#!/usr/bin/env bash

set -e

function performance_tests::run() {
  log::info "Running performance tests"

  ls "$TEST_DIR/components/$TEST_COMPONENTS/*.js" | xargs -I {} echo -n "--from-file={}" | xargs kubectl create configmap -n perf-test test-scripts
  kubectl get configmaps -oyaml -n perf-test test-scripts
}
