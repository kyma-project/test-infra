#!/usr/bin/env bash

set -e

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on a real Gardener cluster.
#
#Expected common vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - GARDENER_REGION - Gardener compute region
# - GARDENER_ZONES - Gardener compute zones inside the region
# - GARDENER_CLUSTER_VERSION - Version of the Kubernetes cluster
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME - Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME - Name of the secret configured in the gardener project to access the cloud provider
# - MACHINE_TYPE - (optional) machine type
#
#Please look in each provider script for provider specific requirements

function delete_cluster(){
    local name="$1"
    set +e
    log::info "Deleting cluster '${name}'"
    kubectl annotate shoot "${name}" confirmation.gardener.cloud/deletion=true --overwrite
    kubectl delete shoot "${name}" --wait=true
    set -e
}

function provisionCluster() {
    export DOMAIN_NAME=$1
    export DEFINITION_PATH=$2

    log::info "Creating cluster: ${DOMAIN_NAME}"
    # create the cluster
    # shellcheck disable=SC2002
    cat "${DEFINITION_PATH}" | envsubst | kubectl create -f -

    # wait for the cluster to be ready
    kubectl wait --for condition="ControlPlaneHealthy" --timeout=10m shoot "${DOMAIN_NAME}"
    log::info "Cluster ${DOMAIN_NAME} was created succesfully"
}

function provisionIngress() {
    export DOMAIN_NAME=$1

    log::info "Install ingress"

    # switch to the new cluster
    kubectl get secrets "${DOMAIN_NAME}.kubeconfig" -o jsonpath="{.data.kubeconfig}" | base64 -d > "${RESOURCES_PATH}/kubeconfig--busola--${DOMAIN_NAME}.yaml"
    export KUBECONFIG="${RESOURCES_PATH}/kubeconfig--busola--$DOMAIN_NAME.yaml"

    # ask for new certificates
    # shellcheck disable=SC2002
    cat "${RESOURCES_PATH}/wildcardCert.yaml" | envsubst | kubectl apply -f -

    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
    helm repo update

    # shellcheck disable=SC2002
    cat "${RESOURCES_PATH}/nginxValues.yaml" | envsubst | helm install ingress-nginx --namespace=kube-system -f - ingress-nginx/ingress-nginx

    # wait for ingress controller to start
    kubectl wait --namespace kube-system \
      --for=condition=ready pod \
      --selector=app.kubernetes.io/component=controller \
      --timeout=120s
    
    log::info "Ingress is ready"
}

function provisionBusola() {
    export DOMAIN_NAME=$1

    busola_namespace="busola"

    log::info "Installing Busola on the cluster: ${DOMAIN_NAME}"

    export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"
    kubectl get secrets "${DOMAIN_NAME}.kubeconfig" -o jsonpath="{.data.kubeconfig}" | base64 -d > "${RESOURCES_PATH}/kubeconfig--busola--${DOMAIN_NAME}.yaml"
    export KUBECONFIG="${RESOURCES_PATH}/kubeconfig--busola--$DOMAIN_NAME.yaml"

    # delete old installation
    namespace_exists=$(kubectl get ns -o json | jq -r ".items | .[] | .metadata | select(.name == \"$busola_namespace\") | .name")
    if [[ "$namespace_exists" == "$busola_namespace" ]]; then
        log::info "namespace busola exists, deleting..."
        kubectl delete ns "$busola_namespace" --wait=true
    fi

    log::info "Busola installation started"

    # install busola
    FULL_DOMAIN="${DOMAIN_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.canary.k8s-hana.ondemand.com"

    find "${BUSOLA_SOURCES_DIR}/resources" -name "*.yaml" \
         -exec sed -i "s/%DOMAIN%/${FULL_DOMAIN}/g" "{}" \;

    kubectl create namespace "$busola_namespace"
    kubectl apply --namespace "$busola_namespace" -k "${BUSOLA_SOURCES_DIR}/resources"

    TERM=dumb kubectl cluster-info
    log::info "Please generate params for using k8s http://enkode.surge.sh/"
    log::info "Kyma busola Url:"
    log::info "https://busola.${DOMAIN_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.canary.k8s-hana.ondemand.com?auth=generated_params_in_previous_step"
}

function provisionKyma2(){
    export KYMA_VERSION=$1
    export DOMAIN_NAME=$2

    log::info "Installing Kyma version: ${KYMA_VERSION} on the cluster : ${DOMAIN_NAME}"

    # switch to the new cluster
    export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"
    kubectl get secrets "${DOMAIN_NAME}.kubeconfig" -o jsonpath="{.data.kubeconfig}" | base64 -d > "${RESOURCES_PATH}/kubeconfig--kyma--${DOMAIN_NAME}.yaml"
    export KUBECONFIG="${RESOURCES_PATH}/kubeconfig--kyma--${DOMAIN_NAME}.yaml"

    kyma::install_cli

    TERM=dumb kyma alpha deploy \
    --kubeconfig="${RESOURCES_PATH}/kubeconfig--kyma--${DOMAIN_NAME}.yaml" \
    --profile=evaluation \
    --source="${KYMA_VERSION}" \
    --value global.environment.gardener=true \
    --concurrency=4
}

ENABLE_TEST_LOG_COLLECTOR=false

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export BUSOLA_SOURCES_DIR="${KYMA_PROJECT_DIR}/busola"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"

# All provides require these values, each of them may check for additional variables
requiredVars=(
    GARDENER_PROVIDER
    KYMA_PROJECT_DIR
    GARDENER_REGION
    GARDENER_ZONES
    GARDENER_CLUSTER_VERSION
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
    BUSOLA_PROVISION_TYPE
)

utils::check_required_vars "${requiredVars[@]}"

if [[ $GARDENER_PROVIDER == "gcp" ]]; then
    # shellcheck source=prow/scripts/lib/gardener/gcp.sh
    #source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gcp.sh"
    log::info "Provisioning on gcp"
else
    ## TODO what should I put here? Is this a backend?
    log::error "GARDENER_PROVIDER ${GARDENER_PROVIDER} is not yet supported"
    exit 1
fi

readonly COMMON_NAME_PREFIX="n"
readonly KYMA_NAME_SUFFIX="kyma"
readonly BUSOLA_NAME_SUFFIX="busola"

RESOURCES_PATH="${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/busola/"
KYMA_COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${KYMA_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
BUSOLA_COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${BUSOLA_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")


export KUBECONFIG="${GARDENER_KYMA_PROW_KUBECONFIG}"

if [[ $BUSOLA_PROVISION_TYPE == "KYMA" ]]; then
    log::info "Kyma cluster name: ${KYMA_COMMON_NAME}"
    delete_cluster "${KYMA_COMMON_NAME}"
    provisionCluster "${KYMA_COMMON_NAME}" "${RESOURCES_PATH}/cluster-kyma.yaml"
    provisionKyma2 "main" "${KYMA_COMMON_NAME}"
elif [[ $BUSOLA_PROVISION_TYPE == "BUSOLA" ]]; then
    log::info "Busola cluster name: ${BUSOLA_COMMON_NAME}"
    if [[ $RECREATE_CLUSTER == "true" ]]; then
        delete_cluster "${BUSOLA_COMMON_NAME}"
        provisionCluster "${BUSOLA_COMMON_NAME}" "${RESOURCES_PATH}/cluster-busola.yaml"
        provisionIngress "${BUSOLA_COMMON_NAME}"
    fi
    provisionBusola "${BUSOLA_COMMON_NAME}"
else
    log::error "Wrong value for BUSOLA_PROVISION_TYPE: '$BUSOLA_PROVISION_TYPE'"
    exit 1
fi
