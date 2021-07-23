#!/bin/bash
#
# Run fast-integration tests.
# Install k3s and Kyma CLI to provision Kyma on k3d cluster as prerequisite.

set -o errexit
set -o pipefail

export KYMA_SOURCES_DIR="./kyma"

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

install_k3s() {
  echo "Installing k3s..."
  # TODO pin version and explore flags
  curl -sfL https://get.k3s.io | sh -

  # echo "Setting kube config env var "
  # export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

  # echo "Setting kube config for outside access"
  # mkdir -p ~/.kube
  # cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
  # chmod 600 ~/.kube/config
}

# get_os() {
#   local host_os
#   case "$(uname -s)" in
#   Darwin)
#     host_os=darwin
#     ;;
#   Linux)
#     host_os=linux
#     ;;
#   *)
#     echo >&2 -e "Unsupported host OS. Must be Linux or Mac OS X."
#     exit 1
#     ;;
#   esac
#   echo "$host_os"
# }

install_cli() {
  echo "Installing Kyma CLI..."
  local install_dir
  declare -r install_dir="/usr/local/bin"
  # mkdir -p "/usr/local/bin"
  mkdir -p "$install_dir"

  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  if [[ -z "$os" || "$os" =~ ^(darwin|linux)$ ]]; then
    echo >&2 -e "Unsupported host OS. Must be Linux or Mac OS X."
    exit 1
  else
    readonly os
  fi

  pushd "$install_dir" || exit

  echo "Install Kyma CLI for ${os} in ${install_dir}"

  curl -Lo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}"
  chmod +x kyma
  # curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}?alt=media"
  # chmod +x kyma

  popd
  kyma version --client
}

deploy_kyma() {
  echo "Provisioning Kyma via k3d ..."
  
  # TODO pin version 
  kyma alpha provision k3s --ci

  # kyma alpha deploy -p evaluation --component cluster-essentials,serverless --atomic --ci --value "$REGISTRY_VALUES" --value global.ingress.domainName="$DOMAIN" --value "serverless.webhook.values.function.resources.defaultPreset=M" -s local -w $KYMA_SOURCES_DIR
  # kyma alpha deploy --ci --profile "$executionProfile" --value global.isBEBEnabled=true --source=local --workspace "${kymaSourcesDir}" --verbose
  # kyma alpha deploy --ci --value global.isBEBEnabled=true --source=local --workspace "${KYMA_SOURCES_DIR}" --verbose

  echo "Deploying Kyma..."
  kyma alpha deploy --ci --verbose --source=local --workspace "${KYMA_SOURCES_DIR}"
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
install_k3s
install_cli
deploy_kyma
run_tests
