#!/bin/bash
# Source development/install-prow.sh
# Changes:
#  - Removed secret creation
#  - Removed default ingress `ing` after installation
#  - Added TLS configuration
#  - Added Branch Protector deployment

set -o errexit

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly KUBECONFIG=${KUBECONFIG:-"${HOME}/.kube/config"}
readonly PROW_CLUSTER_DIR="$( cd "${CURRENT_DIR}/cluster" && pwd )"

if [ -z "$BUCKET_NAME" ]; then
      echo "\$BUCKET_NAME is empty"
      exit 1
fi

if [ -z "$KEYRING_NAME" ]; then
      echo "\$KEYRING_NAME is empty"
      exit 1
fi

if [ -z "$ENCRYPTION_KEY_NAME" ]; then
      echo "\$ENCRYPTION_KEY_NAME is empty"
      exit 1
fi

if [ -z "${LOCATION}" ]; then
    LOCATION="global"
fi


# requried by secretspopulator
if [ -z "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
      echo "\$GOOGLE_APPLICATION_CREDENTIALS is empty"
      exit 1
fi

if [ -z "$PROJECT" ]; then
      echo "\$PROJECT is empty"
      exit 1
fi

kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole cluster-admin --user "$(gcloud config get-value account)"

# Deploy NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.21.0/deploy/mandatory.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.21.0/deploy/provider/cloud-generic.yaml

# Create secrets
go run "${CURRENT_DIR}/../development/tools/cmd/secretspopulator/main.go" --project="${PROJECT}" --location "${LOCATION}" --bucket "${BUCKET_NAME}" --keyring "${KEYRING_NAME}" --key "${ENCRYPTION_KEY_NAME}" --kubeconfig "${KUBECONFIG}" --secrets-def-file="${PROW_CLUSTER_DIR}/required-secrets.yaml"

# Create ConfigMap with Kyma images for deck
kubectl create configmap branding --from-file "${CURRENT_DIR}/branding"

# Deploy GCE SSD StorageClass
kubectl apply -f cluster/09-gce-ssd-retain_storageclass.yaml

# Deploy ghProxy
kubectl apply -f cluster/10-ghproxy.yaml

# Deploy Prow
kubectl apply -f cluster/starter.yaml

# Enable https redirection on deck
kubectl patch deployment deck --patch "$(cat cluster/00-deck-patch.yaml)"

# Patch workload volume for plank/deck/sinker
kubectl patch deployment deck --patch "$(cat cluster/12-deck-patch-workload.yaml)"
kubectl patch deployment sinker --patch "$(cat cluster/13-sinker-patch-workload.yaml)"
kubectl patch deployment plank --patch "$(cat cluster/14-plank-patch-workload.yaml)"

# Install cert-manager
kubectl apply -f cluster/01-cert-manager.yaml
kubectl apply -f cluster/02-cluster-issuer.yaml

# Install secure ingress
kubectl apply -f cluster/03-tls-ing_ingress.yaml

# Install branch protector
kubectl apply -f cluster/04-branchprotector_cronjob.yaml

# Install tiller
kubectl apply -f cluster/05-tiller.yaml

# Install pushgateway
kubectl apply -f cluster/06-pushgateway_deployment.yaml

# Install PodDisruptionBudgets
kubectl apply -f cluster/07-kube-system_poddisruptionbudgets.yaml

# Install prow-addons-ctrl-manager
kubectl apply -f cluster/08-prow-addons-ctrl-manager.yaml

# Remove Insecure ingress 
kubectl delete ingress ing
