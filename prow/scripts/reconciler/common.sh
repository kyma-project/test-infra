#!/usr/bin/env bash

readonly RECONCILER_SUFFIX="-reconciler"
readonly RECONCILER_NAMESPACE=reconciler
readonly RECONCILER_TIMEOUT=1200 # in secs
readonly RECONCILER_DELAY=15 # in secs


function reconciler::export_nightly_cluster_name(){
  echo ">>> Export nightly cluster name"
  day="$(date +%a | tr "[:upper:]" "[:lower:]" | cut -c1-2)"
  export INPUT_CLUSTER_NAME="${INPUT_CLUSTER_NAME}-${day}"
}

function reconciler::delete_cluster_if_exists(){
  echo ">>> Delete cluster with reconciler if exists"
  export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"
  for i in mo tu we th fr sa su
  do
    local name="${INPUT_CLUSTER_NAME}-${i}"
    local namespace="garden-${GARDENER_KYMA_PROW_PROJECT_NAME}"
    set +e
    existing_shoot=$(kubectl get shoot "${name}" -ojsonpath="{ .metadata.name }")
    if [ -n "${existing_shoot}" ]; then
      echo "Cluster found and deleting '${name}'"
      kubectl annotate shoot "${name}" confirmation.gardener.cloud/deletion=true \
          --overwrite \
          -n "${namespace}" \
          --kubeconfig "${GARDENER_KYMA_PROW_KUBECONFIG}"
      kubectl delete shoot "${name}" \
        --wait=true \
        --kubeconfig "${GARDENER_KYMA_PROW_KUBECONFIG}" \
        -n "${namespace}"

      echo "We wait 120s for Gardener Shoot to settle after cluster deletion"
      sleep 120
    else
      echo "Cluster '${name}' does not exist"
    fi
    set -e
  done
}

# reconciler::reprovision_cluster will generate new cluster name
# and start provisioning again
function reconciler::reprovision_cluster() {
    echo "cluster provisioning failed, trying provision new cluster"
    echo "cleaning damaged cluster first"
    local namespace="garden-${GARDENER_KYMA_PROW_PROJECT_NAME}"

    kubectl annotate shoot "${INPUT_CLUSTER_NAME}" confirmation.gardener.cloud/deletion=true \
        --overwrite \
        -n "${namespace}" \
        --kubeconfig "${GARDENER_KYMA_PROW_KUBECONFIG}"
    kubectl delete shoot "${INPUT_CLUSTER_NAME}" \
      --wait=true \
      --kubeconfig "${GARDENER_KYMA_PROW_KUBECONFIG}" \
      -n "${namespace}"
    echo "building new cluster name"

    reconciler::provision_cluster
}

function reconciler::provision_cluster() {
    echo "Provision reconciler cluster"
    export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"
    export DOMAIN_NAME="${INPUT_CLUSTER_NAME}"
    echo "Creating cluster: ${INPUT_CLUSTER_NAME}"

    # catch cluster provisioning errors and try provision new one
    trap reconciler::reprovision_cluster ERR

    set +e
    # create the cluster
    cat <<EOF | kubectl create -f -
apiVersion: core.gardener.cloud/v1beta1
kind: Shoot
metadata:
  name: $DOMAIN_NAME
spec:
  purpose: development
  cloudProfileName: gcp
  kubernetes:
    version: 1.26.5
  provider:
    controlPlaneConfig:
      apiVersion: gcp.provider.extensions.gardener.cloud/v1alpha1
      kind: ControlPlaneConfig
      zone: $GARDENER_ZONES
    infrastructureConfig:
      apiVersion: gcp.provider.extensions.gardener.cloud/v1alpha1
      kind: InfrastructureConfig
      networks:
        workers: 10.250.0.0/16
    type: gcp
    workers:
      - machine:
          image:
            name: gardenlinux
            version: 934.8.0
          type: n1-standard-4
        maxSurge: 1
        maxUnavailable: 0
        maximum: 4
        minimum: 2
        name: worker-dev
        volume:
          size: 20Gi
          type: pd-ssd
        zones:
          - $GARDENER_ZONES
  networking:
    nodes: 10.250.0.0/16
    pods: 100.96.0.0/11
    services: 100.64.0.0/13
    type: calico
  maintenance:
    timeWindow:
      begin: 010000+0000
      end: 020000+0000
    autoUpdate:
      kubernetesVersion: true
      machineImageVersion: true
  addons:
    kubernetesDashboard:
      enabled: false
    nginxIngress:
      enabled: false
  hibernation:
    enabled: false
  region: $GARDENER_REGION
  secretBindingName: $GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
EOF

    set -e

    # wait for the cluster to be ready
    kubectl wait --for condition="ControlPlaneHealthy" --timeout=20m shoot "${INPUT_CLUSTER_NAME}"
    echo "Cluster ${INPUT_CLUSTER_NAME} was created successfully"

    # disable trap for cluster provisioning errors to not call it for later errors
    trap - ERR
}

function reconciler::deploy() {
  # Deploy reconciler to cluster
  echo ">>> Deploying Reconciler in the cluster"
  make -C tools/reconciler deploy-reconciler
}

# Checks whether reconciler is ready
function reconciler::wait_until_is_ready() {
  echo ">>> Wait until reconciler is in ready state"
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
      echo "Reconciler is successfully installed"
      break
    fi

    if [ "$RECONCILER_TIMEOUT" -ne 0 ] && [ "$iterationsLeft" -le 0 ]; then
      echo "Current state of pods in reconciler namespace"
      pods_info=$(kubectl get po -n "${RECONCILER_NAMESPACE}")
      echo "${pods_info}"
      echo "Timeout reached while waiting for reconciler to be ready. Exiting"
      exit 1
    fi
    sleep $RECONCILER_DELAY
    iterationsLeft=$(( iterationsLeft-1 ))
  done
}

# Waits until the test-pod is in ready state
function reconciler::wait_until_test_pod_is_ready() {
  echo ">>> Wait until test-pod is in ready state"
  iterationsLeft=$(( RECONCILER_TIMEOUT/RECONCILER_DELAY ))
  while : ; do
    testPodStatus=$(kubectl get po -n "${RECONCILER_NAMESPACE}" test-pod -ojsonpath='{.status.containerStatuses[?(@.name == "test-pod")].ready}')
    if [ "${testPodStatus}" = "true" ]; then
      echo "Test pod is ready"
      break
    fi
    if [ "$RECONCILER_TIMEOUT" -ne 0 ] && [ "$iterationsLeft" -le 0 ]; then
      echo "Timeout reached while initializing test pod. Exiting"
      exit 1
    fi
    echo "Waiting for test pod to be ready..."
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
      echo "Test pod is deleted"
      break
    fi
    if [ "$RECONCILER_TIMEOUT" -ne 0 ] && [ "$iterationsLeft" -le 0 ]; then
      echo "Timeout reached while initializing test pod. Exiting"
      exit 1
    fi
    echo "Waiting for test pod to be deleted..."
    sleep $RECONCILER_DELAY
    iterationsLeft=$(( iterationsLeft-1 ))
  done
}

# Initializes test pod which will send reconcile requests to reconciler
function reconciler::initialize_test_pod() {
  echo ">>> Set up test pod environment"
  if [[ ! $KYMA_UPGRADE_SOURCE ]]; then
    KYMA_UPGRADE_SOURCE="main"
  fi
  echo "Kyma version to reconcile: ${KYMA_UPGRADE_SOURCE}"

  # move to reconciler directory
  if [ ! -f "resources/kcp/charts/mothership-reconciler/values.yaml" ]; then
    echo "Mothership reconciler values file not found! (resources/kcp/charts/mothership-reconciler/values.yaml)"
    exit 1
  fi
  echo "************* Current Reconciler Image To Be Used **************"
  cat < "resources/kcp/charts/mothership-reconciler/values.yaml" | grep -o 'mothership_reconciler.*\"'
  echo "****************************************************************"
  # Create reconcile request payload with kubeconfig, domain, and version to the test-pod
  domain="$(kubectl get cm shoot-info -n kube-system -o jsonpath='{.data.domain}')"

  # shellcheck disable=SC2086
  kc="$(cat ${KUBECONFIG})"

  pushd "tools/reconciler" || { echo "Failed to change dir to: tools/reconciler"; exit 1; }
  local tplFile="./e2e-test/template-kyma-main.json"
  if [[ "$KYMA_UPGRADE_SOURCE" =~ ^2\.19\.[0-9]+$ ]] ; then
    tplFile="./e2e-test/template-kyma-2-19.json"
  fi
  echo "Calling reconciler by using JSON template '$tplFile' as payload"

  sed -i "s/example.com/$domain/" "$tplFile"
  # shellcheck disable=SC2016
  jq --arg kubeconfig "${kc}" --arg version "${KYMA_UPGRADE_SOURCE}" '.kubeconfig = $kubeconfig | .kymaConfig.version = $version' "$tplFile" > body.json

  # Copy the reconcile request payload and kyma reconciliation scripts to the test-pod
  tar -zcvf - ./body.json e2e-test/*.sh | kubectl exec -i -n "${RECONCILER_NAMESPACE}" test-pod -c test-pod -- tar -zxvf - -C /tmp --strip-components=1
  popd
}

# Only triggers reconciliation of Kyma
function reconciler::trigger_kyma_reconcile() {
  # Trigger Kyma reconciliation using reconciler
  echo ">>> Trigger the reconciliation through test pod"
  echo "Reconcile Kyma in the same cluster"
  kubectl exec -n "${RECONCILER_NAMESPACE}" test-pod -c test-pod -- sh -c ". /tmp/request-reconcile.sh"
  if [[ $? -ne 0 ]]; then
      echo "Failed to trigger reconciliation"
      exit 1
  fi
}

# Waits until Kyma reconciliation is in ready state
function reconciler::wait_until_kyma_reconciled() {
  echo ">>> Wait until reconciliation is complete"
  iterationsLeft=$(( RECONCILER_TIMEOUT/RECONCILER_DELAY ))
  while : ; do
    status=$(kubectl exec -n "${RECONCILER_NAMESPACE}" test-pod -c test-pod -- sh -c ". /tmp/get-reconcile-status.sh" | xargs || true)

    if [ -z "${status}" ]; then
      echo "Failed to retrieve reconciliation status. Retrying previous call in debug mode"
      kubectl exec -v=8 -n "${RECONCILER_NAMESPACE}" test-pod -c test-pod -- sh -c ". /tmp/get-reconcile-status.sh" || true
      status=$(kubectl exec -n "${RECONCILER_NAMESPACE}" test-pod -c test-pod -- sh -c ". /tmp/get-reconcile-status.sh" | xargs || true)
    fi

    if [ "${status}" = "ready" ]; then
      echo "Kyma is reconciled"
      break
    fi

    if [ "${status}" = "error" ]; then
      echo "Failed to reconcile Kyma. Exiting"
      kubectl logs -n "${RECONCILER_NAMESPACE}" -l app.kubernetes.io/name=mothership-reconciler -c mothership-reconciler --tail -1
      exit 1
    fi

    if [ -z "${status}" ]; then
      echo "Failed to retrieve reconciliation status. Checking if API server is reachable by asking for its version"
      kubectl version -v=8
    fi

    if [ "$RECONCILER_TIMEOUT" -ne 0 ] && [ "$iterationsLeft" -le 0 ]; then
      echo "Timeout reached on Kyma reconciliation. Exiting"
      kubectl logs -n "${RECONCILER_NAMESPACE}" -l app.kubernetes.io/name=mothership-reconciler -c mothership-reconciler --tail -1
      exit 1
    fi

    sleep $RECONCILER_DELAY
    echo "Waiting for reconciliation to finish, current status: ${status} ...."
    iterationsLeft=$(( iterationsLeft-1 ))
  done
}

# Deploy test pod
function reconciler::deploy_test_pod() {
  # Deploy a test pod
  echo ">>> Deploy test-pod in the cluster which will trigger reconciliation"
  test_pod_name=$(kubectl get po test-pod -n "${RECONCILER_NAMESPACE}" -ojsonpath="{ .metadata.name }" --ignore-not-found)
  if [ -n "${test_pod_name}" ]; then
    echo "Found existing pod: test-pod"
    kubectl delete po test-pod -n "${RECONCILER_NAMESPACE}"
    reconciler::wait_until_test_pod_is_deleted
  fi
  kubectl run -n "${RECONCILER_NAMESPACE}" --image=alpine:3.14.1 --restart=Never test-pod -- sh -c "sleep 36000"
}

function reconciler::disable_sidecar_injection_reconciler_ns() {
    echo ">>> Disabling sidecar injection for reconciler namespace"
    kubectl label namespace "${RECONCILER_NAMESPACE}" istio-injection=disabled --overwrite
}

# Export shoot cluster kubeconfig to ENV
function reconciler::export_shoot_cluster_kubeconfig() {
  echo "Export shoot cluster kubeconfig to ENV"
  export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"
  local shoot_kubeconfig="/tmp/shoot-kubeconfig.yaml"
  cat <<EOF | kubectl create -f - --raw "/apis/core.gardener.cloud/v1beta1/namespaces/garden-kyma-prow/shoots/${INPUT_CLUSTER_NAME}/adminkubeconfig" | jq -r ".status.kubeconfig" | base64 -d > "${shoot_kubeconfig}"
{
    "apiVersion": "authentication.gardener.cloud/v1alpha1",
    "kind": "AdminKubeconfigRequest",
    "spec": {
        "expirationSeconds": 10800
    }
}
EOF
  export KUBECONFIG="${shoot_kubeconfig}"
}

# Break Kyma to test reconciler repair mechanism
function reconciler::break_kyma() {
  echo "Delete all deployments from kyma-system ns"
  kubectl delete deploy -n kyma-system --all
}

function utils::check_required_vars() {
  echo "Checks if all provided variables are initialized"
    local discoverUnsetVar=false
    for var in "$@"; do
      if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
      fi
    done
    if [ "${discoverUnsetVar}" = true ] ; then
      exit 1
    fi
}

utils::generate_commonName() {

    local OPTIND

    while getopts ":n:p:" opt; do
        case $opt in
            n)
                local namePrefix="$OPTARG" ;;
            p)
                local id="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2; ;;
        esac
    done

    namePrefix=$(echo "$namePrefix" | tr '_' '-')
    namePrefix=${namePrefix#-}

    local randomNameSuffix
    randomNameSuffix=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
    # return value
    # shellcheck disable=SC2034
    utils_generate_commonName_return_commonName=$(echo "$namePrefix$id$randomNameSuffix" | tr "[:upper:]" "[:lower:]" )
}

gardener::cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        echo "AN ERROR OCCURED! Take a look at preceding log entries."
    fi

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    # describe nodes to file in artifacts directory
    kubectl describe nodes > "$ARTIFACTS/kubectl_describe.log"

    if  [[ "${CLEANUP_CLUSTER}" == "true" ]] ; then
      local namespace="garden-${GARDENER_KYMA_PROW_PROJECT_NAME}"
      echo "Deprovision cluster: \"${CLUSTER_NAME}\""
      kubectl annotate shoot "${CLUSTER_NAME}" confirmation.gardener.cloud/deletion=true \
              --overwrite \
              -n "${namespace}" \
              --kubeconfig "${GARDENER_KYMA_PROW_KUBECONFIG}"
      kubectl delete shoot "${CLUSTER_NAME}" \
        --wait=true \
        --kubeconfig "${GARDENER_KYMA_PROW_KUBECONFIG}" \
        -n "${namespace}"
    fi

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    echo "Job is finished ${MSG}"
    set -e

    exit "${EXIT_STATUS}"
}
