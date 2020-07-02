#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

discoverUnsetVar=false
for var in REPO_OWNER REPO_NAME DOCKER_PUSH_REPOSITORY KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_COMPUTE_ZONE CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS GARDENER_KYMA_PROW_PROJECT_NAME GARDENER_KYMA_PROW_KUBECONFIG GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME; do
  if [ -z "${!var}" ] ; then
    echo "ERROR: $var is not set"
    discoverUnsetVar=true
  fi
done
if [ "${discoverUnsetVar}" = true ] ; then
  exit 1
fi

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KCP_SOURCES_DIR="/home/prow/go/src/github.com/kyma-project/control-plane"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"

export GARDENER_PROJECT_NAME="${GARDENER_KYMA_PROW_PROJECT_NAME}"
export GARDENER_APPLICATION_CREDENTIALS="${GARDENER_KYMA_PROW_KUBECONFIG}"
export GARDENER_AZURE_SECRET_NAME="${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}"

# Enforce lowercase
readonly REPO_OWNER=$(echo "${REPO_OWNER}" | tr '[:upper:]' '[:lower:]')
export REPO_OWNER
readonly REPO_NAME=$(echo "${REPO_NAME}" | tr '[:upper:]' '[:lower:]')
export REPO_NAME

readonly CURRENT_TIMESTAMP=$(date +%Y%m%d)
readonly RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c10)

if [[ "$BUILD_TYPE" == "pr" ]]; then
  # In case of PR, operate on PR number
  readonly COMMON_NAME_PREFIX="gkekcpint-pr"
  COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}")
  KCP_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-control-plane-integration/${REPO_OWNER}/${REPO_NAME}:PR-${PULL_NUMBER}"
  export KCP_INSTALLER_IMAGE
else
  # Otherwise (master), operate on triggering commit id
  readonly COMMON_NAME_PREFIX="gkekcpint-commit"
  readonly COMMIT_ID=$(cd "$KCP_SOURCES_DIR" && git rev-parse --short HEAD)
  COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}")
  KCP_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-control-plane-integration/${REPO_OWNER}/${REPO_NAME}:COMMIT-${COMMIT_ID}"
  export KCP_INSTALLER_IMAGE
fi

readonly KCP_DEVELOPMENT_ARTIFACTS_BUCKET="${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/kcp"
if [[ "$BUILD_TYPE" == "pr" ]]; then
  KCP_VERSION="PR-${PULL_NUMBER}"
else
  KCP_VERSION="master-${COMMIT_ID}"
fi
readonly KCP_ARTIFACTS="${KCP_DEVELOPMENT_ARTIFACTS_BUCKET}/${KCP_VERSION}"

### Cluster name must be less than 40 characters!
COMMON_NAME=$(echo "${COMMON_NAME}" | tr "[:upper:]" "[:lower:]")
export CLUSTER_NAME="${COMMON_NAME}"

export GCLOUD_NETWORK_NAME="gke-long-lasting-net"
export GCLOUD_SUBNET_NAME="gke-long-lasting-subnet"

### For provision-gke-cluster.sh
export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"

#Local variables
DNS_SUBDOMAIN="${COMMON_NAME}"
KCP_SCRIPTS_DIR="${KCP_SOURCES_DIR}/installation/scripts"
KCP_RESOURCES_DIR="${KCP_SOURCES_DIR}/installation/resources"

INSTALLER_YAML="${KCP_RESOURCES_DIR}/installer.yaml"
INSTALLER_CR="${KCP_RESOURCES_DIR}/installer-cr.yaml.tpl"

function cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        shout "AN ERROR OCCURED! Take a look at preceding log entries."
        echo
    fi

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    if [ -n "${CLEANUP_CLUSTER}" ]; then
        shout "Deprovision cluster: \"${CLUSTER_NAME}\""
        date

        #save disk names while the cluster still exists to remove them later
        DISKS=$(kubectl get pvc --all-namespaces -o jsonpath="{.items[*].spec.volumeName}" | xargs -n1 echo)
        export DISKS

        #Delete cluster
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/deprovision-gke-cluster.sh"

        #Delete orphaned disks
        shout "Delete orphaned PVC disks..."
        date
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/delete-disks.sh"
    fi

    if [ -n "${CLEANUP_GATEWAY_DNS_RECORD}" ]; then
        shout "Delete Gateway DNS Record"
        date
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${GATEWAY_DNS_FULL_NAME}" --address="${GATEWAY_IP_ADDRESS}" --dryRun=false
    fi

    if [ -n "${CLEANUP_GATEWAY_IP_ADDRESS}" ]; then
        shout "Release Gateway IP Address"
        date
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/release-ip-address.sh --project="${CLOUDSDK_CORE_PROJECT}" --ipname="${GATEWAY_IP_ADDRESS_NAME}" --region="${CLOUDSDK_COMPUTE_REGION}" --dryRun=false
    fi


    if [ -n "${CLEANUP_DOCKER_IMAGE}" ]; then
        shout "Delete temporary KCP-Installer Docker image"
        date
        KYMA_INSTALLER_IMAGE="${KCP_INSTALLER_IMAGE}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/delete-image.sh"
    fi

    if [ -n "${CLEANUP_APISERVER_DNS_RECORD}" ]; then
        shout "Delete Apiserver proxy DNS Record"
        date
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${APISERVER_DNS_FULL_NAME}" --address="${APISERVER_IP_ADDRESS}" --dryRun=false
    fi

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    shout "Job is finished ${MSG}"
    date
    set -e

    exit "${EXIT_STATUS}"
}

function createCluster() {
  #Used to detect errors for logging purposes
  ERROR_LOGGING_GUARD="true"

  shout "Authenticate"
  date
  init
  DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"
  export DNS_DOMAIN
  DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
  export DOMAIN

  shout "Reserve IP Address for Ingressgateway"
  date
  GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
  GATEWAY_IP_ADDRESS=$(IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/reserve-ip-address.sh)
  CLEANUP_GATEWAY_IP_ADDRESS="true"
  echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

  shout "Create DNS Record for Ingressgateway IP"
  date
  GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
  CLEANUP_GATEWAY_DNS_RECORD="true"
  IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-dns-record.sh

  NETWORK_EXISTS=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/network-exists.sh")
  if [ "$NETWORK_EXISTS" -gt 0 ]; then
    shout "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
    date
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-network-with-subnet.sh"
  else
    shout "Network ${GCLOUD_NETWORK_NAME} exists"
  fi

  shout "Provision cluster: \"${CLUSTER_NAME}\""
  date
  export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
  if [ -z "$MACHINE_TYPE" ]; then
    export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
  fi
  if [ -z "${CLUSTER_VERSION}" ]; then
    export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
  fi
  CLEANUP_CLUSTER="true"
  env ADDITIONAL_LABELS="created-at=${CURRENT_TIMESTAMP}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/provision-gke-cluster.sh
}

function applyKymaOverrides() {
  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "application-resource-tests-overrides" \
    --data "application-operator.tests.enabled=false" \
    --data "tests.application_connector_tests.enabled=false" \
    --data "application-registry.tests.enabled=false" \
    --data "console-backend-service.tests.enabled=false" \
    --data "test.acceptance.service-catalog.enabled=false" \
    --data "test.acceptance.external_solution.enabled=false" \
    --data "console.test.acceptance.enabled=false" \
    --data "test.external_solution.event_mesh.enabled=false"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "core-test-ui-acceptance-overrides" \
    --data "test.acceptance.ui.logging.enabled=true" \
    --label "component=core"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "application-registry-overrides" \
    --data "application-registry.deployment.args.detailedErrorResponse=true" \
    --label "component=application-connector"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "monitoring-config-overrides" \
    --data "global.alertTools.credentials.slack.channel=${KYMA_ALERTS_CHANNEL}" \
    --data "global.alertTools.credentials.slack.apiurl=${KYMA_ALERTS_SLACK_API_URL}" \
    --data "pushgateway.enabled=true" \
    --label "component=monitoring"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "istio-overrides" \
    --data "gateways.istio-ingressgateway.loadBalancerIP=${GATEWAY_IP_ADDRESS}" \
    --label "component=istio"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "api-gateway-overrides" \
    --data "tests.env.gatewayName=compass-istio-gateway" \
    --data "tests.env.gatewayNamespace=compass-system" \
    --label "component=api-gateway"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "dex-overrides" \
    --data "global.istio.gateway.name=compass-istio-gateway" \
    --data "global.istio.gateway.namespace=compass-system" \
    --label "component=dex"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "ory-overrides" \
    --data "global.istio.gateway.name=compass-istio-gateway" \
    --data "global.istio.gateway.namespace=compass-system" \
    --label "component=ory"
}

function applyCompassOverrides() {
  NAMESPACE="compass-installer"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "compass-auditlog-mock-tests" \
    --data "global.externalServicesMock.enabled=true" \
    --data "gateway.gateway.auditlog.enabled=true" \
    --data "gateway.gateway.auditlog.authMode=oauth" \
    --label "component=compass"
}

function applyKebResources() {

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: kcp-system
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: auditlog-script
  namespace: kcp-system
data:
  script: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: kcp-auditlog-secret
  namespace: kcp-system
type: Opaque
data:
  auditlog-user: "dXNyCg=="
  auditlog-password: "dXNyCg=="
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kcp-auditlog-config
  namespace: kcp-system
data:
  auditlog-url: "http://dummy.url"
  auditlog-config-path: "/path"
  auditlog-security-path: "/path"
  auditlog-tenant: "tnt"
EOF
}

function applyControlPlaneOverrides() {
  NAMESPACE="kcp-installer"

  #Create Config map for Provisioner
  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "provisioner-overrides" \
    --data "global.provisioning.enabled=true" \
    --data "provisioner.security.skipTLSCertificateVeryfication=true" \
    --data "provisioner.gardener.kubeconfig=$(base64 -w 0 < "${GARDENER_APPLICATION_CREDENTIALS}")" \
    --data "provisioner.gardener.project=$GARDENER_PROJECT_NAME" \
    --label "component=kcp"

  if [ "${RUN_PROVISIONER_TESTS}" == "true" ]; then
    # Change timeout for kyma test to 3h
    export KYMA_TEST_TIMEOUT=3h

    # Create Config map for Provisioner Tests
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "provisioner-tests-overrides" \
      --data "provisioner.tests.enabled=true" \
      --data "provisioner.tests.gardener.azureSecret=$GARDENER_AZURE_SECRET_NAME" \
      --label "component=kcp"
  fi

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "compass-auditlog-mock-tests" \
    --data "global.externalServicesMock.enabled=true" \
    --data "gateway.gateway.auditlog.enabled=true" \
    --data "gateway.gateway.auditlog.authMode=oauth" \
    --label "component=compass"

  applyKebResources
}

function applyCommonOverrides() {
  NAMESPACE=${1}

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "installation-config-overrides" \
    --data "global.domainName=${DOMAIN}" \
    --data "global.loadBalancerIP=${GATEWAY_IP_ADDRESS}"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "feature-flags-overrides" \
    --data "global.enableAPIPackages=true" \
    --data "global.disableLegacyConnectivity=true"


  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "cluster-certificate-overrides" \
    --data "global.tlsCrt=${TLS_CERT}" \
    --data "global.tlsKey=${TLS_KEY}"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --namespace "${NAMESPACE}" --name "global-ingress-overrides" \
    --data "global.ingress.domainName=${DOMAIN}" \
    --data "global.ingress.tlsCrt=${TLS_CERT}" \
    --data "global.ingress.tlsKey=${TLS_KEY}" \
    --data "global.environment.gardener=false"
}

function installKyma() {
  kubectl create namespace "kyma-installer"
  applyCommonOverrides "kyma-installer"
  applyKymaOverrides
  
  TMP_DIR="/tmp/kcp-gke-integration"

  gsutil cp "${KCP_ARTIFACTS}/kyma-installer.yaml" ${TMP_DIR}/kyma-installer.yaml
  gsutil cp "${KCP_ARTIFACTS}/is-kyma-installed.sh" ${TMP_DIR}/is-kyma-installed.sh
  chmod +x ${TMP_DIR}/is-kyma-installed.sh
  kubectl apply -f ${TMP_DIR}/kyma-installer.yaml

  shout "Installation triggered"
  date
  "${TMP_DIR}"/is-kyma-installed.sh --timeout 30m
}

function installCompass() {
  kubectl create namespace "compass-installer"
  applyCommonOverrides "compass-installer"
  applyCompassOverrides

  TMP_DIR="/tmp/kcp-gke-integration"

  gsutil cp "${KCP_ARTIFACTS}/compass-installer.yaml" ${TMP_DIR}/compass-installer.yaml
  gsutil cp "${KCP_ARTIFACTS}/is-compass-installed.sh" ${TMP_DIR}/is-compass-installed.sh
  chmod +x ${TMP_DIR}/is-compass-installed.sh
  kubectl apply -f ${TMP_DIR}/compass-installer.yaml

  shout "Installation triggered"
  date
  "${TMP_DIR}"/is-compass-installed.sh --timeout 30m
}

function installControlPlane() {
  kcpUnsetVar=false

  # shellcheck disable=SC2043
  for var in GATEWAY_IP_ADDRESS ; do
    if [ -z "${!var}" ] ; then
      echo "ERROR: $var is not set"
      kcpUnsetVar=true
    fi
  done
  if [ "${kcpUnsetVar}" = true ] ; then
    exit 1
  fi

  shout "Build Kyma Control Plane Installer Docker image"
  date
  CLEANUP_DOCKER_IMAGE="true"
  KCP_INSTALLER_IMAGE="${KCP_INSTALLER_IMAGE}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-control-plane-image.sh

  shout "Apply Kyma Control Plane config"
  date
  kubectl create namespace "kcp-installer"
  applyCommonOverrides "kcp-installer"
  applyControlPlaneOverrides

  echo "Manual concatenating yamls"
  "${KCP_SCRIPTS_DIR}"/concat-yamls.sh "${INSTALLER_YAML}" "${INSTALLER_CR}" \
  | sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KCP_INSTALLER_IMAGE};" \
  | sed -e "s/__VERSION__/0.0.1/g" \
  | sed -e "s/__.*__//g" \
  | kubectl apply -f-
  
  shout "Installation triggered"
  date
  "${KCP_SCRIPTS_DIR}"/is-installed.sh --timeout 30m

  if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
    shout "Create DNS Record for Apiserver proxy IP"
    date
    APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
    CLEANUP_APISERVER_DNS_RECORD="true"
    IP_ADDRESS=${APISERVER_IP_ADDRESS} DNS_FULL_NAME=${APISERVER_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
  fi
}

trap cleanup EXIT INT

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    shout "Execute Job Guard"
    export JOB_NAME_PATTERN="(pre-control-plane-components-.*)|(pre-control-plane-tests-.*)"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

shout "Create new cluster"
date
createCluster

shout "Install tiller"
date
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
"${KCP_SCRIPTS_DIR}"/install-tiller.sh

shout "Generate self-signed certificate"
date
CERT_KEY=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-self-signed-cert.sh")
TLS_CERT=$(echo "${CERT_KEY}" | head -1)
TLS_KEY=$(echo "${CERT_KEY}" | tail -1)

shout "Install Kyma"
date
installKyma

shout "Install Compass"
date
installCompass


shout "Install Control Plane"
date
installControlPlane

shout "Test Kyma, Compass and Control Plane"
date
"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh

shout "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
