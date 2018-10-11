#!/bin/bash

kubectl delete -f https://raw.githubusercontent.com/kubernetes/test-infra/a202e595a33ac92ab503f913f2d710efabd3de21/prow/cluster/starter.yaml
kubectl delete pods -l created-by-prow=true
kubectl delete secret hmac-token
kubectl delete secret oauth-token
kubectl delete clusterrolebinding cluster-admin-binding
kubectl delete -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/mandatory.yaml