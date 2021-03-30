#!/bin/bash

set -e

BASE_PATH=${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/busola/

# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"

export KYMA_VERSION=$1
if [ $# -eq 0 ]; then
    export KYMA_VERSION=master
fi

export DOMAIN_NAME=$2


echo "We will instal Kyma version: ${KYMA_VERSION} on the cluster : ${DOMAIN_NAME}"

# We create the cluster
cat "${BASE_PATH}/devCluster.yaml" | envsubst | kubectl create -f -

# #we wait for the cluster to be ready
kubectl wait --for condition="ControlPlaneHealthy" --timeout=10m shoot ${DOMAIN_NAME}

# #we switch to the new cluster
kubectl get secrets $DOMAIN_NAME.kubeconfig -o jsonpath={.data.kubeconfig} | base64 -d > "${BASE_PATH}/kubeconfig--kyma--${DOMAIN_NAME}.yaml"
export KUBECONFIG="${BASE_PATH}/kubeconfig--kyma--${DOMAIN_NAME}.yaml"

# this is in teh case we want to use the new installer
kyma::install_cli

kyma alpha deploy \
--kubeconfig=${BASE_PATH}/kubeconfig--kyma--$DOMAIN_NAME.yaml \
--profile=evaluation \
--source=$KYMA_VERSION \
--value global.environment.gardener=true \
--workers-count=4

echo 'Kyma Console Url:'
echo 
kubectl get virtualservice console-web -n kyma-system -o jsonpath='{ .spec.hosts[0] }'
echo
echo 'User admin@kyma.cx'

