#!/bin/bash

set -o errexit

git clone git@github.com:kubernetes/test-infra.git
cd test-infra
git checkout a202e595a33ac92ab503f913f2d710efabd3de21

# Deploy NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/mandatory.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/cloud-generic.yaml

# Follow the installation steps in https://github.com/kubernetes/test-infra/blob/master/prow/getting_started.md#how-to-turn-up-a-new-cluster
kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole cluster-admin --user $(gcloud config get-value account)

## Create an HMAC token
hmac_token="$(openssl rand -hex 20)"
echo "hmac-token: $hmac_token"

if [ "$OAUTH" == "" ]; then
    echo -n "Enter OAuth2 token that has read and write access to the bot account, followed by [ENTER]: "
    read oauth_token
else
    oauth_token="$OAUTH"
fi

kubectl create secret generic hmac-token --from-literal=hmac=$hmac_token
kubectl create secret generic oauth-token --from-literal=oauth=$oauth_token

kubectl apply -f prow/cluster/starter.yaml

# Add annotations to Prow Ingress 
kubectl annotate ingress ing kubernetes.io/ingress.class=nginx nginx.ingress.kubernetes.io/ssl-redirect=false

# Change Deck Service type to LoadBalancer
kubectl patch service deck -p '{"spec": {"type": "LoadBalancer"}}'

cd ..
rm -rf test-infra

# Deploy the plugin configurations
kubectl create configmap plugins \
  --from-file=plugins.yaml=$(pwd)/plugins.yaml --dry-run -o yaml \
  | kubectl replace configmap plugins -f -