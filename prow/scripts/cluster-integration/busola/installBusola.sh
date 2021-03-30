#!/bin/bash
set -e

names=($( cat ../names.txt ))

export BUSOLA_BRANCH_NAME=$1
if [ $# -eq 0 ]; then
    export BUSOLA_BRANCH_NAME=main
fi

export DOMAIN_NAME=$2
if [ ${#DOMAIN_NAME} -eq 0 ]; then
    randName=$[$RANDOM % ${#names[@]}]
    export DOMAIN_NAME=${names[$randName]}
fi

export KUBECONFIG=./kubeconfigs/kubeconfig-garden-hasselhoff.yaml

echo "We will instal Busola version: $BUSOLA_BRANCH_NAME on the cluster : $DOMAIN_NAME"
# We create the cluster
cat devBusola.yaml | envsubst | kubectl create -f -

# #we wait for the cluster to be ready
kubectl wait --for condition="ControlPlaneHealthy" --timeout=10m shoot $DOMAIN_NAME

# #we switch to the new cluster
kubectl get secrets $DOMAIN_NAME.kubeconfig -o jsonpath={.data.kubeconfig} | base64 -d > ./kubeconfigs/kubeconfig--hasselhoff--$DOMAIN_NAME.yaml
export KUBECONFIG=./kubeconfigs/kubeconfig--hasselhoff--$DOMAIN_NAME.yaml

# # we ask for new certificates
cat wildcardCert.yaml | envsubst | kubectl apply -f -

helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

cat nginxValues.yaml | envsubst | helm install ingress-nginx --namespace=kube-system -f - ingress-nginx/ingress-nginx

#wait for ingress controller to start
kubectl wait --namespace kube-system \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=120s

#install busola
rm -rf busola
git clone -b $BUSOLA_BRANCH_NAME https://github.com/kyma-project/busola.git
cd busola/resources
export KUBECONFIG=./../../kubeconfigs/kubeconfig--hasselhoff--$DOMAIN_NAME.yaml
./apply-resources.sh $DOMAIN_NAME.hasselhoff.shoot.canary.k8s-hana.ondemand.com

kubectl cluster-info
echo "Please generate params for using k8s http://enkode.surge.sh/"
echo "Kyma busola Url:"
echo "https://busola.${DOMAIN_NAME}.hasselhoff.shoot.canary.k8s-hana.ondemand.com?auth=generated_params_in_previous_step"