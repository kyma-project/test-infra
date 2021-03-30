#!/bin/bash
set -e

names=($( cat ../names.txt ))

export KYMA_VERSION=$1
if [ $# -eq 0 ]; then
    export KYMA_VERSION=master
fi

export DOMAIN_NAME=$2
if [ ${#DOMAIN_NAME} -eq 0 ]; then
    randName=$[$RANDOM % ${#names[@]}]
    export DOMAIN_NAME=${names[$randName]}
fi

export KUBECONFIG=./kubeconfigs/kubeconfig-garden-hasselhoff.yaml

export KYMA_PASSWORD=$(openssl rand -base64 16 | tr -d '\n=/iIlLOo0')
echo "We will instal Kyma version: $KYMA_VERSION on the cluster : $DOMAIN_NAME"
echo "Console admin password: $KYMA_PASSWORD"
# We create the cluster
cat devCluster.yaml | envsubst | kubectl create -f -

# #we wait for the cluster to be ready
kubectl wait --for condition="ControlPlaneHealthy" --timeout=10m shoot $DOMAIN_NAME

# #we switch to the new cluster
kubectl get secrets $DOMAIN_NAME.kubeconfig -o jsonpath={.data.kubeconfig} | base64 -d > ./kubeconfigs/kubeconfig--hasselhoff--$DOMAIN_NAME.yaml
export KUBECONFIG=./kubeconfigs/kubeconfig--hasselhoff--$DOMAIN_NAME.yaml
# #we upgrade the CLI
brew upgrade kyma-cli
# # we install kyma

kyma install \
--kubeconfig=./kubeconfigs/kubeconfig--hasselhoff--$DOMAIN_NAME.yaml \
--profile=evaluation \
--source=$KYMA_VERSION \
--fallback-level 5 \
-p=$KYMA_PASSWORD

echo 'Kyma Console Url:'
echo 
kubectl get virtualservice console-web -n kyma-system -o jsonpath='{ .spec.hosts[0] }'
echo
echo 'User admin@kyma.cx'
echo 'password:'
kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode
echo 'k8s API Url:'
echo "https://api.${DOMAIN_NAME}.hasselhoff.shoot.canary.k8s-hana.ondemand.com"