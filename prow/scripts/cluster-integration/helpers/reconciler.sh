#!/usr/bin/env bash

readonly RECONCILER_SUFFIX="-reconciler"
readonly RECONCILER_NAMESPACE=reconciler
readonly RECONCILER_TIMEOUT=1200 # in secs
readonly RECONCILER_DELAY=15 # in secs
readonly LOCAL_KUBECONFIG="$HOME/.kube/config"

function reconciler::delete_cluster_if_exists(){
  export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"
  for i in {1..5}
  do
    local name="${INPUT_CLUSTER_NAME}${i}"
    set +e
    existing_shoot=$(kubectl get shoot "${name}" -ojsonpath="{ .metadata.name }")
    if [ -n "${existing_shoot}" ]; then
      log::info "Cluster found and deleting '${name}'"
      gardener::deprovision_cluster \
            -p "${GARDENER_KYMA_PROW_PROJECT_NAME}" \
            -c "${name}" \
            -f "${GARDENER_KYMA_PROW_KUBECONFIG}" \
            -w "true"

      log::info "We wait 120s for Gardener Shoot to settle after cluster deletion"
      sleep 120
    else
      log::info "Cluster '${name}' does not exist"
    fi
    set -e
  done
}

function reconciler::provision_cluster() {
    export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"
    export DOMAIN_NAME="${INPUT_CLUSTER_NAME}"
    export DEFINITION_PATH="${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/reconciler/shoot-template.yaml"
    log::info "Creating cluster: ${INPUT_CLUSTER_NAME}"
    # create the cluster
    envsubst < "${DEFINITION_PATH}" | kubectl create -f -

    # wait for the cluster to be ready
    kubectl wait --for condition="ControlPlaneHealthy" --timeout=10m shoot "${INPUT_CLUSTER_NAME}"
    log::info "Cluster ${INPUT_CLUSTER_NAME} was created successfully"

}

function reconciler::deploy() {
  # Deploy reconciler to cluster
  log::banner "Deploying Reconciler in the cluster"
  cd "${CONTROL_PLANE_RECONCILER_DIR}"  || { echo "Failed to change dir to: ${CONTROL_PLANE_RECONCILER_DIR}"; exit 1; }
  make deploy-reconciler
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
      log::info "Reconciler is successfully installed"
      break
    fi

    if [ "$RECONCILER_TIMEOUT" -ne 0 ] && [ "$iterationsLeft" -le 0 ]; then
      log::info "Current state of pods in reconciler namespace"
      pods_info=$(kubectl get po -n "${RECONCILER_NAMESPACE}")
      log::info "${pods_info}"
      log::info "Timeout reached while waiting for reconciler to be ready. Exiting"
      exit 1
    fi
    sleep $RECONCILER_DELAY
    iterationsLeft=$(( iterationsLeft-1 ))
  done
}

# Waits until the test-pod is in ready state
function reconciler::wait_until_test_pod_is_ready() {
  iterationsLeft=$(( RECONCILER_TIMEOUT/RECONCILER_DELAY ))
  while : ; do
    testPodStatus=$(kubectl get po -n "${RECONCILER_NAMESPACE}" test-pod -ojsonpath='{.status.containerStatuses[?(@.name == "test-pod")].ready}')
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

# Waits until the test-pod is deleted completely
function reconciler::wait_until_test_pod_is_deleted() {
  iterationsLeft=$(( RECONCILER_TIMEOUT/RECONCILER_DELAY ))
  while : ; do
    testPodName=$(kubectl get po -n "${RECONCILER_NAMESPACE}" test-pod -ojsonpath='{.metadata.name}' --ignore-not-found)
    if [ -z "${testPodName}" ]; then
      log::info "Test pod is deleted"
      break
    fi
    if [ "$RECONCILER_TIMEOUT" -ne 0 ] && [ "$iterationsLeft" -le 0 ]; then
      log::info "Timeout reached while initializing test pod. Exiting"
      exit 1
    fi
    log::info "Waiting for test pod to be deleted..."
    sleep $RECONCILER_DELAY
    iterationsLeft=$(( iterationsLeft-1 ))
  done
}

# Initializes test pod which will send reconcile requests to reconciler
function reconciler::initialize_test_pod() {
  # Define KUBECONFIG env variable
  export KUBECONFIG="${LOCAL_KUBECONFIG}"

  if [[ ! $KYMA_UPGRADE_SOURCE ]]; then
    KYMA_UPGRADE_SOURCE="main"
  fi
  log::info "Kyma version to reconcile: ${KYMA_UPGRADE_SOURCE}"

  # move to reconciler directory
  cd "${CONTROL_PLANE_RECONCILER_DIR}"  || { echo "Failed to change dir to: ${CONTROL_PLANE_RECONCILER_DIR}"; exit 1; }

  # Create reconcile request payload with kubeconfig, domain, and version to the test-pod
  domain="$(kubectl get cm shoot-info -n kube-system -o jsonpath='{.data.domain}')"
  sed -i "s/example.com/$domain/" ./e2e-test/template.json

  # shellcheck disable=SC2086
  kc="$(cat ${KUBECONFIG})"
  # shellcheck disable=SC2016
  jq --arg kubeconfig "${kc}" --arg version "${KYMA_UPGRADE_SOURCE}" '.kubeconfig = $kubeconfig | .kymaConfig.version = $version' ./e2e-test/template.json > body.json

  # Copy the reconcile request payload and kyma reconciliation scripts to the test-pod
  kubectl cp body.json -c test-pod reconciler/test-pod:/tmp
  kubectl cp ./e2e-test/reconcile-kyma.sh -c test-pod reconciler/test-pod:/tmp
  kubectl cp ./e2e-test/get-reconcile-status.sh -c test-pod reconciler/test-pod:/tmp
  kubectl cp ./e2e-test/request-reconcile.sh -c test-pod reconciler/test-pod:/tmp
}

# Triggers reconciliation of Kyma and waits until reconciliation is in ready state
function reconciler::reconcile_kyma() {
  set +e
  # Trigger Kyma reconciliation using reconciler
  log::banner "Reconcile Kyma in the same cluster until it is ready"
  kubectl exec -it -n "${RECONCILER_NAMESPACE}" test-pod -c test-pod -- sh -c ". /tmp/reconcile-kyma.sh"
  if [[ $? -ne 0 ]]; then
      log::error "Failed to reconcile Kyma"
      kubectl logs -n "${RECONCILER_NAMESPACE}" -l app.kubernetes.io/name=mothership-reconciler -c mothership-reconciler --tail -1
      exit 1
  fi
  set -e
}

# Only triggers reconciliation of Kyma
function reconciler::trigger_kyma_reconcile() {
  # Trigger Kyma reconciliation using reconciler
  log::banner "Reconcile Kyma in the same cluster"
  kubectl exec -n "${RECONCILER_NAMESPACE}" test-pod -c test-pod -- sh -c ". /tmp/request-reconcile.sh"
  if [[ $? -ne 0 ]]; then
      log::error "Failed to trigger reconciliation"
      exit 1
  fi
}

# Waits until Kyma reconciliation is in ready state
function reconciler::wait_until_kyma_reconciled() {
  iterationsLeft=$(( RECONCILER_TIMEOUT/RECONCILER_DELAY ))
  while : ; do
    status=$(kubectl exec -n "${RECONCILER_NAMESPACE}" test-pod -c test-pod -- sh -c ". /tmp/get-reconcile-status.sh" | xargs)
    if [ "${status}" = "ready" ]; then
      log::info "Kyma is reconciled"
      break
    fi

    if [ "${status}" = "error" ]; then
      log::error "Failed to reconcile Kyma. Exiting"
      kubectl logs -n "${RECONCILER_NAMESPACE}" -l app.kubernetes.io/name=mothership-reconciler -c mothership-reconciler --tail -1
      exit 1
    fi

    if [ "$RECONCILER_TIMEOUT" -ne 0 ] && [ "$iterationsLeft" -le 0 ]; then
      log::error "Timeout reached on Kyma reconciliation. Exiting"
      kubectl logs -n "${RECONCILER_NAMESPACE}" -l app.kubernetes.io/name=mothership-reconciler -c mothership-reconciler --tail -1
      exit 1
    fi

    sleep $RECONCILER_DELAY
    log::info "Waiting for reconciliation to finish, current status: ${status} ...."
    iterationsLeft=$(( iterationsLeft-1 ))
  done
}

# Deploy test pod
function reconciler::deploy_test_pod() {
  # Deploy a test pod
  log::banner "Deploying test-pod in the cluster"
  test_pod_name=$(kubectl get po test-pod -n "${RECONCILER_NAMESPACE}" -ojsonpath="{ .metadata.name }" --ignore-not-found)
  if [ -n "${test_pod_name}" ]; then
    log::info "Found existing pod: test-pod"
    kubectl delete po test-pod -n "${RECONCILER_NAMESPACE}"
    reconciler::wait_until_test_pod_is_deleted
  fi
  kubectl run -n "${RECONCILER_NAMESPACE}" --image=alpine:3.14.1 --restart=Never test-pod -- sh -c "sleep 36000"
}

function reconciler::disable_sidecar_injection_reconciler_ns() {
    log::info "Disabling sidecar injection for reconciler namespace"
    kubectl label namespace "${RECONCILER_NAMESPACE}" istio-injection=disabled --overwrite
}

# Connect to Gardener cluster
function reconciler::connect_to_gardener_cluster() {
    export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"
}

# Connect to reconciler long running cluster
function reconciler::connect_to_shoot_cluster() {
  reconciler::connect_to_gardener_cluster
  local shoot_kubeconfig="/tmp/shoot-kubeconfig.yaml"
  kubectl get secret "${INPUT_CLUSTER_NAME}.kubeconfig"  -ogo-template="{{ .data.kubeconfig | base64decode }}" > "${shoot_kubeconfig}"
  cat "${shoot_kubeconfig}" > "${LOCAL_KUBECONFIG}"
  export KUBECONFIG="${shoot_kubeconfig}"
}

# Break Kyma to test reconciler repair mechanism
function reconciler::break_kyma() {
  log::banner "Delete all deployments from kyma-system ns"
  kubectl delete deploy -n kyma-system --all
}