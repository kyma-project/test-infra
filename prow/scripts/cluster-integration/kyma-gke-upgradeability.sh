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
#  - BOT_GITHUB_TOKEN
#
# Permissions: In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
#  - Compute Admin
#  - Kubernetes Engine Admin
#  - Kubernetes Engine Cluster Admin
#  - DNS Administrator
#  - Service Account User
#  - Storage Admin

set -o errexit

discoverUnsetVar=false

for var in REPO_OWNER REPO_NAME DOCKER_PUSH_REPOSITORY KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_DNS_ZONE_NAME GOOGLE_APPLICATION_CREDENTIALS KYMA_ARTIFACTS_BUCKET; do
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
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

trap cleanup EXIT

#!Put cleanup code in this function!
cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?

    if [[ "${ERROR_LOGGING_GUARD}" = "true" ]]; then
        shout "AN ERROR OCCURED! Take a look at preceding log entries."
        echo
    fi

    kubectl get pod --all-namespaces
    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    if [[ -n "${CLEANUP_CLUSTER}" ]]; then
        shout "Deprovision cluster: \"${CLUSTER_NAME}\""
        date

        #save disk names while the cluster still exists to remove them later
        DISKS=$(kubectl get pvc --all-namespaces -o jsonpath="{.items[*].spec.volumeName}" | xargs -n1 echo)
        export DISKS

        #Delete cluster
        "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/deprovision-gke-cluster.sh
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi

        #Delete orphaned disks
        shout "Delete orphaned PVC disks..."
        date
        "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/delete-disks.sh
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    fi

    if [[ -n "${CLEANUP_GATEWAY_DNS_RECORD}" ]]; then
        shout "Delete Gateway DNS Record"
        date
        IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/delete-dns-record.sh
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    fi

    if [[ -n "${CLEANUP_GATEWAY_IP_ADDRESS}" ]]; then
        shout "Release Gateway IP Address"
        date
        IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/release-ip-address.sh
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    fi

    if [[ -n "${CLEANUP_REMOTEENVS_DNS_RECORD}" ]]; then
        shout "Delete Remote Environments DNS Record"
        date
        IP_ADDRESS=${REMOTEENVS_IP_ADDRESS} DNS_FULL_NAME=${REMOTEENVS_DNS_FULL_NAME} "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/delete-dns-record.sh
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    fi

    if [[ -n "${CLEANUP_REMOTEENVS_IP_ADDRESS}" ]]; then
        shout "Release Remote Environments IP Address"
        date
        IP_ADDRESS_NAME=${REMOTEENVS_IP_ADDRESS_NAME} "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/release-ip-address.sh
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    fi

    if [[ -n "${CLEANUP_DOCKER_IMAGE}" ]]; then
        shout "Delete temporary Kyma-Installer Docker image"
        date
        "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/delete-image.sh
        TMP_STATUS=$?
        if [[ ${TMP_STATUS} -ne 0 ]]; then EXIT_STATUS=${TMP_STATUS}; fi
    fi


    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    shout "Job is finished ${MSG}"
    date
    set -e

    exit "${EXIT_STATUS}"
}

# Enforce lowercase
readonly REPO_OWNER=$(echo "${REPO_OWNER}" | tr '[:upper:]' '[:lower:]')
export REPO_OWNER
readonly REPO_NAME=$(echo "${REPO_NAME}" | tr '[:upper:]' '[:lower:]')
export REPO_NAME

function generateAndExportClusterName() {
    RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c10)

    if [[ "$BUILD_TYPE" == "pr" ]]; then
        # In case of PR, operate on PR number
        COMMON_NAME=$(echo "gke-upgrade-pr-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
        KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-upgradeability/${REPO_OWNER}/${REPO_NAME}:PR-${PULL_NUMBER}"
        export KYMA_INSTALLER_IMAGE
    elif [[ "$BUILD_TYPE" == "release" ]]; then
        readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
        readonly RELEASE_VERSION=$(cat "${SCRIPT_DIR}/../../RELEASE_VERSION")
        shout "Reading release version from RELEASE_VERSION file, got: ${RELEASE_VERSION}"
        COMMON_NAME=$(echo "gke-upgrade-rel-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
    else
        # Otherwise (master), operate on triggering commit id
        readonly COMMIT_ID=$(cd "$KYMA_SOURCES_DIR" && git rev-parse --short HEAD)
        COMMON_NAME=$(echo "gke-upgrade-commit-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
        KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/gke-upgradeability/${REPO_OWNER}/${REPO_NAME}:COMMIT-${COMMIT_ID}"
        export KYMA_INSTALLER_IMAGE
    fi

    ### Cluster name must be less than 40 characters!
    export CLUSTER_NAME="${COMMON_NAME}"
}

generateAndExportClusterName

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

DNS_SUBDOMAIN="${COMMON_NAME}"
shout "Authenticate"
date
init
DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"

shout "Reserve IP Address for Ingressgateway"
date
GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
GATEWAY_IP_ADDRESS=$(IP_ADDRESS_NAME=${GATEWAY_IP_ADDRESS_NAME} "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/reserve-ip-address.sh)
CLEANUP_GATEWAY_IP_ADDRESS="true"
echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

shout "Create DNS Record for Ingressgateway IP"
date
GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
CLEANUP_GATEWAY_DNS_RECORD="true"
IP_ADDRESS=${GATEWAY_IP_ADDRESS} DNS_FULL_NAME=${GATEWAY_DNS_FULL_NAME} "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/create-dns-record.sh

shout "Reserve IP Address for Remote Environments"
date
REMOTEENVS_IP_ADDRESS_NAME="remoteenvs-${COMMON_NAME}"
REMOTEENVS_IP_ADDRESS=$(IP_ADDRESS_NAME=${REMOTEENVS_IP_ADDRESS_NAME} "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/reserve-ip-address.sh)
CLEANUP_REMOTEENVS_IP_ADDRESS="true"
echo "Created IP Address for Remote Environments: ${REMOTEENVS_IP_ADDRESS}"

shout "Create DNS Record for Remote Environments IP"
date
REMOTEENVS_DNS_FULL_NAME="gateway.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
CLEANUP_REMOTEENVS_DNS_RECORD="true"
IP_ADDRESS=${REMOTEENVS_IP_ADDRESS} DNS_FULL_NAME=${REMOTEENVS_DNS_FULL_NAME} "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/create-dns-record.sh

shout "Generate self-signed certificate"
date
DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
export DOMAIN
CERT_KEY=$("${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/generate-self-signed-cert.sh)
TLS_CERT=$(echo "${CERT_KEY}" | head -1)
TLS_KEY=$(echo "${CERT_KEY}" | tail -1)


shout "Provision cluster: \"${CLUSTER_NAME}\""
date
### For provision-gke-cluster.sh
export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"
if [ -z "$MACHINE_TYPE" ]; then
      export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
fi
if [ -z "${CLUSTER_VERSION}" ]; then
      export CLUSTER_VERSION="${DEFAULT_CLUSTER_VERSION}"
fi
"${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/provision-gke-cluster.sh
CLEANUP_CLUSTER="true"


#Local variables
KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"
KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"
INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"

shout "Install Tiller"
date
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
"${KYMA_SCRIPTS_DIR}"/install-tiller.sh

shout "Apply Kyma config from latest release - pre releases are omitted"
date

mkdir -p /tmp/kyma-gke-upgradeability

function getLastReleaseVersion() {
    version=$(curl --silent --fail --show-error "https://api.github.com/repos/kyma-project/kyma/releases?access_token=${BOT_GITHUB_TOKEN}" \
     | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-1].tag_name')

    echo ${version}
}

LAST_RELEASE_VERSION=$(getLastReleaseVersion)
echo "Use released artifacts from version ${LAST_RELEASE_VERSION}"
curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${LAST_RELEASE_VERSION}/kyma-config-cluster.yaml" --output /tmp/kyma-gke-upgradeability/last-release-config.yaml
sed -e "s/__DOMAIN__/${DOMAIN}/g" /tmp/kyma-gke-upgradeability/last-release-config.yaml \
    | sed -e "s/__REMOTE_ENV_IP__/${REMOTEENVS_IP_ADDRESS}/g" \
    | sed -e "s/__TLS_CERT__/${TLS_CERT}/g" \
    | sed -e "s/__TLS_KEY__/${TLS_KEY}/g" \
    | sed -e "s/__EXTERNAL_PUBLIC_IP__/${GATEWAY_IP_ADDRESS}/g" \
    | sed -e "s/__SKIP_SSL_VERIFY__/true/g" \
    | sed -e "s/__.*__//g" \
    | kubectl apply -f-

shout "Trigger installation"
date
kubectl label installation/kyma-installation action=install
"${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout 30m

echo "Delete the kyma-installation CR"
# Remove the finalizer form kyma-installation the merge type is used because strategic is not supported on CRD.
# More info about merge strategy can be found here: https://tools.ietf.org/html/rfc7386
kubectl patch Installation kyma-installation -n default --patch '{"metadata":{"finalizers":null}}' --type=merge
kubectl delete Installation -n default kyma-installation

if [[ "$BUILD_TYPE" == "release" ]]; then
    echo "Use released artifacts"
    gsutil cp "${KYMA_ARTIFACTS_BUCKET}/${RELEASE_VERSION}/kyma-config-upgrade-cluster.yaml" /tmp/kyma-gke-upgradeability/new-release-upgrade-cluster.yaml

    kubectl apply -f /tmp/kyma-gke-integration/new-release-upgrade-cluster.yaml
else
    shout "Build Kyma Installer Docker image"
    # use the KYMA_INSTALLER_IMAGE defined on top
    date
    "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/cluster-integration/create-image.sh
    CLEANUP_DOCKER_IMAGE="true"

    shout "Manual concatenating and applying installer.yaml and installer-cr-cluster.yaml YAMLs"
    "${KYMA_SCRIPTS_DIR}"/concat-yamls.sh "${INSTALLER_YAML}" "${INSTALLER_CR}" \
    | sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" \
    | sed -e "s/__VERSION__/0.0.1/g" \
    | sed -e "s/__.*__//g" \
    | kubectl apply -f-

    kubectl get deploy/kyma-installer -n kyma-installer -o yaml
fi

shout "Trigger update"
date
kubectl label installation/kyma-installation action=install
"${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout 10m

shout "Test Kyma"
date
"${KYMA_SCRIPTS_DIR}"/testing.sh

shout "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
