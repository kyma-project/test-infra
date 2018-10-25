#!/usr/bin/env bash

set -o errexit

usage () {
    echo "Provide correct GCP bucket name, keyring and encryption key!"
    exit 1
}

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SECRET_NAME="gcp-prow"
FILES=("secret")
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
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
        ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

if [[ -z "${BUCKET}" ]] || [[ -z "${KEYRING}" ]] || [[ -z "${KEY}" ]]; then
    usage
fi

##########

gsutil acl get gs://${BUCKET} > /dev/null

##########

TMP_DIR=`mktemp -d "${CURRENT_DIR}/temp-XXXXXXXXXX"`
trap "rm -rf ${TMP_DIR}" EXIT

for FILE in "${FILES[@]}"
do
    ENCRYPTED_FILE="${FILE}.${EXTENSTION}"
    gsutil cp gs://${BUCKET}/${ENCRYPTED_FILE} ${TMP_DIR}/${ENCRYPTED_FILE}
    gcloud kms decrypt --location "global" --keyring "${KEYRING}" --key "${KEY}" --ciphertext-file "${TMP_DIR}/${ENCRYPTED_FILE}" --plaintext-file "${TMP_DIR}/${FILE}"
    rm ${TMP_DIR}/${ENCRYPTED_FILE}
done

##########

if [ "$(ls -A ${TMP_DIR})" ]; then
    kubectl create secret generic ${SECRET_NAME} --from-file=${TMP_DIR}
fi