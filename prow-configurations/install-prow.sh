#!/bin/bash

set -o errexit

## Create an HMAC token
hmac_token="$(openssl rand -hex 20)"
echo $hmac_token > hmac_token.txt
echo "Token hmac stored in hmac_token.txt file"

if [ "$OAUTH" == "" ]; then
    echo -n "Enter OAuth2 token that has read and write access to the bot account, followed by [ENTER]: (input will not be printed)"
    read -s oauth_token
else
    oauth_token="$OAUTH"
fi

echo

if [ ${#oauth_token} -lt 1 ]; then
  echo "OAuth2 token not provided";
  exit -1;
fi

kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole cluster-admin --user $(gcloud config get-value account)

# Deploy NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.20.0/deploy/mandatory.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.20.0/deploy/provider/cloud-generic.yaml

# Follow the installation steps in https://github.com/kubernetes/test-infra/blob/master/prow/getting_started.md#how-to-turn-up-a-new-cluster

kubectl create secret generic hmac-token --from-literal=hmac=$hmac_token
kubectl create secret generic oauth-token --from-literal=oauth=$oauth_token

kubectl apply -f https://raw.githubusercontent.com/kubernetes/test-infra/a202e595a33ac92ab503f913f2d710efabd3de21/prow/cluster/starter.yaml

# Add annotations to Prow Ingress 
kubectl annotate ingress ing kubernetes.io/ingress.class=nginx nginx.ingress.kubernetes.io/ssl-redirect=false
kubectl patch ingress ing --type=json -p='[{"op": "replace", "path": "/spec/rules/0/http/paths/0/path", "value":"/"}]'

# # Deploy the plugin configurations
# kubectl create configmap plugins \
#   --from-file=plugins.yaml=$(pwd)/plugins.yaml --dry-run -o yaml \
#   | kubectl replace configmap plugins -f -