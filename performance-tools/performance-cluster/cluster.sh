#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
export PERFORMACE_CLUSTER_SETUP="true"

if [ -f "../../prow/scripts/library.sh" ]; then
    source "../../prow/scripts/library.sh"

elif [ -f "../test-infra/prow/scripts/library.sh" ]; then
    source "../test-infra/prow/scripts/library.sh"

else
    echo "File 'library.sh' can't be found."
    exit 1;
fi

if [ $# -lt "1" ]; then
    echo "Usage:  $0 --action (create or delete) --cluster-grade (production or development)"
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
        --infra)
            checkInputParameterValue "$2"
            INFRA="$2"
            checkInfraInputParameterValue "$2"
            shift # past argument
            shift # past value
        ;;
        --name)
            checkInputParameterValue "$2"
            INPUT_CLUSTER_NAME="$2"
            checkInfraInputParameterValue "$2"
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


export ACTION
export CLUSTER_GRADE
export INFRA
export INPUT_CLUSTER_NAME

if [[ "${ACTION}" == "" ]]; then
    shoutFail "--action is required"
    exit 1
fi

if [[ "${ACTION}" == "create" ]] && [[ "${CLUSTER_GRADE}" == "" ]]; then
    shoutFail "--cluster-grade is required"
    exit 1
fi


shout "Cluster Grade ${CLUSTER_GRADE}"

if [[ "${INPUT_CLUSTER_NAME}" == "" ]]; then
    shoutFail "Environment INPUT_CLUSTER_NAME is required"
    exit 1
fi

if [[ "${INFRA}" != "aks" ]] && [[ ! -f "${GOOGLE_APPLICATION_CREDENTIALS}" ]]; then
    shoutFail "Environment GOOGLE_APPLICATION_CREDENTIALS with service_account credentials is required."
    exit 1
fi

if [[ "${CLUSTER_GRADE}" == "development" ]] && [[ ! -d "${GOPATH}/src/github.com/kyma-project/kyma" ]]; then
    shoutFail "Directory ${GOPATH}/src/github.com/kyma-project/kyma does not exists."
    exit 1
fi


setupCluster() {

    if [[ ${INFRA} == "gke" ]]; then
        setupClusterGKE
    elif [[ ${INFRA} == "aks" ]]; then
        setupClusterAKS
    else
        shoutFail "No cluster infra specified, make sure to either define 'gke' or 'aks'."
        exit 1
    fi

}

setupClusterGKE() {

    set +o errexit
    source "${SCRIPTS_DIR}/scripts/kyma-gke-cluster.sh"
    set -o errexit

}

setupClusterAKS() {

    set +o errexit
    source "${SCRIPTS_DIR}/scripts/kyma-aks-cluster.sh"
    set -o errexit

}

setupCluster

shout "${ACTION} finished with success"
