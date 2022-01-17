#!/usr/bin/env bash

set -e

function integration_tests::install_kyma() {
  log::info "Installing Kyma from local source using components file"
  kyma deploy --ci --components-file "$PWD/components.yaml" --source=local --workspace "${KYMA_SOURCES_DIR}" --verbose
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
  - name: "ory"
  - name: "api-gateway"
EOF
}

function api-gateway::prepare_test_environments() {
  log::info "Prepare test environment variables"

  # Preparing needed environment variables for API Gateway tests, these can be moved later on.
  export TEST_HYDRA_ADDRESS="https://oauth2.${CLUSTER_NAME}.kyma-prow.shoot.canary.k8s-hana.ondemand.com"
  export TEST_REQUEST_TIMEOUT="120"
  export TEST_REQUEST_DELAY="10"
  export TEST_DOMAIN="${CLUSTER_NAME}.kyma-prow.shoot.canary.k8s-hana.ondemand.com"
  export TEST_CLIENT_TIMEOUT=30s
}

function api-gateway::configure_ory_hydra() {
  log::info "Prepare test environment variables"

  kubectl -n kyma-system set env deployment ory-hydra LOG_LEAK_SENSITIVE_VALUES="true"
  kubectl -n kyma-system set env deployment ory-hydra URLS_LOGIN="https://ory-hydra-login-consent.${CLUSTER_NAME}.kyma-prow.shoot.canary.k8s-hana.ondemand.com/login"
  kubectl -n kyma-system set env deployment ory-hydra URLS_CONSENT="https://ory-hydra-login-consent.${CLUSTER_NAME}.kyma-prow.shoot.canary.k8s-hana.ondemand.com/consent"
  kubectl -n kyma-system set env deployment ory-hydra URLS_SELF_ISSUER="https://oauth2.${CLUSTER_NAME}.kyma-prow.shoot.canary.k8s-hana.ondemand.com/"
  kubectl -n kyma-system set env deployment ory-hydra URLS_SELF_PUBLIC="https://oauth2.${CLUSTER_NAME}.kyma-prow.shoot.canary.k8s-hana.ondemand.com/"
  kubectl -n kyma-system rollout restart deployment ory-hydra
  kubectl -n kyma-system rollout status deployment ory-hydra
}

function api-gateway::deploy_login_consent_app() {
  log::info "Deploying Ory login consent app for tests"

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
  - ory-hydra-login-consent.${CLUSTER_NAME}.kyma-prow.shoot.canary.k8s-hana.ondemand.com
  http:
  - match:
    - uri:
        exact: /login
    - uri:
        exact: /consent
    route:
    - destination:
        host: ory-hydra-login-consent
        port:
          number: 80
EOF
  kubectl apply -f "$PWD/ory-hydra-login-consent.yaml"

  log::success "App deployed"
}

function api-gateway::launch_tests() {
  log::info "Running Kyma API-Gateway tests"

  pushd "${KYMA_SOURCES_DIR}/tests/integration/api-gateway/gateway-tests"
  go test -v ./main_test.go
  popd

  log::success "Tests completed"
}
