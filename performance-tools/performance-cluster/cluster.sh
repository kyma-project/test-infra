#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any

export INPUT_CLUSTER_NAME="production"
export DOCKER_PUSH_REPOSITORY="eu.gcr.io/kyma-project"
export DOCKER_PUSH_DIRECTORY="/develop"
export KYMA_PROJECT_DIR="kyma-project"
export CLOUDSDK_CORE_PROJECT="sap-kyma-prow"
export CLOUDSDK_COMPUTE_REGION="europe-west3"
export CLOUDSDK_COMPUTE_ZONE="europe-west3-a"
export CLOUDSDK_DNS_ZONE_NAME="build-kyma" #GCloud DNS Zone Name (NOT it's DNS name!)
export GOOGLE_APPLICATION_CREDENTIALS=/etc/credentials/sa-gke-kyma-integration/service-account.json
export DOCKER_IN_DOCKER_ENABLED="true"
export REPO_OWNER="kyma-project"
export REPO_NAME="kyma"
export SRC_DIR="src"
export CURRENT_PATH="$PWD"
export ACTION=""
export CLUSTER_GRADE=""
export CLOUDSDK_PROJECT="kyma-project"

# shellcheck disable=SC1090
source "${CURRENT_PATH}/scripts/library.sh"

if [ $# -lt "1" ]; then
        echo "Usage:  $0 --action (create or delete) --cluster-grade (production or development) -repo-owner (if development) --repo-name (if development)"
        exit 1;
fi

POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"

    case ${key} in
        --action)
            checkInputParameterValue "$2"
            ACTION="${2}"
            checkActionInputParameterValue "$2"
            shift # past argument
            shift # past value
        ;;
        --cluster-grade)
            checkInputParameterValue "$2"
            CLUSTER_GRADE="$2"
            checkClusterGradeInputParameterValue "$2"
            shift # past argument
            shift # past value
        ;;
        --repo-owner)
            checkInputParameterValue "$2"
            OWNER="$2"
            shift
            shift
        ;;
        --repo-name)
            checkInputParameterValue "$2"
            NAME="${2}"
            shift # past argument
            shift # past value
        ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
        ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters


if [[ "${CLUSTER_GRADE}" == "" ]] || [[ "${ACTION}" == "" ]]; then
    shoutFail "--action and --cluster-grade required"
    exit 0
fi

shout "Cluster Grade ${CLUSTER_GRADE}"

if [[ "${CLUSTER_GRADE}" == "development" ]]; then
    export INPUT_CLUSTER_NAME="${CLUSTER_GRADE}"
fi

if [[ "${CLUSTER_GRADE}" == "development" ]] && [[ "${OWNER}" == "" ]] && [[ "${NAME}" == "" ]]; then
    shoutFail "--repo-name and --repo-owner for cluster grade development"
    exit 0
fi

setupCluster() {
    gcloud container clusters list --project $CLOUDSDK_PROJECT
    # atx-prow2          europe-west3-a  1.11.8-gke.6    35.198.132.189  n1-standard-1  1.11.7-gke.6 *   3          RUNNING
    # gcloud container clusters get-credentials atx-prow2 --zone $CLOUDSDK_COMPUTE_ZONE --project $CLOUDSDK_PROJECT

    set +o errexit
    # shellcheck disable=SC1090
    source "scripts/kyma-gke-cluster.sh"
    set -o errexit

}

setupCluster

shout "${ACTION} finished with success"