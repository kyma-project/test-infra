#!/bin/bash
set -e

BASE_PATH=${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/busola/

export DOMAIN_NAME=$1

echo "We will instal Busola on the cluster: ${DOMAIN_NAME}"
# We create the cluster
cat ${BASE_PATH}/devBusola.yaml | envsubst | kubectl create -f -

# #we wait for the cluster to be ready
kubectl wait --for condition="ControlPlaneHealthy" --timeout=10m shoot "${DOMAIN_NAME}"

# #we switch to the new cluster
kubectl get secrets "${DOMAIN_NAME}.kubeconfig" -o jsonpath={.data.kubeconfig} | base64 -d > "${BASE_PATH}/kubeconfig--busola--${DOMAIN_NAME}.yaml"
export KUBECONFIG="${BASE_PATH}/kubeconfig--busola--$DOMAIN_NAME.yaml"

# # we ask for new certificates
cat "${BASE_PATH}/wildcardCert.yaml" | envsubst | kubectl apply -f -

helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

cat "${BASE_PATH}/nginxValues.yaml" | envsubst | helm install ingress-nginx --namespace=kube-system -f - ingress-nginx/ingress-nginx

#wait for ingress controller to start
kubectl wait --namespace kube-system \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=120s

#install busola
set +e
FULL_DOMAIN="${DOMAIN_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.canary.k8s-hana.ondemand.com"
for i in "${BUSOLA_SOURCES_DIR}/resources/**{/*,}.yaml"; do
    sed -i '' "s/%DOMAIN%/${FULL_DOMAIN}/g" ${i}
done
set -e

echo "kubectl"

kubectl apply -k "${BUSOLA_SOURCES_DIR}/resources"

kubectl cluster-info
echo "Please generate params for using k8s http://enkode.surge.sh/"
echo "Kyma busola Url:"
echo "https://busola.${DOMAIN_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.canary.k8s-hana.ondemand.com?auth=generated_params_in_previous_step"
