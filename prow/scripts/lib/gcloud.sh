#!/usr/bin/env bash

function gcloud::authenticate() {
    echo "Authenticating"
    gcloud auth activate-service-account --key-file "${GOOGLE_APPLICATION_CREDENTIALS}" || exit 1
}
