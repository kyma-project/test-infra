#!/usr/bin/env bash

if [ -z "$PROJECT" ]; then
      echo "\$PROJECT is empty"
      exit 1
fi

if [ -z "$CLUSTER_NAME" ]; then
      echo "\$CLUSTER_NAME is empty"
      exit 1
fi

if [ -z "$ZONE" ]; then
      echo "\$ZONE is empty"
      exit 1
fi

#NUM_NODES=${NUM_NODES:-2} # default 2 nodes
NUM_NODES=4

echo "Provisioning cluster '${CLUSTER_NAME}' (${NUM_NODES} nodes) in project '${PROJECT}' and zone '${ZONE}'"

gcloud container --project "${PROJECT}" clusters create "${CLUSTER_NAME}" \
  --zone "${ZONE}" --issue-client-certificate --enable-basic-auth --machine-type n1-standard-1 --num-nodes "${NUM_NODES}"

if [ -z "$WORKLOAD_CLUSTER_NAMER" ]; then
    echo "Provisioning workload cluster '${WORKLOAD_CLUSTER_NAME}' (${NUM_NODES} nodes) in project '${PROJECT}' and zone '${ZONE}'"
    gcloud container --project "${PROJECT}" clusters create "${WORKLOAD_CLUSTER_NAME}" \
        --zone "${ZONE}" --issue-client-certificate --enable-basic-auth --machine-type n1-standard-1 --num-nodes "${NUM_NODES}"

    echo "Adding high load node pool to cluster ${WORKLOAD_CLUSTER_NAME}' in project '${PROJECT}' and zone '${ZONE}'"
    gcloud container node-pools create high-load \
        --project "${PROJECT}" \
        --cluster "${WORKLOAD_CLUSTER_NAME}" \
        --zone "${ZONE}" \
        --machine-type n1-standard-16 \
        --node-taints resources-usage=high:NoSchedule \
        --num-nodes 0 \
        --enable-autoscaling \
        --max-nodes 5 \
        --min-nodes 0 \
        --disk-type pd-ssd
fi

