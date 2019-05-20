#!/usr/bin/env bash

NAME=""
DATA=""
PARSED_DATA=""

LABELS="kyma-project.io/installation= installer=overrides "
NAMESPACE="kyma-installer"
VALUE_REGEXP="^[a-z0-9.-]+=[a-z0-9.-]+$"

function checkScriptInput {
    if [[ -z "${1}" ]] || [[ "${1:0:2}" == "--" ]]; then
        echo "Missing parameter value"
        exit 1
    fi
}

function checkValue {
    [[ "$1" =~ ${VALUE_REGEXP} ]] || { echo "error: incorrect data/label value. Exiting..." && exit 1; }
}

function usage {
    echo "Incorrect input. Available flags: --name, --data, --label"
    exit 1
}

function checkName {
    if [[ -z "${NAME}" ]]; then
	    echo "error: configmap name has not been specified"
	    exit 1
    fi
}

function checkIfExists {
    kubectl get configmap -n "${NAMESPACE}" "${NAME}" >> /dev/null 2>&1 \
        && echo "error: configmap ${NAME} already exists in the ${NAMESPACE} namespace. Exiting..."  \
        && exit 1
}

function parseData {
    IFS=' ' read -r -a tokens <<< "${DATA}"
    for token in "${tokens[@]}"; do
        PARSED_DATA+=" --from-literal=${token}"
    done
}

function createConfigmap {
    kubectl create configmap -n ${NAMESPACE} ${NAME} ${PARSED_DATA}
    kubectl label configmap -n ${NAMESPACE} ${NAME} ${LABELS}
}

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"
    case ${key} in
        --name)
            checkScriptInput "$2"
            NAME="$2"
            shift # past argument
            shift # past value
            ;;
        --data)
            checkScriptInput "$2"
            checkValue "$2"
            DATA+="$2 "
            shift # past argument
            shift # past value
            ;;
        --label)
            checkScriptInput "$2"
            checkValue "$2"
            LABELS+="$2 "
            shift # past argument
            shift # past value
            ;;
        *)    # unknown option
            usage
            ;;
    esac
done

checkName
checkIfExists
parseData
createConfigmap
