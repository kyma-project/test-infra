#!/usr/bin/env bash

set -o errexit
set -o pipefail

export KYMA_SOURCES_DIR="./kyma"
export LOCAL_KYMA_DIR="./local-kyma"

prereq_test() {
    command -v node >/dev/null 2>&1 || { echo >&2 "node not found"; exit 1; }
    command -v npm >/dev/null 2>&1 || { echo >&2 "npm not found"; exit 1; }
    command -v jq >/dev/null 2>&1 || { echo >&2 "jq not found"; exit 1; }
    command -v helm >/dev/null 2>&1 || { echo >&2 "helm not found"; exit 1; }
    command -v kubectl >/dev/null 2>&1 || { echo >&2 "kubectl not found"; exit 1; }
}

load_env() {
    ENV_FILE=".env"
    if [ -f "${ENV_FILE}" ]; then
        export $(xargs < "${ENV_FILE}")
    fi
}

prepare_k3s() {
    echo " --- Preparing k3s --- "

    pushd ${LOCAL_KYMA_DIR}
    # ./create-cluster-k3s.sh
    # copied here
    set -o errexit

    echo "-> Starting Docker registry"
    sudo mkdir -p /etc/rancher/k3s
    sudo cp registries.yaml /etc/rancher/k3s
    docker run -d \
    -p 5000:5000 \
    --restart=always \
    --name registry.localhost \
    -v $PWD/registry:/var/lib/registry \
    eu.gcr.io/kyma-project/test-infra/docker-registry-2:20200202

    echo "-> Starting cluster"
    curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION="v1.19.7+k3s1" K3S_KUBECONFIG_MODE=777 INSTALL_K3S_EXEC="server --disable traefik" sh -
    mkdir -p ~/.kube
    cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
    chmod 600 ~/.kube/config
    # end

    REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /registry.localhost)
    echo "Registry IP: ${REGISTRY_IP} registry.localhost" >> /etc/hosts
    echo "-> Kubernetes version:"
    kubectl version
    echo "get nodes:"
    kubectl get nodes
    echo "get pods:"
    kubectl get pods

    popd
}

host::os() {
  local host_os
  case "$(uname -s)" in
  Darwin)
    host_os=darwin
    ;;
  Linux)
    host_os=linux
    ;;
  *)
    echo >&2 -e "Unsupported host OS. Must be Linux or Mac OS X."
    exit 1
    ;;
  esac
  echo "${host_os}"
}

install_kyma_cli() {
  local settings
  local kyma_version
  mkdir -p "/usr/local/bin"
  os=$(host::os)

  pushd "/usr/local/bin" || exit

  echo "Install kyma CLI ${os} locally to /usr/local/bin..."

  curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}?alt=media"
  chmod +x kyma
  kyma_version=$(kyma version --client)
  echo "Kyma CLI version: ${kyma_version}"
  echo "OK"

  popd || exit

  eval "${settings}"
}

deploy_kyma() {
    echo "-> Starting Kyma deploy:"
    # kyma alpha deploy -p evaluation --component cluster-essentials,serverless --atomic --ci --value "$REGISTRY_VALUES" --value global.ingress.domainName="$DOMAIN" --value "serverless.webhook.values.function.resources.defaultPreset=M" -s local -w $KYMA_SOURCES_DIR
    # kyma alpha deploy --ci --profile "$executionProfile" --value global.isBEBEnabled=true --source=local --workspace "${kymaSourcesDir}" --verbose
    # kyma alpha deploy --ci --value global.isBEBEnabled=true --source=local --workspace "${KYMA_SOURCES_DIR}" --verbose
    kyma alpha deploy --ci --verbose
    # kyma alpha deploy --ci --components-file "$PWD/components.yaml" --value global.isBEBEnabled=true --source=local --workspace "${KYMA_SOURCES_DIR}" --verbose

    echo "Kyma deploy done"
    kubectl get pods
}

run_tests() {
    pushd "${KYMA_SOURCES_DIR}/tests/fast-integration"
    if [[ -v COMPASS_INTEGRATION_ENABLED && -v CENTRAL_APPLICATION_GATEWAY_ENABLED ]]; then
        make ci-application-connectivity-2-compass
    elif [[ -v COMPASS_INTEGRATION_ENABLED ]]; then
        make ci-compass
    elif [[ -v CENTRAL_APPLICATION_GATEWAY_ENABLED ]]; then
        make ci-application-connectivity-2
    else
        make ci
    fi
    popd
}

prereq_test
load_env
prepare_k3s
install_kyma_cli
deploy_kyma
run_tests
