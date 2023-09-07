#!/usr/bin/env bash

set -e

function integration_tests::install_kyma() {
  log::info "Installing Kyma from local source using components file"
  kyma deploy --ci --components-file "$PWD/components.yaml" --source=local --workspace "${KYMA_SOURCES_DIR}" --verbose
}

function api-gateway::prepare_components_file() {
  log::info "Preparing Kyma installation with Ory and API-Gateway"

cat << EOF > "$PWD/components.yaml"
defaultNamespace: kyma-system
prerequisites:
  - name: "cluster-essentials"
  - name: "istio"
    namespace: "istio-system"
  - name: "certificates"
    namespace: "istio-system"
components:
  - name: "istio-resources"
  - name: "ory"
  - name: "api-gateway"
EOF
}

function api-gateway::prepare_components_file_istio_only() {
  log::info "Preparing Kyma installation with Istio and API-Gateway"

cat << EOF > "$PWD/components.yaml"
defaultNamespace: kyma-system
prerequisites:
  - name: "cluster-essentials"
  - name: "istio"
    namespace: "istio-system"
  - name: "certificates"
    namespace: "istio-system"
components:
  - name: "istio-resources"
  - name: "api-gateway"
  - name: "ory" # Until drop of ory oathkeeper Ory needs to be deployed for noop and OAuth2 scenarios
EOF
}

function api-gateway::prepare_test_environments() {
  log::info "Prepare test environment variables"

  # Preparing needed environment variables for API Gateway tests, these can be moved later on.
  export TEST_HYDRA_ADDRESS="https://oauth2.${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.live.k8s-hana.ondemand.com"
  export TEST_REQUEST_TIMEOUT="120"
  export TEST_REQUEST_DELAY="10"
  export TEST_DOMAIN="${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.live.k8s-hana.ondemand.com"
  export TEST_CLIENT_TIMEOUT=30s
  export TEST_CONCURENCY="8"
  export EXPORT_RESULT="true"
  export KYMA_DOMAIN="${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.live.k8s-hana.ondemand.com"
}

function api-gateway::prepare_test_env_integration_tests() {
  log::info "Prepare test environment variables for integration tests"
  export KYMA_DOMAIN="${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.live.k8s-hana.ondemand.com"
}

function api-gateway::configure_ory_hydra() {
  log::info "Prepare test environment variables"

  kubectl -n kyma-system set env deployment ory-hydra LOG_LEAK_SENSITIVE_VALUES="true"
  kubectl -n kyma-system set env deployment ory-hydra URLS_LOGIN="https://ory-hydra-login-consent.${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.live.k8s-hana.ondemand.com/login"
  kubectl -n kyma-system set env deployment ory-hydra URLS_CONSENT="https://ory-hydra-login-consent.${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.live.k8s-hana.ondemand.com/consent"
  kubectl -n kyma-system set env deployment ory-hydra URLS_SELF_ISSUER="https://oauth2.${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.live.k8s-hana.ondemand.com/"
  kubectl -n kyma-system set env deployment ory-hydra URLS_SELF_PUBLIC="https://oauth2.${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.live.k8s-hana.ondemand.com/"
  kubectl -n kyma-system scale deployment.apps ory-hydra --replicas=1
  kubectl -n kyma-system rollout restart deployment ory-hydra
  kubectl -n kyma-system rollout status deployment ory-hydra
}

function api-gateway::deploy_login_consent_app() {
  log::info "Deploying Ory login consent app for tests"

  kubectl -n istio-system rollout status deployment istiod
  kubectl -n istio-system rollout status deployment istio-ingressgateway

cat << EOF > "$PWD/ory-hydra-login-consent.yaml"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ory-hydra-login-consent
  namespace: kyma-system
spec:
  selector:
    matchLabels:
      app: ory-hydra-login-consent
      version: v1
  template:
    metadata:
      labels:
        app: ory-hydra-login-consent
        version: v1
    spec:
      containers:
        - name: login-consent
          image: ${TEST_ORY_IMAGE}
          env:
            - name: HYDRA_ADMIN_URL
              value: http://ory-hydra-admin.kyma-system.svc.cluster.local:4445
            - name: BASE_URL
              value: ""
            - name: PORT
              value: "3000"
          ports:
          - containerPort: 3000
---
kind: Service
apiVersion: v1
metadata:
  name: ory-hydra-login-consent
  namespace: kyma-system
spec:
  selector:
    app: ory-hydra-login-consent
    version: v1
  ports:
    - name: http-login-consent
      protocol: TCP
      port: 80
      targetPort: 3000
---
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: ory-hydra-login-consent
  namespace: kyma-system
  labels:
    app: ory-hydra-login-consent
spec:
  gateways:
  - kyma-system/kyma-gateway
  hosts:
  - ory-hydra-login-consent.${CLUSTER_NAME}.kyma-prow.shoot.live.k8s-hana.ondemand.com
  http:
  - match:
    - uri:
        exact: /login
    - uri:
        exact: /consent
    route:
    - destination:
        host: ory-hydra-login-consent.kyma-system.svc.cluster.local
        port:
          number: 80
EOF
  kubectl wait deployment/istiod -n istio-system --timeout=60s --for condition=available
  kubectl apply -f "$PWD/ory-hydra-login-consent.yaml"
  kubectl wait deployment ory-hydra-login-consent -n kyma-system --timeout=60s --for condition=available
  log::success "App deployed"
}

function api-gateway::launch_tests() {
  log::info "Running Kyma API-Gateway tests"
  kubectl get validatingwebhookconfigurations
  pushd "${KYMA_SOURCES_DIR}/tests/components/api-gateway"
  make test
  popd

  log::success "Tests completed"
}

function api-gateway::launch_integration_tests() {
  log::info "Running API-Gateway integration tests"
  pushd "${API_GATEWAY_SOURCES_DIR}"
  make install-kyma
  make test-integration
  popd

  log::success "Tests completed"
}

function istio::get_version() {
  pushd "${KYMA_SOURCES_DIR}"
  istio_version=$(git show "${KYMA_VERSION}:resources/istio/Chart.yaml" | grep appVersion | sed -n "s/appVersion: //p")
  popd
}
