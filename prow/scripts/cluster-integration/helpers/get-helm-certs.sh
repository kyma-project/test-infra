#!/usr/bin/env bash

# Description: Waits for the cluster to generate Helm Client Certs and downloads them into HELM_HOME

function getHelmCerts() {
    RETRY_COUNT=3
    RETRY_TIME_SEC=5

    for (( i = 0; i < RETRY_COUNT; i++ )); do
        mkdir -p "$(helm home)"

        echo "---> Get Helm secrets and put then into $(helm home)"
        kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 --decode > "$(helm home)/ca.pem"
        kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 --decode > "$(helm home)/cert.pem"
        kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 --decode > "$(helm home)/key.pem"

        if [[ "${i}" -lt "${RETRY_COUNT}" ]]; then
            echo "---> Unable to get Helm Certs. Waiting for ${RETRY_TIME_SEC}. Attempt ${i} of ${RETRY_COUNT}"
        else
            echo "---> Unable to get Helm Certs after ${RETRY_COUNT} attempts. Exitting"
            exit 1
        fi

        sleep "${RETRY_TIME_SEC}"
    done
    exit 0
}