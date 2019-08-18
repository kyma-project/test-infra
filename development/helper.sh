#!/bin/bash

# Adjust clustername if you want, this should not collide with any cluster on GKE that already exists
if [ -z "${CLUSTER_NAME}" ]; then
    CLUSTER_NAME=prow-$(whoami)
fi

#Add check and set for workload cluster.

# temporary secret store location, should be deleted after installing the cluster
SECRET_FOLDER=.secrets

if [ -z "${OAUTH}" ]; then
    OAUTH="FILL_ME_IN"
fi
if [ -z "${PROJECT}" ]; then
    PROJECT="FILL_ME_IN"
fi
if [ -z "${ZONE}" ]; then
    ZONE="FILL_ME_IN"
fi
if [ -z "${LOCATION}" ]; then
    LOCATION="FILL_ME_IN" # key locations for KMS
fi
if [ -z "${BUCKET_NAME}" ]; then
    BUCKET_NAME="FILL_ME_IN"
fi
if [ -z "${KEYRING_NAME}" ]; then
    KEYRING_NAME="FILL_ME_IN"
fi
if [ -z "${ENCRYPTION_KEY_NAME}" ]; then
    ENCRYPTION_KEY_NAME="FILL_ME_IN"
fi
if [ -z "${KUBECONFIG}" ]; then
    KUBECONFIG="FILL_ME_IN" # e.g. /Users/sample-user/.kube/config
fi

####
#### DO NOT CHANGE ANYTHING BELOW THIS LINE
####
export CLUSTER_NAME
export SECRET_FOLDER

export OAUTH
export PROJECT
export ZONE
export LOCATION
export BUCKET_NAME
export KEYRING_NAME
export ENCRYPTION_KEY_NAME
export KUBECONFIG

if [ $PROJECT = "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $ZONE = "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $LOCATION = "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $OAUTH = "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $BUCKET_NAME = "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $KEYRING_NAME = "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $ENCRYPTION_KEY_NAME = "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $KUBECONFIG = "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

check_and_trim() {
    str=$1
    len=$2
    lenInc=$((len + 1))
    if [[ ${#str} -ge $lenInc ]]; then
        while : ; do
            str=${str:0:${#str} - 1}
            if [[ ${#str} -le $len ]]; then
                break
            fi
        done
    fi
    echo "$str"
}
