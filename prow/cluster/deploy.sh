#!/bin/bash

# Prow deploy script
# Based on https://github.com/kubernetes/test-infra/blob/master/prow/deploy.sh

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

prow_components=(
"test-pods_namespace.yaml"
"crier_rbac.yaml"
"crier_deployment.yaml"
"crier_service.yaml"
"deck_rbac.yaml"
"deck_deployment.yaml"
"deck_service.yaml"
"gce-ssd-retain_storageclass.yaml"
"ghproxy.yaml"
"gcsweb-kyma-prow_managedcertificate.yaml"
"gcsweb.yaml"
"gcsweb_tls_ingress.yaml"
"hook_rbac.yaml"
"hook_deployment.yaml"
"hook_service.yaml"
"horologium_rbac.yaml"
"horologium_deployment.yaml"
"horologium_service.yaml"
"halogen.yaml"
"pjtester_prowjob-scheduler_rbac.yaml"
"prow_controller_manager_rbac.yaml"
"prow_controller_manager_deployment.yaml"
"prow_controller_manager_service.yaml"
"prowjob_customresourcedefinition.yaml"
"pushgateway_deployment.yaml"
"sinker_rbac.yaml"
"sinker_deployment.yaml"
"sinker_service.yaml"
"statusreconciler_rbac.yaml"
"statusreconciler_deployment.yaml"
"status-kyma-prow_managedcertificate.yaml"
"tls-ing_ingress.yaml"
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
for c in "${prow_components[@]}"; do
  kubectl apply --server-side=true -f "$SCRIPT_DIR/components/$c"
done
