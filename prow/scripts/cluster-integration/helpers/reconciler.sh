#!/usr/bin/env bash

readonly RECONCILER_SUFFIX="-reconciler"
readonly RECONCILER_NAMESPACE=reconciler
readonly RECONCILER_TIMEOUT=1200 # in secs
readonly RECONCILER_DELAY=15 # in secs
readonly LOCAL_KUBECONFIG="$HOME/.kube/config"
readonly MOTHERSHIP_RECONCILER_VALUES_FILE="../../resources/kcp/charts/mothership-reconciler/values.yaml"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

function reconciler::export_nightly_cluster_name(){
  log::info "Export nightly cluster name"
  # shellcheck disable=SC2046
  # shellcheck disable=SC2005
  day=$(echo $(date +%a) | tr "[:upper:]" "[:lower:]" | cut -c1-2)
  export INPUT_CLUSTER_NAME="${INPUT_CLUSTER_NAME}-${day}"
}

function reconciler::delete_cluster_if_exists(){
  log::info "Delete cluster with reconciler if exists"
  export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"
  for i in mo tu we th fr sa su
  do
    local name="${INPUT_CLUSTER_NAME}-${i}"
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

# reconciler::reprovision_cluster will generate new cluster name
# and start provisioning again
function reconciler::reprovision_cluster() {
    log::info "cluster provisioning failed, trying provision new cluster"
    log::info "cleaning damaged cluster first"

    gardener::deprovision_cluster \
      -p "${GARDENER_KYMA_PROW_PROJECT_NAME}" \
      -c "${INPUT_CLUSTER_NAME}" \
      -f "${GARDENER_KYMA_PROW_KUBECONFIG}"
    
    log::info "building new cluster name"

    utils::generate_commonName -n "${COMMON_NAME_PREFIX}"
    COMMON_NAME=${utils_generate_commonName_return_commonName:?}
    export COMMON_NAME
    INPUT_CLUSTER_NAME="${COMMON_NAME}"
    export INPUT_CLUSTER_NAME
    reconciler::provision_cluster
}

function reconciler::provision_cluster() {
    log::info "Provision reconciler cluster"
    export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"
    export DOMAIN_NAME="${INPUT_CLUSTER_NAME}"
    export DEFINITION_PATH="${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/reconciler/shoot-template.yaml"
    log::info "Creating cluster: ${INPUT_CLUSTER_NAME}"

    # catch cluster provisioning errors and try provision new one
    trap reconciler::reprovision_cluster ERR

    set +e
    # create the cluster
    envsubst < "${DEFINITION_PATH}" | kubectl create -f -
    set -e

    # wait for the cluster to be ready
    kubectl wait --for condition="ControlPlaneHealthy" --timeout=20m shoot "${INPUT_CLUSTER_NAME}"
    log::info "Cluster ${INPUT_CLUSTER_NAME} was created successfully"

    # disable trap for cluster provisioning errors to not call it for later errors
    trap - ERR
}

function reconciler::deploy() {
  # Deploy reconciler to cluster
  log::banner "Deploying Reconciler in the cluster"
  cd "${CONTROL_PLANE_RECONCILER_DIR}"  || { echo "Failed to change dir to: ${CONTROL_PLANE_RECONCILER_DIR}"; exit 1; }
  make deploy-reconciler
}

# Checks whether reconciler is ready
function reconciler::wait_until_is_ready() {
  log::info "Wait until reconciler is in ready state"
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
  log::info "Wait until test-pod is in ready state"
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
  log::info "Set up test pod environment"
  # Define KUBECONFIG env variable
  export KUBECONFIG="${LOCAL_KUBECONFIG}"

  if [[ ! $KYMA_UPGRADE_SOURCE ]]; then
    KYMA_UPGRADE_SOURCE="main"
  fi
  log::info "Kyma version to reconcile: ${KYMA_UPGRADE_SOURCE}"

  # move to reconciler directory
  cd "${CONTROL_PLANE_RECONCILER_DIR}"  || { echo "Failed to change dir to: ${CONTROL_PLANE_RECONCILER_DIR}"; exit 1; }
  if [ ! -f "$MOTHERSHIP_RECONCILER_VALUES_FILE" ]; then
    log::error "Mothership reconciler values file not found! ($MOTHERSHIP_RECONCILER_VALUES_FILE)"
    exit 1
  fi
  echo "************* Current Reconciler Image To Be Used **************"
  cat < "$MOTHERSHIP_RECONCILER_VALUES_FILE" | grep -o 'mothership_reconciler.*\"'
  echo "****************************************************************"
  # Create reconcile request payload with kubeconfig, domain, and version to the test-pod
  domain="$(kubectl get cm shoot-info -n kube-system -o jsonpath='{.data.domain}')"

  # shellcheck disable=SC2086
  kc="$(cat ${KUBECONFIG})"

  local tplFile="./e2e-test/template-kyma-main.json"
  if [[ "$KYMA_UPGRADE_SOURCE" =~ ^2\.0\.[0-9]+$ ]] ; then
    tplFile="./e2e-test/template-kyma-2-0-x.json"
  elif [[ "$KYMA_UPGRADE_SOURCE" =~ ^2\.[1-3]\.[0-9]+$ ]] ; then
    tplFile="./e2e-test/template-kyma-2-1-x.json"
  elif [[ "$KYMA_UPGRADE_SOURCE" =~ ^2\.[4-5]\.[0-9]+$ ]] ; then
    tplFile="./e2e-test/template-kyma-2-4-x.json"
  fi

  log::info "Calling reconciler by using JSON template '$tplFile' as payload"

  sed -i "s/example.com/$domain/" "$tplFile"
  # shellcheck disable=SC2016
  jq --arg kubeconfig "${kc}" --arg version "${KYMA_UPGRADE_SOURCE}" '.kubeconfig = $kubeconfig | .kymaConfig.version = $version' "$tplFile" > body.json

  # Copy the reconcile request payload and kyma reconciliation scripts to the test-pod
  kubectl cp body.json -c test-pod reconciler/test-pod:/tmp
  kubectl cp ./e2e-test/reconcile-kyma.sh -c test-pod reconciler/test-pod:/tmp
  kubectl cp ./e2e-test/get-reconcile-status.sh -c test-pod reconciler/test-pod:/tmp
  kubectl cp ./e2e-test/request-reconcile.sh -c test-pod reconciler/test-pod:/tmp
}

# Only triggers reconciliation of Kyma
function reconciler::trigger_kyma_reconcile() {
  # Trigger Kyma reconciliation using reconciler
  log::info "Trigger the reconciliation through test pod"
  log::banner "Reconcile Kyma in the same cluster"
  kubectl exec -n "${RECONCILER_NAMESPACE}" test-pod -c test-pod -- sh -c ". /tmp/request-reconcile.sh"
  if [[ $? -ne 0 ]]; then
      log::error "Failed to trigger reconciliation"
      exit 1
  fi
}

# Waits until Kyma reconciliation is in ready state
function reconciler::wait_until_kyma_reconciled() {
  log::info "Wait until reconciliation is complete"
  iterationsLeft=$(( RECONCILER_TIMEOUT/RECONCILER_DELAY ))
  while : ; do
    status=$(kubectl exec -n "${RECONCILER_NAMESPACE}" test-pod -c test-pod -- sh -c ". /tmp/get-reconcile-status.sh" | xargs || true)

    if [ -z "${status}" ]; then
      log::info "Failed to retrieve reconciliation status. Retrying previous call in debug mode"
      kubectl exec -v=8 -n "${RECONCILER_NAMESPACE}" test-pod -c test-pod -- sh -c ". /tmp/get-reconcile-status.sh" || true
      status=$(kubectl exec -n "${RECONCILER_NAMESPACE}" test-pod -c test-pod -- sh -c ". /tmp/get-reconcile-status.sh" | xargs || true)
    fi

    if [ "${status}" = "ready" ]; then
      log::info "Kyma is reconciled"
      break
    fi

    if [ "${status}" = "error" ]; then
      log::error "Failed to reconcile Kyma. Exiting"
      kubectl logs -n "${RECONCILER_NAMESPACE}" -l app.kubernetes.io/name=mothership-reconciler -c mothership-reconciler --tail -1
      exit 1
    fi

    if [ -z "${status}" ]; then
      log::info "Failed to retrieve reconciliation status. Checking if API server is reachable by asking for its version"
      kubectl version -v=8
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
  log::info "Deploy test-pod in the cluster which will trigger reconciliation"
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

# Export shoot cluster kubeconfig to ENV
function reconciler::export_shoot_cluster_kubeconfig() {
  log::info "Export shoot cluster kubeconfig to ENV"
  export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"
  local shoot_kubeconfig="/tmp/shoot-kubeconfig.yaml"
  cat <<EOF | kubectl replace -f - --raw "/apis/core.gardener.cloud/v1beta1/namespaces/garden-kyma-prow/shoots/${INPUT_CLUSTER_NAME}/adminkubeconfig" | jq -r ".status.kubeconfig" | base64 -d > "${shoot_kubeconfig}"
{
    "apiVersion": "authentication.gardener.cloud/v1alpha1",
    "kind": "AdminKubeconfigRequest",
    "spec": {
        "expirationSeconds": 10800
    }
}
EOF
  cat "${shoot_kubeconfig}" > "${LOCAL_KUBECONFIG}"
  export KUBECONFIG="${shoot_kubeconfig}"
}

# Break Kyma to test reconciler repair mechanism
function reconciler::break_kyma() {
  log::banner "Delete all deployments from kyma-system ns"
  kubectl delete deploy -n kyma-system --all
}
