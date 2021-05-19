#!/usr/bin/env bash

# Check if required vars are set or not
function eventing::check_required_vars() {
  if [[ -z ${CREDENTIALS_DIR} ]]; then
    echo "required variable CREDENTIALS_DIR is missing"
    exit 1
  fi
}

# Create a Kubernetes Secret which contains the EventMesh service key
function eventing::create_eventmesh_secret() {
  eventing::check_required_vars

  pushd "${CREDENTIALS_DIR}"

  SECRET_NAME=event-mesh
  SECRET_NAMESPACE=default

  SERVICE_KEY_VALUE=$(base64 -i serviceKey | tr -d '\n')

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: ${SECRET_NAME}
  namespace: ${SECRET_NAMESPACE}
  labels:
    kyma-project.io/event-mesh: "true"
data:
  serviceKey: "${SERVICE_KEY_VALUE}"
EOF

  popd
}

# Create a Kubernetes Secret which is needed by the Eventing Backend controller
function eventing::create_eventing_backend_secret() {
  eventing::check_required_vars

  pushd "${CREDENTIALS_DIR}"

  SECRET_NAME=eventing-backend
  SECRET_NAMESPACE=default

  MANAGEMENT=$(jq -r  '.management' < serviceKey | tr -d '[:space:]' | base64 | tr -d '\n')
  MESSAGING=$(jq -r '.messaging' < serviceKey | tr -d '[:space:]' | base64 | tr -d '\n')
  NAMESPACE=$(jq -r '.namespace' < serviceKey | tr -d '[:space:]' | base64 | tr -d '\n')
  SERVICE_INSTANCE_ID=$(jq -r '.serviceinstanceid' < serviceKey | tr -d '[:space:]' | base64 | tr -d '\n')
  XS_APP_NAME=$(jq -r '.xsappname' < serviceKey | tr -d '[:space:]' | base64 | tr -d '\n')

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: ${SECRET_NAME}
  namespace: ${SECRET_NAMESPACE}
data:
  management: "${MANAGEMENT}"
  messaging: "${MESSAGING}"
  namespace: "${NAMESPACE}"
  serviceinstanceid: "${SERVICE_INSTANCE_ID}"
  xsappname: "${XS_APP_NAME}"
EOF

  popd
}

# Create a Kubernetes Secret which is needed by the Eventing Publisher and Subscription Controller
function eventing::create_eventing_secret() {
  eventing::check_required_vars

  pushd "${CREDENTIALS_DIR}"

  SECRET_NAME=eventing
  SECRET_NAMESPACE=kyma-system

  # delete the default Eventing secret
  kubectl delete secret -n ${SECRET_NAMESPACE} ${SECRET_NAME}

  HTTP_REST=$(jq -r '.messaging' < serviceKey | jq -c '.[] | select(.broker.type | contains("saprestmgw"))')
  BEB_NAMESPACE=$(jq -r '.namespace' < serviceKey | tr -d '[:space:]' | base64 | tr -d '\n')
  CLIENT_ID=$(echo "$HTTP_REST" | jq -r '.oa2.clientid' | tr -d '[:space:]' | base64 | tr -d '\n')
  CLIENT_SECRET=$(echo "$HTTP_REST" | jq -r '.oa2.clientsecret' | tr -d '[:space:]' | base64 | tr -d '\n')
  EMS_PUBLISH_URL=$(echo "$HTTP_REST" | jq -r '.uri' | tr -d '[:space:]' | base64 | tr -d '\n')
  TOKEN_ENDPOINT=$(echo "$HTTP_REST" | jq -r '.oa2.tokenendpoint' | tr -d '[:space:]' | base64 | tr -d '\n')

  # create Eventing secret with the proper values
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: ${SECRET_NAME}
  namespace: ${SECRET_NAMESPACE}
data:
  beb-namespace: "${BEB_NAMESPACE}"
  client-id: "${CLIENT_ID}"
  client-secret: "${CLIENT_SECRET}"
  ems-publish-url: "${EMS_PUBLISH_URL}"
  token-endpoint: "${TOKEN_ENDPOINT}"
EOF

  popd
}
