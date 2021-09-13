#!/usr/bin/env bash

readonly RECONCILER_SUFFIX="-reconciler"
readonly RECONCILER_NAMESPACE=reconciler
readonly RECONCILER_TIMEOUT=1200 # in secs
readonly RECONCILER_DELAY=10 # in secs


function reconciler::deploy() {
  # Deploy reconciler to cluster
  log::banner "Deploying Reconciler in the cluster"
  cd "${RECONCILER_SOURCES_DIR}"  || { echo "Failed to change dir to: ${RECONCILER_SOURCES_DIR}"; exit 1; }
  make deploy
}

# Checks whether reconciler is ready
function reconciler::wait_until_is_ready() {
  iterationsLeft=$(( RECONCILER_TIMEOUT/RECONCILER_DELAY ))
  while : ; do
    reconcilerCountDeploys=0
    readyCountDeploys=0
    for deploy in $(kubectl get deploy -n "${RECONCILER_NAMESPACE}" -ojsonpath='{ .items[*].metadata.name }'); do
      case $deploy in *"$RECONCILER_SUFFIX")
        reconcilerCountDeploys=$(( reconcilerCountDeploys+1 ))
        specReplicas=$(kubectl get deploy -n "${RECONCILER_NAMESPACE}" "${deploy}" -ojsonpath="{ .spec.replicas }")
        readyReplicas=$(kubectl get deploy -n "${RECONCILER_NAMESPACE}" "${deploy}" -ojsonpath="{ .status.readyReplicas }")
        if [[ specReplicas -eq readyReplicas ]]; then
            readyCountDeploys=$(( readyCountDeploys+1 ))
        fi
        ;;
      esac
    done

    if [ "${reconcilerCountDeploys}" -eq "${readyCountDeploys}" ] ; then
      log::info "Reconciler succesfully installed"
      break
    fi

    if [ "$RECONCILER_TIMEOUT" -ne 0 ] && [ "$iterationsLeft" -le 0 ]; then
      log::info "Timeout reached while waiting for reconciler to be ready. Exiting"
      exit 1
    fi
    sleep $RECONCILER_DELAY
    iterationsLeft=$(( iterationsLeft-1 ))
  done
}

# Waits until the test-pod deployment is in ready state
function reconciler::wait_until_test_pod_is_ready() {
  iterationsLeft=$(( RECONCILER_TIMEOUT/RECONCILER_DELAY ))
  while : ; do
    testPodStatus=$(kubectl get po -n reconciler test-pod -ojsonpath='{.status.containerStatuses[*].ready}')
    if [ "${testPodStatus}" = "true" ]; then
      log::info "Test pod is ready"
      break
    fi
    if [ "$RECONCILER_TIMEOUT" -ne 0 ] && [ "$iterationsLeft" -le 0 ]; then
      log::info "Timeout reached while initializing test pod. Exiting"
      exit 1
    fi
    log::info "Waiting for test pod to be ready..."
    sleep $RECONCILER_DELAY
    iterationsLeft=$(( iterationsLeft-1 ))
  done
}

# Initializes test pod which will send reconcile requests to reconciler
function reconciler::initialize_test_pod() {
  # Define KUBECONFIG env variable
  export KUBECONFIG="$HOME/.kube/config"

  if [[ ! $KYMA_UPGRADE_SOURCE ]]; then
    KYMA_UPGRADE_SOURCE="main"
  fi
  log::info "Kyma version to reconcile: ${KYMA_UPGRADE_SOURCE}"

  # move to reconciler directory
  cd "${RECONCILER_SOURCES_DIR}"  || { echo "Failed to change dir to: ${RECONCILER_SOURCES_DIR}"; exit 1; }

  # Create reconcile request payload with kubeconfig to the test-pod
  # shellcheck disable=SC2086
  kc="$(cat ${KUBECONFIG})"
  # shellcheck disable=SC2016
  jq --arg kubeconfig "${kc}" --arg version "${KYMA_UPGRADE_SOURCE}" '.kubeconfig = $kubeconfig | .kymaConfig.version = $version' ./scripts/e2e-test/template.json > body.json

  # Copy the reconcile request payload and kyma reconciliation script to the test-pod
  kubectl cp body.json reconciler/test-pod:/tmp
  kubectl cp  ./scripts/e2e-test/reconcile-kyma.sh reconciler/test-pod:/tmp
}

# Triggers reconciliation of Kyma
function reconciler::reconcile_kyma() {
  # Trigger Kyma reconciliation using reconciler
  log::banner "Reconcile Kyma in the same cluster until it is ready"
  kubectl exec -it -n reconciler test-pod -- sh -c ". /tmp/reconcile-kyma.sh"
  log::info "test-pod exited!"
}

# Deploy test pod
function reconciler::deploy_test_pod() {
  # Deploy a test pod
  log::banner "Deploying test-pod in the cluster"
  kubectl run -n reconciler --image=alpine:3.14.1 --restart=Never test-pod -- sh -c "sleep 36000"
}

function reconciler::disable_sidecar_injection_reconciler_ns() {
    log::info "Disabling sidecar injection for reconciler namespace"
    kubectl label namespace reconciler istio-injection=disabled --overwrite
}

function reconciler::pre_upgrade_test_fast_integration_kyma_1_24() {
    log::info "Running pre-upgrade Kyma Fast Integration tests"

    # Define KUBECONFIG env variable
    export KUBECONFIG="$HOME/.kube/config"

    pushd "${KYMA_PROJECT_DIR}/kyma-1.24/tests/fast-integration"
    make ci-pre-upgrade
    popd

    log::success "Tests completed"
}