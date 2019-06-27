#!/usr/bin/env bash

# Description: Kyma Upgradeability plan on GKE. The purpose of this script is to install last Kyma release on real GKE cluster, upgrade it with current changes and trigger testing.
#
#
# Expected vars:
#
#  - INPUT_CLUSTER_NAME - name for the new cluster
#  - DOCKER_PUSH_REPOSITORY - Docker repository hostname. Ex. "docker.io/anyrepository"
#  - DOCKER_PUSH_DIRECTORY - Docker "top-level" directory (with leading "/")
#   Ex. "/perf"
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

for var in INPUT_CLUSTER_NAME DOCKER_PUSH_REPOSITORY DOCKER_PUSH_DIRECTORY CLOUDSDK_CORE_PROJECT CLOUDSDK_COMPUTE_REGION CLOUDSDK_COMPUTE_ZONE GOOGLE_APPLICATION_CREDENTIALS DOCKER_IN_DOCKER_ENABLED ACTION; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

if [ -f "../../prow/scripts/library.sh" ]; then
    export TEST_INFRA_SOURCES_DIR="../.."

elif [ -f "../test-infra/prow/scripts/library.sh" ]; then
    export TEST_INFRA_SOURCES_DIR="../test-infra"

else
	echo "File 'library.sh' can't be found."
    exit 1;
fi

export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

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

source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

shout "Authenticate"
date
init

if [[ "${ACTION}" == "delete" ]]; then

    shout "Cleanup"
    date
    source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/cleanup-cluster.sh"

elif [[ "${ACTION}" == "create" ]]; then
    shout "Create new cluster"
    date

    if [[ "${CLUSTER_GRADE}" == "" ]]; then
      shoutFail "ERROR: ${CLUSTER_GRADE} is not set"
      exit 0
    fi

    if [[ "${CLUSTER_GRADE}" == "production" ]]; then
        export REPO_OWNER="kyma-project"
        export REPO_NAME="kyma"
        shout "Production"
        mkdir -p /${SRC_DIR}/${REPO_OWNER}/${REPO_NAME}
        git clone https://github.com/${REPO_OWNER}/${REPO_NAME}.git ${SRC_DIR}/${REPO_OWNER}/${REPO_NAME}
        export KYMA_SOURCES_DIR="${SRC_DIR}/${REPO_OWNER}/${REPO_NAME}"
    else
        for var in REPO_OWNER REPO_NAME; do
            if [ -z "${!var}" ] ; then
                echo "ERROR: $var is not set"
                discoverUnsetVar=true
            fi
        done
        export KYMA_SOURCES_DIR="${GOPATH}/src/github.com/kyma-project/kyma"
    fi

    export KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"

    source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-cluster.sh"
    shout "Install tiller"
    date

    shout "Account is: $(gcloud config get-value account)"

    gcloud container clusters get-credentials $INPUT_CLUSTER_NAME --zone $CLOUDSDK_COMPUTE_ZONE --project $CLOUDSDK_CORE_PROJECT
    
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value account)"
    "${KYMA_SCRIPTS_DIR}"/install-tiller.sh

    shout "Install kyma"
    date

    source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/install-kyma.sh"

    source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-helm-certs.sh"
else
   shoutFail "None of the actions met"
fi

shout "Success"