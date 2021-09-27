#!/usr/bin/env bash

function integration_tests::install_kyma() {
  log::info "Installing Kyma from local source using components file"
  kyma deploy --ci --components-file "$PWD/components.yaml" --value global.isBEBEnabled=true --source=local --workspace "${KYMA_SOURCES_DIR}" --verbose
}

function api-gateway::prepare_components_file() {
  log::info "Preparing Kyma installation with Dex, Ory and API-Gateway"

cat << EOF > "$PWD/components.yaml"
defaultNamespace: kyma-system
prerequisites:
  - name: "cluster-essentials"
  - name: "istio"
    namespace: "istio-system"
  - name: "certificates"
    namespace: "istio-system"
components:
  - name: "dex"
  - name: "ory"
  - name: "api-gateway"
EOF
}

function api-gateway::prepare_test_environments() {
  log::info "Prepare test environment variables"

  # Preparing needed environment variables for API Gateway tests, these can be moved later on.
  export TEST_HYDRA_ADDRESS="https://oauth2.${CLUSTER_NAME}.kyma-prow.shoot.canary.k8s-hana.ondemand.com"
  TEST_USER_EMAIL="$(kubectl get secret -n kyma-system admin-user --template="{{.data.email}}" | base64 --decode)"
  export TEST_USER_EMAIL
  TEST_USER_PASSWORD="$(kubectl get secret -n kyma-system admin-user --template="{{.data.password}}" | base64 --decode)"
  export TEST_USER_PASSWORD
  export TEST_REQUEST_TIMEOUT="120"
  export TEST_REQUEST_DELAY="10"
  export INGRESSGATEWAY_ADDRESS="istio-ingressgateway.istio-system.svc.cluster.local"
  export TEST_DOMAIN="${CLUSTER_NAME}.kyma-prow.shoot.canary.k8s-hana.ondemand.com"
  export TEST_CLIENT_TIMEOUT=30s
  export TEST_RETRY_MAX_ATTEMPTS="5"
  export TEST_RETRY_DELAY="5"
  export TEST_GATEWAY_NAME="kyma-gateway"
  export TEST_GATEWAY_NAMESPACE="kyma-system"
}

function api-gateway::launch_tests() {
  log::info "Running Kyma API-Gateway tests"

  pushd "${KYMA_SOURCES_DIR}/tests/integration/api-gateway/gateway-tests"
  go test -v ./main_test.go
  popd

  log::success "Tests completed"
}
