#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

function performance_tests::run() {
  log::info "Running performance tests"

  kubectl create namespace perf-test
  find "$TEST_DIR/components/$TEST_COMPONENTS" -name "*.js" | xargs -I {} echo -n "--from-file={}" | xargs kubectl create configmap -n perf-test test-scripts
  kubectl get configmaps -oyaml -n perf-test test-scripts
  kubectl create -n perf-test -f "$SCRIPT_DIR/job.yaml"
  kubectl wait --for=condition=ready pod -n perf-test -l job-name=k6
  kubectl logs -n perf-test -f jobs/k6
}
