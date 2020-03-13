#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

discoverUnsetVar=false
for var in DOCKER_PUSH_REPOSITORY DOCKER_PUSH_DIRECTORY KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_COMPUTE_ZONE CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS GARDENER_KYMA_PROW_PROJECT_NAME GARDENER_KYMA_PROW_KUBECONFIG GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME; do
  if [ -z "${!var}" ] ; then
    echo "ERROR: $var is not set"
    discoverUnsetVar=true
  fi
done
if [ "${discoverUnsetVar}" = true ] ; then
  exit 1
fi

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"

export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"
export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"

export GARDENER_PROJECT_NAME="${GARDENER_KYMA_PROW_PROJECT_NAME}"
export GARDENER_APPLICATION_CREDENTIALS="${GARDENER_KYMA_PROW_KUBECONFIG}"
export GARDENER_AZURE_SECRET_NAME="${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}"

readonly REPO_OWNER="kyma-project"
readonly REPO_NAME="kyma"
readonly CURRENT_TIMESTAMP=$(date +%Y%m%d)

readonly RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c10)

if [[ "$BUILD_TYPE" == "pr" ]]; then
    # In case of PR, operate on PR number
    readonly COMMON_NAME_PREFIX="gkecompint-pr"
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}")
    KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-compass-integration/${REPO_OWNER}/${REPO_NAME}:PR-${PULL_NUMBER}"
    export KYMA_INSTALLER_IMAGE
elif [[ "$BUILD_TYPE" == "release" ]]; then
    readonly COMMON_NAME_PREFIX="gkecompint-rel"
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${RANDOM_NAME_SUFFIX}")
    readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
    readonly RELEASE_VERSION=$(cat "${SCRIPT_DIR}/../../RELEASE_VERSION")
    shout "Reading release version from RELEASE_VERSION file, got: ${RELEASE_VERSION}"
else
    # Otherwise (master), operate on triggering commit id
    readonly COMMON_NAME_PREFIX="gkecompint-commit"
    readonly COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)
    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}")
    KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-compass-integration/${REPO_OWNER}/${REPO_NAME}:COMMIT-${COMMIT_ID}"
    export KYMA_INSTALLER_IMAGE
fi

readonly STANDARIZED_NAME=$(echo "${COMMON_NAME}" | tr "[:upper:]" "[:lower:]")
readonly DNS_SUBDOMAIN="${STANDARIZED_NAME}"

export CLUSTER_NAME="${STANDARIZED_NAME}"
export GCLOUD_NETWORK_NAME="gke-long-lasting-net"
export GCLOUD_SUBNET_NAME="gke-long-lasting-subnet"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

function removeCluster() {
  #Turn off exit-on-error so that next step is executed even if previous one fails.
  set +e

  # CLUSTER_NAME variable is used in other scripts so we need to change it for a while
  ORIGINAL_CLUSTER_NAME=${CLUSTER_NAME}
  CLUSTER_NAME=$1

  EXIT_STATUS=$?

  shout "Fetching OLD_TIMESTAMP from cluster to be deleted"
  readonly OLD_TIMESTAMP=$(gcloud container clusters describe "${CLUSTER_NAME}" --zone="${GCLOUD_COMPUTE_ZONE}" --project="${GCLOUD_PROJECT_NAME}" --format=json | jq --raw-output '.resourceLabels."created-at-readable"')

  shout "Delete cluster $CLUSTER_NAME"
  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/deprovision-gke-cluster.sh
  TMP_STATUS=$?
  if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

  shout "Delete Gateway DNS Record"
  date
  GATEWAY_IP_ADDRESS=$(gcloud compute addresses describe "${CLUSTER_NAME}" --format json --region "${CLOUDSDK_COMPUTE_REGION}" | jq '.address' | tr -d '"')
  GATEWAY_DNS_FULL_NAME="*.${CLUSTER_NAME}.${DNS_DOMAIN}"
  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${GATEWAY_DNS_FULL_NAME}" --address="${GATEWAY_IP_ADDRESS}" --dryRun=false
  TMP_STATUS=$?
  if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

  shout "Release Gateway IP Address"
  date
  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/release-ip-address.sh --project="${CLOUDSDK_CORE_PROJECT}" --region="${CLOUDSDK_COMPUTE_REGION}" --ipname="${CLUSTER_NAME}" --dryRun=false
  TMP_STATUS=$?
  if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

  echo "Remove DNS Record for Apiserver Proxy IP"
  APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
  APISERVER_IP_ADDRESS=$(gcloud dns record-sets list --zone "${CLOUDSDK_DNS_ZONE_NAME}" --name "${APISERVER_DNS_FULL_NAME}" --format="value(rrdatas[0])")
  if [[ -n ${APISERVER_IP_ADDRESS} ]]; then
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${APISERVER_DNS_FULL_NAME}" --address="${APISERVER_IP_ADDRESS}" --dryRun=false
    TMP_STATUS=$?
    if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
  fi

  shout "Delete temporary Kyma-Installer Docker image"
  date

  KYMA_INSTALLER_IMAGE="${KYMA_INSTALLER_IMAGE}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-image.sh
  TMP_STATUS=$?
  if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

  MSG=""
  if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
  shout "Job is finished ${MSG}"
  date

  # Revert previous value for CLUSTER_NAME variable
  CLUSTER_NAME=${ORIGINAL_CLUSTER_NAME}
  set -e
}

function createCluster() {
  shout "Reserve IP Address for Ingressgateway"
  date
  GATEWAY_IP_ADDRESS_NAME="${STANDARIZED_NAME}"
  GATEWAY_IP_ADDRESS=$(IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/reserve-ip-address.sh)
  echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

  shout "Create DNS Record for Ingressgateway IP"
  date
  GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
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

  if [ -z "$MACHINE_TYPE" ]; then
    export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
  fi
  if [ -z "${CLUSTER_VERSION}" ]; then
    export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
  fi
  env ADDITIONAL_LABELS="created-at=${CURRENT_TIMESTAMP}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/provision-gke-cluster.sh
}

function waitUntilInstallerApiAvailable() {
  shout "Waiting for Installer API"

  attempts=5
  for ((i=1; i<=attempts; i++)); do
    numberOfLines=$(kubectl api-versions | grep -c "installer.kyma-project.io")
    if [[ "$numberOfLines" == "1" ]]; then
      echo "API found"
      break
    elif [[ "${i}" == "${attempts}" ]]; then
      echo "ERROR: API not found, exit"
      exit 1
    fi

    echo "Sleep for 3 seconds"
    sleep 3
  done
}

function installKyma() {
  kymaUnsetVar=false

  # shellcheck disable=SC2043
  for var in GATEWAY_IP_ADDRESS ; do
    if [ -z "${!var}" ] ; then
      echo "ERROR: $var is not set"
      kymaUnsetVar=true
    fi
  done
  if [ "${kymaUnsetVar}" = true ] ; then
    exit 1
  fi

  shout "Build Kyma-Installer Docker image"
  date
  KYMA_INSTALLER_IMAGE="${KYMA_INSTALLER_IMAGE}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/create-image.sh

  KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"
  INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
  INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster-with-compass.yaml.tpl"

  shout "Generate self-signed certificate"
  date
  CERT_KEY=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-self-signed-cert.sh")
  TLS_CERT=$(echo "${CERT_KEY}" | head -1)
  TLS_KEY=$(echo "${CERT_KEY}" | tail -1)

  shout "Apply Kyma config"
  date

  sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" "${INSTALLER_YAML}" \
    | kubectl apply -f-

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "installation-config-overrides" \
    --data "global.domainName=${DOMAIN}" \
    --data "global.loadBalancerIP=${GATEWAY_IP_ADDRESS}"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "application-resource-tests-overrides" \
    --data "application-operator.tests.enabled=false" \
    --data "tests.application_connector_tests.enabled=false" \
    --data "application-registry.tests.enabled=false" \
    --data "console-backend-service.tests.enabled=false" \
    --data "test.acceptance.service-catalog.enabled=false" \
    --data "test.acceptance.external_solution.enabled=false" \
    --data "console.test.acceptance.enabled=false"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "core-test-ui-acceptance-overrides" \
    --data "test.acceptance.ui.logging.enabled=true" \
    --label "component=core"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "cluster-certificate-overrides" \
    --data "global.tlsCrt=${TLS_CERT}" \
    --data "global.tlsKey=${TLS_KEY}"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "monitoring-config-overrides" \
    --data "global.alertTools.credentials.slack.channel=${KYMA_ALERTS_CHANNEL}" \
    --data "global.alertTools.credentials.slack.apiurl=${KYMA_ALERTS_SLACK_API_URL}" \
    --label "component=monitoring"

  "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "istio-overrides" \
    --data "gateways.istio-ingressgateway.loadBalancerIP=${GATEWAY_IP_ADDRESS}" \
    --label "component=istio"

  if [ "${RUN_PROVISIONER_TESTS}" == "true" ]; then
    # Change timeout for kyma test to 2h
    export KYMA_TEST_TIMEOUT=2h

    # Create Config map for Provisioner Tests
    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "provisioner-tests-overrides" \
      --data "global.provisioning.enabled=true" \
      --data "provisioner.security.skipTLSCertificateVeryfication=true" \
      --data "provisioner.tests.enabled=true" \
      --data "provisioner.gardener.kubeconfig=$(base64 -w 0 < "${GARDENER_APPLICATION_CREDENTIALS}")" \
      --data "provisioner.gardener.project=$GARDENER_PROJECT_NAME" \
      --data "provisioner.tests.gardener.azureSecret=$GARDENER_AZURE_SECRET_NAME" \
      --label "component=compass"
  fi

  waitUntilInstallerApiAvailable

  shout "Trigger installation"
  date

  sed -e "s/__VERSION__/0.0.1/g" "${INSTALLER_CR}"  | sed -e "s/__.*__//g" | kubectl apply -f-
  "${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout 30m

  if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
    shout "Create DNS Record for Apiserver proxy IP"
    date
    APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
    IP_ADDRESS=${APISERVER_IP_ADDRESS} DNS_FULL_NAME=${APISERVER_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
  fi
}

trap 'removeCluster ${CLUSTER_NAME}' EXIT INT

shout "Authenticate"
date
init

DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"
export DNS_DOMAIN
DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
export DOMAIN

shout "Create new cluster"
date
createCluster

shout "Install tiller"
date
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
"${KYMA_SCRIPTS_DIR}"/install-tiller.sh

shout "Install Kyma with Compass"
date
installKyma
"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-helm-certs.sh"

shout "Test Kyma with Compass"
date
"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh
