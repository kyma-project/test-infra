#!/bin/bash

readonly DEVELOPMENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly PROW_CLUSTER_DIR="$( cd "${DEVELOPMENT_DIR}/../prow/cluster" && pwd )"

kubectl delete -f "${PROW_CLUSTER_DIR}/starter.yaml"
kubectl delete pods -l created-by-prow=true
kubectl delete secret hmac-token
kubectl delete secret oauth-token
kubectl delete secret sa-vm-kyma-integration
kubectl delete secret sa-gke-kyma-integration
kubectl delete secret sa-gcs-plank
kubectl delete secret sa-gcr-push
kubeclt delete secret kyma-bot-npm-token
kubeclt delete secret sa-kyma-artifacts
kubectl delete clusterrolebinding cluster-admin-binding
kubectl delete -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/mandatory.yaml