#!/usr/bin/env bash

set -o errexit

usage () {
    echo "Provide correct GCP bucket name, keyring, encryption key and location!"
    exit 1
}

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
FILES=("sa-gke-kyma-integration" "sa-gke-vm-integration")
EXTENSTION="encrypted"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --bucket)
            BUCKET=$2
            shift # past argument
            shift # past value
        ;;
        --keyring)
            KEYRING=$2
            shift # past argument
            shift # past value
        ;;
        --key)
            KEY=$2
            shift # past argument
            shift # past value
        ;;
        --location)
            LOCATION=$2
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

if [[ -z "${BUCKET}" ]] || [[ -z "${KEYRING}" ]] || [[ -z "${KEY}" ]] || [[ -z "${LOCATION}" ]]; then
    usage
fi

gsutil acl get gs://${BUCKET} > /dev/null

TMP_DIR=`mktemp -d "${CURRENT_DIR}/temp-XXXXXXXXXX"`
trap "rm -rf ${TMP_DIR}" EXIT

for FILE in "${FILES[@]}"
do
    ENCRYPTED_FILE="${FILE}.${EXTENSTION}"
    gsutil cp gs://${BUCKET}/${ENCRYPTED_FILE} ${TMP_DIR}/${FILE}
    gcloud kms decrypt --location "${LOCATION}" --keyring "${KEYRING}" --key "${KEY}" --ciphertext-file "${TMP_DIR}/${FILE}" --plaintext-file "${TMP_DIR}/${FILE}"
    kubectl create secret generic "${FILE}" --from-file="${TMP_DIR}/${FILE}"
done
