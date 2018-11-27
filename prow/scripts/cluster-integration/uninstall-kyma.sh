#!/usr/bin/env bash

################################################################################
# This script waits for uninstallation of PVC and Ingress objects
# from Kyma cluster. Once these are gone, we can safely delete
# the entire cluster, because no other external resources will be left.
################################################################################

set -o errexit

pvc_count() {
    RES=$(kubectl get pvc --all-namespaces 2>/dev/null | wc -l)
    echo "${RES}"
}

ingress_count() {
    RES=$(kubectl get ingress --all-namespaces 2>/dev/null | wc -l)
    echo "${RES}"
}

START_TIME=$(date +%s)
ELAPSED_TIME=0
SLEEP_TIME=10
MAX_TIME=600
LOOP="true"

echo "Waiting for Kyma uninstallation. Time limit: ${MAX_TIME}[s]"
while [ "${LOOP}" = "true" ]; do
    #object counts
    PVC_COUNT=$(pvc_count)
    INGRESS_COUNT=$(ingress_count)

    #calculate time
    TIME=$(date +%s)
    ELAPSED_TIME=$((TIME-START_TIME ))

    #loop end conditions
    if [ "${ELAPSED_TIME}" -gt "${MAX_TIME}" ]; then
        LOOP="false"
    fi

    if [ "${INGRESS_COUNT}" -eq 0 ] && [ "${PVC_COUNT}" -eq 0 ] ; then
        LOOP="false"
    fi

    if [ "${LOOP}" = "true" ]; then
        echo "PVC objects left: ${PVC_COUNT}, Ingress objects left: ${INGRESS_COUNT}"
        echo "Awaiting for uninstallation, elapsed time: ${ELAPSED_TIME}[s]"
        sleep "${SLEEP_TIME}"
    fi
done

if [ "${ELAPSED_TIME}" -gt "${MAX_TIME}" ]; then
    echo "Kyma unistallation didn't complete within the time limt of ${MAX_TIME}[s]..."
fi

if [ "${PVC_COUNT}" -gt 0 ]; then
    echo "Some PVC objects are left:"
    kubectl get pvc --all-namespaces
else
    echo "PVC objects are cleaned up"
fi

if [ "${INGRESS_COUNT}" -gt 0 ]; then
    echo "Some Ingress objects are left:"
    kubectl get ingress --all-namespaces
else
    echo "Ingress objects are cleaned up"
fi
