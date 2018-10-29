#!/bin/bash
# Source prow-configurations/install-prow.sh
# Changes:
#  - Removed secret creation
#  - Remove default ingress `ing` after installation

set -o errexit

kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole cluster-admin --user $(gcloud config get-value account)

# Deploy NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.20.0/deploy/mandatory.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.20.0/deploy/provider/cloud-generic.yaml

# Deploy Prow
kubectl apply -f https://raw.githubusercontent.com/kubernetes/test-infra/a202e595a33ac92ab503f913f2d710efabd3de21/prow/cluster/starter.yaml

# Install cert-manager
kubectl apply -f cluster/01-cert-manager.yaml
kubectl apply -f cluster/02-cluster-issuer.yaml

# Install secure ingress
kubectl apply -f cluster/03-tls-ing_ingerss.yaml

# Remove Insecure ingress 
kubectl delete ingress ing