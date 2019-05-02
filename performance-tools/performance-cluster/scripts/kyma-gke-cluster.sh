#!/usr/bin/env bash

# Description: Kyma Upgradeability plan on GKE. The purpose of this script is to install last Kyma release on real GKE cluster, upgrade it with current changes and trigger testing.
#
#
# Expected vars:
#
#  - INPUT_CLUSTER_NAME - name for the new cluster
#  - DOCKER_PUSH_REPOSITORY - Docker repository hostname. Ex. ""
#  - DOCKER_PUSH_DIRECTORY - Docker "top-level" directory (with leading "/")
#  - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation.
#    Ex. "/home/${USER}/go/src/github.com/kyma-project"
#
#  - CLOUDSDK_CORE_PROJECT - GCP project for all GCP resources used during execution (Service Account, IP Address, DNS Zone, image registry etc.)
#  - CLOUDSDK_COMPUTE_REGION - GCP compute region. Ex. "europe-west3"
#  - CLOUDSDK_COMPUTE_ZONE Ex. "europe-west3-a"
#  - GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path.
#    Ex. "/etc/credentials/sa-gke-kyma-integration/service-account.json"
#
#  - DOCKER_IN_DOCKER_ENABLED true
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
set -o pipefail  # Fail a pipe if any sub-command fails.
discoverUnsetVar=false

for var in INPUT_CLUSTER_NAME DOCKER_PUSH_REPOSITORY DOCKER_PUSH_DIRECTORY KYMA_PROJECT_DIR CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_COMPUTE_ZONE GOOGLE_APPLICATION_CREDENTIALS DOCKER_IN_DOCKER_ENABLED CLUSTER_GRADE ACTION REPO_OWNER REPO_NAME; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi



export TEST_INFRA_PERFORMANCE_TOOLS_CLUSTER_SCRIPTS="${CURRENT_PATH}/scripts/helpers"

export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"
export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
export BUILD_TYPE="master"
readonly CURRENT_TIMESTAMP=$(date +%Y%m%d)
readonly STANDARIZED_NAME=$(echo "${INPUT_CLUSTER_NAME}" | tr "[:upper:]" "[:lower:]")
readonly DNS_SUBDOMAIN="${STANDARIZED_NAME}"

export CLUSTER_NAME="${STANDARIZED_NAME}"

export STANDARIZED_NAME
export REPO_OWNER
export REPO_NAME
export CURRENT_TIMESTAMP

source "${CURRENT_PATH}/scripts/library.sh"

shout "Authenticate"
date
init

date
authenticateDocker

if [[ "${ACTION}" == "delete" ]]; then

    shout "Cleanup"
    date
    source "${TEST_INFRA_PERFORMANCE_TOOLS_CLUSTER_SCRIPTS}/cleanup-cluster.sh"

elif [[ "${ACTION}" == "create" ]]; then
    shout "Create new cluster"
    date

    export SRC_DIR="$(mktemp -d -t src.XXXXXX)"
    ls "/tmp/"

    if [[ "${CLUSTER_GRADE}" == "production" ]]; then
        shout "Production"
        # git clone -b <branch> <remote_repo>
        mkdir -p /${SRC_DIR}/${REPO_OWNER}/${REPO_NAME}
        git clone https://github.com/${REPO_OWNER}/${REPO_NAME}.git ${SRC_DIR}/${REPO_OWNER}/${REPO_NAME}
        #git clone -b ${BRANCH_NAME} --single-branch https://github.com/${REPO_OWNER}/${REPO_NAME}.git ${SRC_DIR}/${REPO_OWNER}/${REPO_NAME}
        export KYMA_SOURCES_DIR="${SRC_DIR}/${REPO_OWNER}/${REPO_NAME}"
    else

        export KYMA_SOURCES_DIR="${GOPATH}/src/github.com/kyma-project/kyma"
    fi

export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"

    source "${TEST_INFRA_PERFORMANCE_TOOLS_CLUSTER_SCRIPTS}/create-cluster.sh"
    shout "Install tiller"
    date

    shout "Account is: $(gcloud config get-value account)"

    gcloud container clusters get-credentials $INPUT_CLUSTER_NAME --zone $CLOUDSDK_COMPUTE_ZONE --project $CLOUDSDK_CORE_PROJECT
    
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
    "${KYMA_SCRIPTS_DIR}"/install-tiller.sh

    shout "Install kyma"
    date

    source "${TEST_INFRA_PERFORMANCE_TOOLS_CLUSTER_SCRIPTS}/install-kyma.sh"

    source "${TEST_INFRA_PERFORMANCE_TOOLS_CLUSTER_SCRIPTS}/get-helm-certs.sh"
else
   shoutFail "None of the actions met"
fi

shout "Success"