#!/usr/bin/env bash

NAME=""
DATA=()
LABELS=("kyma-project.io/installation=" "installer=overrides")

NAMESPACE="kyma-installer"
LABEL_REGEXP="^[a-zA-Z0-9./-]+=[a-zA-Z0-9./-]*$"


function checkScriptInput {
    if [[ -z "${1}" ]] || [[ "${1:0:2}" == "--" ]]; then
        echo "Missing parameter value"
        exit 1
    fi
}

function checkLabel {
    [[ "$1" =~ ${LABEL_REGEXP} ]] || { echo "error: malformed label or label value. Exiting..." && exit 1; }
}

function usage {
    echo "Incorrect input. Available flags: --name, --namespace, --data, --file, --label"
    exit 1
}

function checkName {
    if [[ -z "${NAME}" ]]; then
	    echo "error: secret name has not been specified"
	    exit 1
    fi
}

function checkIfExists {
    kubectl get secret -n "${NAMESPACE}" "${NAME}" >> /dev/null 2>&1 \
        && echo "error: secret ${NAME} already exists in the ${NAMESPACE} namespace. Exiting..."  \
        && exit 1
}

function createSecret {
    kubectl create secret generic -n ${NAMESPACE} ${NAME} "${DATA[@]}"
    kubectl label secret -n ${NAMESPACE} ${NAME} "${LABELS[@]}"
}

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
        --namespace)
            checkScriptInput "$2"
            NAMESPACE="$2"
            shift # past argument
            shift # past value
            ;;
        --data)
            checkScriptInput "$2"
            DATA+=("--from-literal=$2")
            shift # past argument
            shift # past value
            ;;
        --file)
            checkScriptInput "$2"
            DATA+=("--from-file=$2")
            shift # past argument
            shift # past value
            ;;
        --label)
            checkScriptInput "$2"
            checkLabel "$2"
            LABELS+=("$2")
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
createSecret
