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

NUM_NODES=${NUM_NODES:-2} # default 2 nodes

echo "Provisioning cluster '${CLUSTER_NAME}' (${NUM_NODES} nodes) in project '${PROJECT}' and zone '${ZONE}'"

gcloud container --project "${PROJECT}" clusters create "${CLUSTER_NAME}" \
  --zone "${ZONE}" --machine-type n1-standard-4 --num-nodes "${NUM_NODES}"