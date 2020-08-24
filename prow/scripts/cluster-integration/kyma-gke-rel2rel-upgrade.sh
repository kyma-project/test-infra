#!/usr/bin/env bash

# Description: Kyma release-to-release Upgradability plan on GKE.
# The purpose of this script is to install the previous Kyma release on real GKE cluster, upgrade it to the current release and trigger testing.
#
# Expected vars:
#  - REPO_OWNER - Set up by prow, repository owner/organization
#  - REPO_NAME - Set up by prow, repository name
#  - PULL_BASE_REF: Set up by prow, tag that triggered the build
#  - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
#  - CLOUDSDK_CORE_PROJECT - GCP project for all GCP resources used during execution (Service Account, IP Address, DNS Zone, image registry etc.)
#  - CLOUDSDK_COMPUTE_REGION - GCP compute region
#  - CLOUDSDK_DNS_ZONE_NAME - GCP zone name (not its DNS name!)
#  - CLOUDSDK_COMPUTE_ZONE - GCP compute zone
#  - GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path
#  - BOT_GITHUB_TOKEN: Bot github token used for API queries
#  - MACHINE_TYPE (optional): GKE machine type
#  - CLUSTER_VERSION (optional): GKE cluster version
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

for var in REPO_OWNER REPO_NAME KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS BOT_GITHUB_TOKEN CLOUDSDK_COMPUTE_ZONE; do
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
export KYMA_UPDATE_TIMEOUT="25m"
export UPGRADE_TEST_PATH="${KYMA_SOURCES_DIR}/tests/end-to-end/upgrade/chart/upgrade"
export UPGRADE_TEST_HELM_TIMEOUT_SEC=10000s # timeout in sec for helm operation install/test
export UPGRADE_TEST_TIMEOUT_SEC=600 # timeout in sec for e2e upgrade test pods until they reach the terminating state
export UPGRADE_TEST_NAMESPACE="e2e-upgrade-test"
export UPGRADE_TEST_RELEASE_NAME="${UPGRADE_TEST_NAMESPACE}"
export UPGRADE_TEST_RESOURCE_LABEL="kyma-project.io/upgrade-e2e-test"
export UPGRADE_TEST_LABEL_VALUE_PREPARE="prepareData"
export UPGRADE_TEST_LABEL_VALUE_EXECUTE="executeTests"
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

getSourceVersion() {
    releaseIndex=2
    if [[ "${PULL_BASE_REF}" == *"-rc"* ]] ; then
        releaseIndex=1
    fi

    # shellcheck disable=SC2016
    version=$(curl --silent --fail --show-error "https://api.github.com/repos/kyma-project/kyma/releases?access_token=${BOT_GITHUB_TOKEN}" \
     | jq -r --argjson index "${releaseIndex}" 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-$index].tag_name')

    echo "${version}"
}

downloadAssets() {
    mkdir -p /tmp/kyma-gke-upgradeability

    SOURCE_VERSION=$(getSourceVersion)
    TARGET_VERSION="${PULL_BASE_REF}"

    shout "Upgrade from ${SOURCE_VERSION} to ${TARGET_VERSION}"
    date

    if [[ -z "$SOURCE_VERSION" ]]; then
        shoutFail "Couldn't grab latest version from GitHub API, stopping."
        exit 1
    fi

    if [[ "$SOURCE_VERSION" == "1.14.0" ]]; then
        curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${SOURCE_VERSION}/kyma-installer-cluster.yaml" \
            --output /tmp/kyma-gke-upgradeability/original-release-installer.yaml

        curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${TARGET_VERSION}/kyma-installer.yaml" --output /tmp/kyma-gke-upgradeability/upgraded-kyma-installer.yaml
        curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${TARGET_VERSION}/kyma-installer-cr-cluster.yaml" --output /tmp/kyma-gke-upgradeability/upgraded-kyma-installer-cr-cluster.yaml
    else
        curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${SOURCE_VERSION}/kyma-installer.yaml" --output /tmp/kyma-gke-upgradeability/original-kyma-installer.yaml
        curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${SOURCE_VERSION}/kyma-installer-cr-cluster.yaml" --output /tmp/kyma-gke-upgradeability/original-kyma-installer-cr-cluster.yaml

        curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${TARGET_VERSION}/kyma-installer.yaml" --output /tmp/kyma-gke-upgradeability/upgraded-kyma-installer.yaml
        curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${TARGET_VERSION}/kyma-installer-cr-cluster.yaml" --output /tmp/kyma-gke-upgradeability/upgraded-kyma-installer-cr-cluster.yaml
    fi
}

generateAndExportClusterName() {
    readonly REPO_OWNER=$(echo "${REPO_OWNER}" | tr '[:upper:]' '[:lower:]')
    readonly REPO_NAME=$(echo "${REPO_NAME}" | tr '[:upper:]' '[:lower:]')
    readonly RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c5)
    readonly COMMON_NAME_PREFIX="gke-rel-upgrade"

    local versionFrom
    versionFrom=$(echo "${SOURCE_VERSION}" | tr -d ".-")

    local versionTo
    versionTo=$(echo "${TARGET_VERSION}" | tr -d ".-")

    COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${versionFrom}-${versionTo}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")

    ### Cluster name must be less than 40 characters!
    export CLUSTER_NAME="${COMMON_NAME}"
    export GCLOUD_NETWORK_NAME="${COMMON_NAME_PREFIX}-net"
    export GCLOUD_SUBNET_NAME="${COMMON_NAME_PREFIX}-subnet"
}

reserveIPsAndCreateDNSRecords() {
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

generateAndExportCerts() {
    shout "Generate self-signed certificate"
    date
    CERT_KEY=$("${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/generate-self-signed-cert.sh")

    TLS_CERT=$(echo "${CERT_KEY}" | head -1)
    export TLS_CERT
    TLS_KEY=$(echo "${CERT_KEY}" | tail -1)
    export TLS_KEY
}

createNetwork() {
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

createCluster() {
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

installKyma() {
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"

    shout "Apply Kyma config"
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

    if [[ "$SOURCE_VERSION" == "1.14.0" ]]; then
        shout "Use release artifacts from version ${SOURCE_VERSION}"
        date
        kubectl apply -f /tmp/kyma-gke-upgradeability/original-release-installer.yaml
    else
        shout "Use release artifacts from version ${SOURCE_VERSION}"
        date
        kubectl apply -f /tmp/kyma-gke-upgradeability/original-kyma-installer.yaml
        kubectl apply -f /tmp/kyma-gke-upgradeability/original-kyma-installer-cr-cluster.yaml
    fi

    shout "Installation triggered with timeout ${KYMA_INSTALL_TIMEOUT}"
    date
    "${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout ${KYMA_INSTALL_TIMEOUT}
}

checkTestPodTerminated() {
    local retry=0
    local runningPods=0
    local succeededPods=0
    local failedPods=0

    while [ "${retry}" -lt "${UPGRADE_TEST_TIMEOUT_SEC}" ]; do
        # check status phase for each testing pods
        for podName in $(kubectl get pods -n "${UPGRADE_TEST_NAMESPACE}" -o json | jq -sr '.[]|.items[].metadata.name')
        do
            runningPods=$((runningPods + 1))
            phase=$(kubectl get pod "${podName}" -n "${UPGRADE_TEST_NAMESPACE}" -o json | jq '.status.phase')
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
        if [ "${failedPods}" -gt 0 ]
        then
            echo "${failedPods} pod(s) has failed status"
            return 1
        fi

        # exit from function if each pod has succeeded status
        if [ "${runningPods}" == "${succeededPods}" ]
        then
            echo "All pods in ${UPGRADE_TEST_NAMESPACE} namespace have succeeded phase"
            return 0
        fi

        # reset all counters and rerun checking
        delta=$((runningPods-succeededPods))
        echo "${delta} pod(s) in ${UPGRADE_TEST_NAMESPACE} namespace have not terminated phase. Retry checking."
        runningPods=0
        succeededPods=0
        retry=$((retry + 1))
        sleep 5
    done

    echo "The maximum number of attempts: ${retry} has been reached"
    return 1
}

createTestResources() {
    shout "Create e2e upgrade test resources"
    date

    injectTestingAddons
 
    helm install "${UPGRADE_TEST_RELEASE_NAME}" \
        --namespace "${UPGRADE_TEST_NAMESPACE}" \
        --create-namespace \
        "${UPGRADE_TEST_PATH}" \
        --timeout "${UPGRADE_TEST_HELM_TIMEOUT_SEC}" \
        --wait ${HELM_ARGS}

    prepareResult=$?
    if [ "${prepareResult}" != 0 ]; then
        echo "Helm install operation failed: ${prepareResult}"
        exit "${prepareResult}"
    fi

    set +o errexit
    checkTestPodTerminated
    prepareTestResult=$?
    set -o errexit

    echo "Logs for prepare data operation to test e2e upgrade: "
    kubectl logs -n "${UPGRADE_TEST_NAMESPACE}" -l "${UPGRADE_TEST_RESOURCE_LABEL}=${UPGRADE_TEST_LABEL_VALUE_PREPARE}" -c "${TEST_CONTAINER_NAME}"
    if [ "${prepareTestResult}" != 0 ]; then
        echo "Exit status for prepare upgrade e2e tests: ${prepareTestResult}"
        exit "${prepareTestResult}"
    fi
}

upgradeKyma() {
    shout "Delete the kyma-installation CR and kyma-installer deployment"
    # Remove the finalizer form kyma-installation the merge type is used because strategic is not supported on CRD.
    # More info about merge strategy can be found here: https://tools.ietf.org/html/rfc7386
    kubectl patch Installation kyma-installation -n default --patch '{"metadata":{"finalizers":null}}' --type=merge
    kubectl delete Installation -n default kyma-installation

    # Remove the current installer to prevent it performing any action.
    kubectl delete deployment -n kyma-installer kyma-installer

    
    shout "Use release artifacts from version ${TARGET_VERSION}"
    date
    kubectl apply -f /tmp/kyma-gke-upgradeability/upgraded-kyma-installer.yaml
    kubectl apply -f /tmp/kyma-gke-upgradeability/upgraded-kyma-installer-cr-cluster.yaml

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

testKyma() {
    shout "Test Kyma end-to-end upgrade scenarios"
    date

    set +o errexit
    # shellcheck disable=SC2086
    helm test -n "${UPGRADE_TEST_NAMESPACE}" "${UPGRADE_TEST_RELEASE_NAME}" --timeout "${UPGRADE_TEST_HELM_TIMEOUT_SEC}" ${HELM_ARGS}
    testEndToEndResult=$?

    echo "Test e2e upgrade logs: "
    kubectl logs -n "${UPGRADE_TEST_NAMESPACE}" -l "${UPGRADE_TEST_RESOURCE_LABEL}=${UPGRADE_TEST_LABEL_VALUE_EXECUTE}" -c "${TEST_CONTAINER_NAME}"

    if [ "${testEndToEndResult}" != 0 ]; then
        echo "Helm test operation failed: ${testEndToEndResult}"
        exit "${testEndToEndResult}"
    fi
    set -o errexit

    shout "Test Kyma"
    date
    "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh
}

# Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

downloadAssets

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
