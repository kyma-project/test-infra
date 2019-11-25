#!/usr/bin/env bash

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# shellcheck disable=SC1090
source "${CURRENT_DIR}/log.sh"

function context_arg() {
    if [ -n "$KUBE_CONTEXT" ]; then
        echo "--context $KUBE_CONTEXT"
    fi
}

# retries are useful when api call can fail due to the infrastructure issue
function executeKubectlWithRetries() {
    local command="$1"
    local retry=0
    local result=""

    while [[ ${retry} -lt 10 ]]; do
        result=$(${command})
        if [[ $? -eq 0 ]]; then
            echo "${result}"
            return 0
        else
            sleep 5
        fi
        (( retry++ ))
    done
    echo "Maximum retries exceeded: ${result}"
    return 1
}

function cmdGetPodsForSuite() {
    local suiteName=$1
    cmd="kubectl $(context_arg) get pods -l testing.kyma-project.io/suite-name=${suiteName} \
            --all-namespaces \
            --no-headers=true \
            -o=custom-columns=name:metadata.name,ns:metadata.namespace"
    result=$(executeKubectlWithRetries "${cmd}")
    if [[ $? -eq 1 ]]; then
        echo "${result}"
        return 1
    fi
    echo "${result}"
}

function checkTestPodTerminated() {
    local suiteName=$1
    runningPods=false

    pod=""
    namespace=""
    idx=0

    result=$(cmdGetPodsForSuite "${suiteName}")
    if [[ $? -eq 1 ]]; then
        echo "${result}"
        return 1
    fi

    for podOrNs in ${result}
    do
       n=$((idx%2))
       if [[ "$n" == 0 ]];then
         pod=${podOrNs}
         idx=$((idx+1))
         continue
       fi
        namespace=${podOrNs}
        idx=$((idx+1))

        phase=$(executeKubectlWithRetries "kubectl $(context_arg) get pod $pod -n ${namespace} -o jsonpath={.status.phase}")
        if [[ $? -eq 1 ]]; then
            echo "${phase}"
            return 1
        fi
        # A Pod's phase  Failed or Succeeded means pod has terminated.
        # see: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase
        if [ "${phase}" !=  "Succeeded" ] && [ "${phase}" != "Failed" ]
        then
          log::error "Test pod '${pod}' has not terminated, pod phase: ${phase}"
          runningPods=true
        fi
    done

    if [ ${runningPods} = true ];
    then
        return 1
    fi
}

function waitForTestPodsTermination() {
    local retry=0
    local suiteName=$1

    while [ ${retry} -lt 3 ]; do
        checkTestPodTerminated "${suiteName}"
        checkTestPodTerminatedErr=$?
        if [ ${checkTestPodTerminatedErr} -ne 0 ]; then
            echo "Waiting for test pods to terminate..."
            sleep 1
        else
            log::success "OK"
            return 0
        fi
        retry=$((retry + 1))
    done
    log::error "FAILED"
    return 1
}

function printImagesWithLatestTag() {
    retry=10
    while true; do
        # shellcheck disable=SC2046
        images=$(kubectl $(context_arg)  get pods --all-namespaces -o jsonpath="{..image}" |\
        tr -s '[:space:]' '\n' |\
        grep ":latest")

        # TODO(michal-hudy): it shoudn't be done that way, grep returns 1 when no lines match, same bug in kyma repository....
        if [[ $? -lt 2 ]]; then
            break
        fi
        (( retry-- ))
        if [[ ${retry} -eq 0 ]]; then
            log::error "Reached maximum attempts, not trying any longer"
            return 1
        fi
        sleep 5
    done

    if [ ${#images} -ne 0 ]; then
        log::error "${images}"
        log::error "FAILED"
        return 1
    fi
    log::success "OK"
    return 0
}

TESTING_ADDONS_CFG_NAME="testing-addons"

function injectTestingAddons() {
    retry=10
    while true; do
        cat <<EOF | kubectl apply -f -
apiVersion: addons.kyma-project.io/v1alpha1
kind: ClusterAddonsConfiguration
metadata:
  labels:
    addons.kyma-project.io/managed: "true"
  name: ${TESTING_ADDONS_CFG_NAME}
spec:
  repositories:
  - url: "https://github.com/kyma-project/addons/releases/download/0.8.0/index-testing.yaml"
EOF
        if [[ $? -eq 0 ]]; then
            break
        fi
        (( retry-- ))
        if [[ ${retry} -eq 0 ]]; then
            return 1
        fi
        sleep 5
    done

    local retry=0
    while [[ ${retry} -lt 10 ]]; do
        msg=$(kubectl get clusteraddonsconfiguration ${TESTING_ADDONS_CFG_NAME} -o=jsonpath='{.status.phase}')
        if [[ "${msg}" = "Ready" ]]; then
            log::success "Testing addons injected"
            return 0
        fi
        if [[ "${msg}" = "Failed" ]]; then
            log::error "Testing addons configuration failed"
            removeTestingAddons
            return 1
        fi
        echo "Waiting for ready testing addons ${retry}/10.. status: ${msg}"
        retry=$((retry + 1))
        sleep 3
    done
    log::error "Testing addons couldn't be injected"
    return 1
}

function removeTestingAddons() {
    result=$(executeKubectlWithRetries "kubectl delete clusteraddonsconfiguration ${TESTING_ADDONS_CFG_NAME}")
    echo "${result}"
    if [[ $? -eq 1 ]]; then
        return 1
    fi
    log::success "Testing addons removed"
}
