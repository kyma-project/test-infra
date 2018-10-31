#!/bin/bash

DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROW_CLUSTER_DIR="$( cd "${DEVELOPMENT_DIR}/../prow/cluster" && pwd )"

kubectl delete -f ${PROW_CLUSTER_DIR}/starter.yaml
kubectl delete pods -l created-by-prow=true
kubectl delete secret hmac-token
kubectl delete secret oauth-token
kubectl delete secret sa-vm-kyma-integration
kubectl delete secret sa-gke-kyma-integration
kubectl delete clusterrolebinding cluster-admin-binding
kubectl delete -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/mandatory.yaml