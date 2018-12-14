#!/bin/bash
# Source development/install-prow.sh
# Changes:
#  - Removed secret creation
#  - Removed default ingress `ing` after installation
#  - Added TLS configuration
#  - Added Branch Protector deployment

set -o errexit

kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole cluster-admin --user "$(gcloud config get-value account)"

# Deploy NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.21.0/deploy/mandatory.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.21.0/deploy/provider/cloud-generic.yaml

# Deploy Prow
kubectl apply -f cluster/starter.yaml

# Enable https redirection on deck
kubectl patch deployment deck --patch "$(cat cluster/00-deck-patch.yaml)"

# Install cert-manager
kubectl apply -f cluster/01-cert-manager.yaml
kubectl apply -f cluster/02-cluster-issuer.yaml

# Install secure ingress
kubectl apply -f cluster/03-tls-ing_ingress.yaml

# Install branch protector
kubectl apply -f cluster/04-branchprotector_cronjob.yaml

# Remove Insecure ingress 
kubectl delete ingress ing
