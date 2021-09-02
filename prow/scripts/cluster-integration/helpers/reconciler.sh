#!/usr/bin/env bash

readonly RECONCILER_SUFFIX="-reconciler"
readonly RECONCILER_NAMESPACE=reconciler

# Checks whether reconciler is ready
function reconciler::wait_until_reconciler_is_ready() {
  timeout=1200 # in secs
  delay=10 # in secs
  iterationsLeft=$(( timeout/delay ))
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

    if [ "$timeout" -ne 0 ] && [ "$iterationsLeft" -le 0 ]; then
      log::info "Timeout reached while waiting for reconciler to be ready. Exiting"
      exit 1
    fi
    sleep $delay
    iterationsLeft=$(( iterationsLeft-1 ))
  done
}
