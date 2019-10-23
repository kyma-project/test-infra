#!/bin/bash


######################################
#THIS SCRIPT IS OUTDATED. DO NOT USE.#
######################################
set -o errexit

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly PROW_CLUSTER_DIR="$( cd "${CURRENT_DIR}/../prow/cluster" && pwd )"
readonly KUBECONFIG=${KUBECONFIG:-"${HOME}/.kube/config"}

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

## Create an HMAC token
hmac_token="$(openssl rand -hex 32)"
echo "$hmac_token" > hmac_token.txt
echo "Token hmac stored in hmac_token.txt file"

if [ "$OAUTH" == "" ]; then
    echo -n "Enter OAuth2 token that has read and write access to the bot account, followed by [ENTER]: (input will not be printed)"
    read -rs oauth_token
else
    oauth_token="$OAUTH"
fi

echo

if [ ${#oauth_token} -lt 1 ]; then
  echo "OAuth2 token not provided";
  exit -1;
fi

kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole cluster-admin --user "$(gcloud config get-value account)"

# Deploy NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.21.0/deploy/mandatory.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.21.0/deploy/provider/cloud-generic.yaml

# Follow the installation steps in https://github.com/kubernetes/test-infra/blob/master/prow/getting_started.md#how-to-turn-up-a-new-cluster

kubectl create secret generic hmac-token --from-literal=hmac="$hmac_token"
kubectl create secret generic oauth-token --from-literal=oauth="$oauth_token"

# Create GCP secrets
go run "${CURRENT_DIR}/tools/cmd/secretspopulator/main.go" --project="${PROJECT}" --location "${LOCATION}" --bucket "${BUCKET_NAME}" --keyring "${KEYRING_NAME}" --key "${ENCRYPTION_KEY_NAME}" --kubeconfig "${KUBECONFIG}" --secrets-def-file="${PROW_CLUSTER_DIR}/required-secrets.yaml"

kubectl apply -f "${PROW_CLUSTER_DIR}/starter.yaml"
