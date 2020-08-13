#!/usr/bin/env bash

# Description: Kyma Upgradeability plan on GKE. The purpose of this script is to install last Kyma release on real GKE cluster, upgrade it with current changes and trigger testing.
#
#
# Expected vars:
#
#  - REPO_OWNER - Set up by prow, repository owner/organization
#  - REPO_NAME - Set up by prow, repository name
#  - BUILD_TYPE - Set up by prow, pr/master/release
#  - DOCKER_PUSH_REPOSITORY - Docker repository hostname
#  - DOCKER_PUSH_DIRECTORY - Docker "top-level" directory (with leading "/")
#  - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
#  - CLOUDSDK_CORE_PROJECT - GCP project for all GCP resources used during execution (Service Account, IP Address, DNS Zone, image registry etc.)
#  - CLOUDSDK_COMPUTE_REGION - GCP compute region
#  - CLOUDSDK_DNS_ZONE_NAME - GCP zone name (not its DNS name!)
#  - GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path
#  - MACHINE_TYPE (optional): GKE machine type
#  - CLUSTER_VERSION (optional): GKE cluster version
#  - KYMA_ARTIFACTS_BUCKET: GCP bucket
#  - BOT_GITHUB_TOKEN: Bot github token used for API queries
#
# Permissions: In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
#  - Compute Admin
#  - Kubernetes Engine Admin
#  - Kubernetes Engine Cluster Admin
#  - DNS Administrator
#  - Service Account User
#  - Storage Admin
#  - Compute Network Admin

set -o errexit

discoverUnsetVar=false

for var in REPO_OWNER REPO_NAME DOCKER_PUSH_REPOSITORY KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS KYMA_ARTIFACTS_BUCKET BOT_GITHUB_TOKEN GCR_PUSH_GOOGLE_APPLICATION_CREDENTIALS; do
    if [[ -z "${!var}" ]] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [[ "${discoverUnsetVar}" = true ]] ; then
    exit 1
fi

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export KYMA_INSTALL_TIMEOUT="30m"
export KYMA_UPDATE_TIMEOUT="40m"
export UPGRADE_TEST_PATH="${KYMA_SOURCES_DIR}/tests/end-to-end/upgrade/chart/upgrade"
export UPGRADE_TEST_NAMESPACE="e2e-upgrade-test"
export UPGRADE_TEST_RELEASE_NAME="${UPGRADE_TEST_NAMESPACE}"
export UPGRADE_TEST_RESOURCE_LABEL="kyma-project.io/upgrade-e2e-test"
export EXTERNAL_SOLUTION_TEST_PATH="${KYMA_SOURCES_DIR}/tests/end-to-end/external-solution-integration/chart/external-solution"
export EXTERNAL_SOLUTION_TEST_NAMESPACE="integration-test"
export EXTERNAL_SOLUTION_TEST_RELEASE_NAME="${EXTERNAL_SOLUTION_TEST_NAMESPACE}"
export EXTERNAL_SOLUTION_TEST_RESOURCE_LABEL="kyma-project.io/external-solution-e2e-test"
export TEST_RESOURCE_LABEL_VALUE_PREPARE="prepareData"
export HELM_TIMEOUT_SEC=10000s # timeout in sec for helm install/test operation
export TEST_TIMEOUT_SEC=600   # timeout in sec for test pods until they reach the terminating state
export TEST_CONTAINER_NAME="tests"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"
# shellcheck disable=SC1090
source "${KYMA_SCRIPTS_DIR}/testing-common.sh"

cleanup() {
    ## Save status of failed script execution
    EXIT_STATUS=$?

    if [[ "${ERROR_LOGGING_GUARD}" = "true" ]]; then
        shout "AN ERROR OCCURED! Take a look at preceding log entries."
        echo
    fi

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    if [[ -n "${CLEANUP_CLUSTER}" ]]; then
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

    if [[ -n "${CLEANUP_GATEWAY_DNS_RECORD}" ]]; then
        shout "Delete Gateway DNS Record"
        date
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/delete-dns-record.sh --project="${CLOUDSDK_CORE_PROJECT}" --zone="${CLOUDSDK_DNS_ZONE_NAME}" --name="${GATEWAY_DNS_FULL_NAME}" --address="${GATEWAY_IP_ADDRESS}" --dryRun=false
    fi

    if [[ -n "${CLEANUP_GATEWAY_IP_ADDRESS}" ]]; then
        shout "Release Gateway IP Address"
        date
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/release-ip-address.sh --project="${CLOUDSDK_CORE_PROJECT}" --ipname="${GATEWAY_IP_ADDRESS_NAME}" --region="${CLOUDSDK_COMPUTE_REGION}" --dryRun=false
    fi

    if [[ -n "${CLEANUP_DOCKER_IMAGE}" ]]; then
        shout "Delete temporary Kyma-Installer Docker image"
        date
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/delete-image.sh"
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

trap cleanup EXIT INT

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    shout "Execute Job Guard"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

function generateAndExportClusterName() {
    readonly REPO_OWNER=$(echo "${REPO_OWNER}" | tr '[:upper:]' '[:lower:]')
    readonly REPO_NAME=$(echo "${REPO_NAME}" | tr '[:upper:]' '[:lower:]')
    readonly RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c10)

    if [[ "$BUILD_TYPE" == "pr" ]]; then
        readonly COMMON_NAME_PREFIX="gke-upgrade-pr"
        # In case of PR, operate on PR number
        COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
    elif [[ "$BUILD_TYPE" == "release" ]]; then
        readonly COMMON_NAME_PREFIX="gke-upgrade-rel"
        readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
        readonly RELEASE_VERSION=$(cat "${SCRIPT_DIR}/../../RELEASE_VERSION")
        shout "Reading release version from RELEASE_VERSION file, got: ${RELEASE_VERSION}"
        COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
    else
        # Otherwise (master), operate on triggering commit id
        readonly COMMON_NAME_PREFIX="gke-upgrade-commit"
        COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)
        COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
    fi

    ### Cluster name must be less than 40 characters!
    export CLUSTER_NAME="${COMMON_NAME}"

    export GCLOUD_NETWORK_NAME="${COMMON_NAME_PREFIX}-net"
    export GCLOUD_SUBNET_NAME="${COMMON_NAME_PREFIX}-subnet"
}

function reserveIPsAndCreateDNSRecords() {
    DNS_SUBDOMAIN="${COMMON_NAME}"
    shout "Authenticate with GCP"
    date
    init

    DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"

    shout "Reserve IP Address for Ingressgateway"
    date
    GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
    GATEWAY_IP_ADDRESS=$(IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/reserve-ip-address.sh")
    CLEANUP_GATEWAY_IP_ADDRESS="true"
    echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

    shout "Create DNS Record for Ingressgateway IP"
    date
    GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
    CLEANUP_GATEWAY_DNS_RECORD="true"
    IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"

    DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
    export DOMAIN
}

function generateAndExportCerts() {
    shout "Generate self-signed certificate"
    date
    CERT_KEY=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-self-signed-cert.sh")

    TLS_CERT=$(echo "${CERT_KEY}" | head -1)
    export TLS_CERT
    TLS_KEY=$(echo "${CERT_KEY}" | tail -1)
    export TLS_KEY
}

function createNetwork() {
    export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
    NETWORK_EXISTS=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/network-exists.sh")
    if [ "$NETWORK_EXISTS" -gt 0 ]; then
        shout "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
        date
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-network-with-subnet.sh"
    else
        shout "Network ${GCLOUD_NETWORK_NAME} exists"
    fi
}

function createCluster() {
    shout "Provision cluster: \"${CLUSTER_NAME}\""
    date
    ### For provision-gke-cluster.sh
    export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
    export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
    export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"
    if [[ -z "${MACHINE_TYPE}" ]]; then
        export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
    fi
    if [[ -z "${CLUSTER_VERSION}" ]]; then
        export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
    fi

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/provision-gke-cluster.sh"
    CLEANUP_CLUSTER="true"
}

function getLastReleaseVersion() {
    version=$(curl --silent --fail --show-error "https://api.github.com/repos/kyma-project/kyma/releases?access_token=${BOT_GITHUB_TOKEN}" \
     | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-1].tag_name')

    echo "${version}"
}

function installKyma() {
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
    mkdir -p /tmp/kyma-gke-upgradeability
    LAST_RELEASE_VERSION=$(getLastReleaseVersion)

    if [ -z "$LAST_RELEASE_VERSION" ]; then
        shoutFail "Couldn't grab latest version from GitHub API, stopping."
        exit 1
    fi

    shout "Apply Kyma config from version ${LAST_RELEASE_VERSION}"
    date
    kubectl create namespace "kyma-installer"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "installation-config-overrides" \
        --data "global.domainName=${DOMAIN}" \
        --data "global.loadBalancerIP=${GATEWAY_IP_ADDRESS}"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "core-test-ui-acceptance-overrides" \
        --data "test.acceptance.ui.logging.enabled=true" \
        --label "component=core"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "application-registry-overrides" \
        --data "application-registry.deployment.args.detailedErrorResponse=true" \
        --label "component=application-connector"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "cluster-certificate-overrides" \
        --data "global.tlsCrt=${TLS_CERT}" \
        --data "global.tlsKey=${TLS_KEY}"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "istio-overrides" \
        --data "gateways.istio-ingressgateway.loadBalancerIP=${GATEWAY_IP_ADDRESS}" \
        --label "component=istio"

    shout "Use released artifacts from version ${LAST_RELEASE_VERSION}"
    date

    curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${LAST_RELEASE_VERSION}/kyma-installer-cluster.yaml" --output /tmp/kyma-gke-upgradeability/last-release-installer.yaml
    kubectl apply -f /tmp/kyma-gke-upgradeability/last-release-installer.yaml

    shout "Installation triggered with timeout ${KYMA_INSTALL_TIMEOUT}"
    date
    "${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout ${KYMA_INSTALL_TIMEOUT}
}

function checkTestPodTerminated() {
    local namespace=$1
    local retry=0
    local runningPods=0
    local succeededPods=0
    local failedPods=0

    while [[ "${retry}" -lt "${TEST_TIMEOUT_SEC}" ]]; do
        # check status phase for each testing pods
        for podName in $(kubectl get pods -n "${namespace}" -o json | jq -sr '.[]|.items[].metadata.name')
        do
            runningPods=$((runningPods + 1))
            phase=$(kubectl get pod "${podName}" -n "${namespace}" -o json | jq '.status.phase')
            echo "Test pod '${podName}' has phase: ${phase}"

            if [[ "${phase}" == *"Succeeded"* ]]
            then
                succeededPods=$((succeededPods + 1))
            fi

            if [[ "${phase}" == *"Failed"* ]]; then
                failedPods=$((failedPods + 1))
            fi
        done

        # exit permanently if one of cluster has failed status
        if [[ "${failedPods}" -gt 0 ]]
        then
            echo "${failedPods} pod(s) has failed status"
            return 1
        fi

        # exit from function if each pod has succeeded status
        if [[ "${runningPods}" == "${succeededPods}" ]]
        then
            echo "All pods in ${namespace} namespace have succeeded phase"
            return 0
        fi

        # reset all counters and rerun checking
        delta=$((runningPods-succeededPods))
        echo "${delta} pod(s) in ${namespace} namespace have not terminated phase. Retry checking."
        runningPods=0
        succeededPods=0
        retry=$((retry + 1))
        sleep 5
    done

    echo "The maximum number of attempts: ${retry} has been reached"
    return 1
}

function installTestChartOrFail() {
  local path=$1
  local name=$2
  local namespace=$3

  shout "Create ${name} resources"
  date

  helm install "${name}" \
      --namespace "${namespace}" \
      --create-namespace \
      "${path}" \
      --timeout "${HELM_TIMEOUT_SEC}" \
      --set domain="${DOMAIN}" \
      --wait

  prepareResult=$?
  if [[ "${prepareResult}" != 0 ]]; then
      echo "Helm install ${name} operation failed: ${prepareResult}"
      exit "${prepareResult}"
  fi
}

function waitForTestPodToFinish() {
  local name=$1
  local namespace=$2
  local label=$3

  set +o errexit
  checkTestPodTerminated "${namespace}"
  prepareTestResult=$?
  set -o errexit

  echo "Logs for prepare data operation to ${name}: "
  # shellcheck disable=SC2046
  kubectl logs -n "${namespace}" $(kubectl get pod -n "${name}" -l "${label}=${TEST_RESOURCE_LABEL_VALUE_PREPARE}" -o json | jq -r '.items | .[] | .metadata.name') -c "${TEST_CONTAINER_NAME}"
  if [[ "${prepareTestResult}" != 0 ]]; then
      echo "Exit status for prepare ${name}: ${prepareTestResult}"
      exit "${prepareTestResult}"
  fi
}

createTestResources() {
    injectTestingAddons

    # install upgrade test
    installTestChartOrFail "${UPGRADE_TEST_PATH}" "${UPGRADE_TEST_RELEASE_NAME}" "${UPGRADE_TEST_NAMESPACE}"

    # install external-solution test
    installTestChartOrFail "${EXTERNAL_SOLUTION_TEST_PATH}" "${EXTERNAL_SOLUTION_TEST_RELEASE_NAME}" "${EXTERNAL_SOLUTION_TEST_NAMESPACE}"

    # wait for upgrade test to finish
    waitForTestPodToFinish "${UPGRADE_TEST_RELEASE_NAME}" "${UPGRADE_TEST_NAMESPACE}" "${UPGRADE_TEST_RESOURCE_LABEL}"

    # wait for external-solution test to finish
    waitForTestPodToFinish "${EXTERNAL_SOLUTION_TEST_RELEASE_NAME}" "${EXTERNAL_SOLUTION_TEST_NAMESPACE}" "${EXTERNAL_SOLUTION_TEST_RESOURCE_LABEL}"
}

function upgradeKyma() {
    shout "Delete the kyma-installation CR and kyma-installer deployment"
    # Remove the finalizer form kyma-installation the merge type is used because strategic is not supported on CRD.
    # More info about merge strategy can be found here: https://tools.ietf.org/html/rfc7386
    kubectl patch Installation kyma-installation -n default --patch '{"metadata":{"finalizers":null}}' --type=merge
    kubectl delete Installation -n default kyma-installation

    # Remove the current installer to prevent it performing any action.
    kubectl delete deployment -n kyma-installer kyma-installer

    if [[ "$BUILD_TYPE" == "release" ]]; then
        echo "Use released artifacts"
        gsutil cp "${KYMA_ARTIFACTS_BUCKET}/${RELEASE_VERSION}/kyma-installer-cluster.yaml" /tmp/kyma-gke-upgradeability/new-release-kyma-installer.yaml

        shout "Update kyma installer"
        kubectl apply -f /tmp/kyma-gke-upgradeability/new-release-kyma-installer.yaml
    else
        shout "Build Kyma Installer Docker image"
        date
        COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)
        KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-upgradeability/${REPO_OWNER}/${REPO_NAME}:COMMIT-${COMMIT_ID}"
        export KYMA_INSTALLER_IMAGE
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-image.sh"
        CLEANUP_DOCKER_IMAGE="true"

        KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"
        INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
        INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"

        shout "Manual concatenating and applying installer.yaml and installer-cr-cluster.yaml YAMLs"
        "${KYMA_SCRIPTS_DIR}"/concat-yamls.sh "${INSTALLER_YAML}" "${INSTALLER_CR}" \
        | sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" \
        | sed -e "s/__VERSION__/0.0.1/g" \
        | sed -e "s/__.*__//g" \
        | kubectl apply -f-
    fi

    shout "Update triggered with timeout ${KYMA_UPDATE_TIMEOUT}"
    date
    "${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout ${KYMA_UPDATE_TIMEOUT}


    if [ -n "$(kubectl get  service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
        shout "Create DNS Record for Apiserver proxy IP"
        date
        APISERVER_IP_ADDRESS=$(kubectl get  service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
        CLEANUP_APISERVER_DNS_RECORD="true"
        IP_ADDRESS=${APISERVER_IP_ADDRESS} DNS_FULL_NAME=${APISERVER_DNS_FULL_NAME} "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-dns-record.sh"
    fi
}

remove_addons_if_necessary() {
  tdWithAddon=$(kubectl get td --all-namespaces -l testing.kyma-project.io/require-testing-addon=true -o custom-columns=NAME:.metadata.name --no-headers=true)

  if [ -z "$tdWithAddon" ]
  then
      echo "- Removing ClusterAddonsConfiguration which provides the testing addons"
      removeTestingAddons
      if [[ $? -eq 1 ]]; then
        exit 1
      fi
  else
      echo "- Skipping removing ClusterAddonsConfiguration"
  fi
}

function testKyma() {
    shout "Test Kyma"
    date
    "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh
}

# Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

generateAndExportClusterName

reserveIPsAndCreateDNSRecords

generateAndExportCerts

createNetwork

createCluster

installKyma

"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-helm-certs.sh"

createTestResources

upgradeKyma

remove_addons_if_necessary

testKyma

shout "Job finished with success"

# Mark execution as successfully
ERROR_LOGGING_GUARD="false"
