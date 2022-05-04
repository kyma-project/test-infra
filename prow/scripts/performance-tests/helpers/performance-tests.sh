#!/usr/bin/env bash

set -e

function performance_tests::run() {
  log::info "Running performance tests"

  kubectl create namespace perf-test
  find "$TEST_DIR/components/$TEST_COMPONENTS" -name "*.js" | xargs -I {} echo -n "--from-file={}" | xargs kubectl create configmap -n perf-test test-scripts
  kubectl get configmaps -oyaml -n perf-test test-scripts
}
