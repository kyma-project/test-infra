#!/bin/bash

# Prow deploy script
# Based on https://github.com/kubernetes/test-infra/blob/master/prow/deploy.sh

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

prow_components=(
"pjtester_prowjob-scheduler_rbac.yaml"
"web_server_deployment.yaml"
)

function ensure-context() {
  local proj=$1
  local zone=$2
  local cluster=$3
  local context="gke_${proj}_${zone}_${cluster}"
  echo -n "$context"
  kubectl config get-contexts "$context" &> /dev/null && return 0
  echo ": missing, getting credentials..."
  gcloud container clusters get-credentials --project="$proj" --zone="$zone" "$cluster"
  kubectl config get-contexts "$context" > /dev/null
  echo -n "Ensuring contexts exist:"
}

if ! [ -x "$(command -v "kubectl")" ]; then
  echo "ERROR: kubectl is not present. Exiting..."
  exit 1
fi

if [ -n "${GOOGLE_APPLICATION_CREDENTIALS:-}" ]; then
    echo "Detected GOOGLE_APPLICATION_CREDENTIALS. Activating service account..."
    gcloud auth activate-service-account --key-file="$GOOGLE_APPLICATION_CREDENTIALS"
fi

ensure-context sap-kyma-prow europe-west3-a prow

echo " Deploying Prow..."
kubectl apply --server-side=true -k "configs/deployments/prow/overlays/dev"
for c in "${prow_components[@]}"; do
  kubectl apply --server-side=true -f "$SCRIPT_DIR/components/$c"
done
