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

echo "Provisioning cluster '${CLUSTER_NAME}' in project '${PROJECT}' and zone '${ZONE}'"

gcloud container --project "${PROJECT}" clusters create "${CLUSTER_NAME}" \
  --zone "${ZONE}" --machine-type n1-standard-4 --num-nodes 2