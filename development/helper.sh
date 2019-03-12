#!/bin/bash

# Adjust clustername if you want, this should not collide with any cluster on GKE that already exists
export CLUSTER_NAME=prow-`whoami`

# temporary secret store location, should be deleted after installing the cluster
export SECRET_FOLDER=.secrets

export OAUTH="FILL_ME_IN"
export PROJECT="FILL_ME_IN"
export ZONE="FILL_ME_IN"
export LOCATION="FILL_ME_IN" # key locations for KMS
export BUCKET_NAME="FILL_ME_IN"
export KEYRING_NAME="FILL_ME_IN"
export ENCRYPTION_KEY_NAME="FILL_ME_IN"
export KUBECONFIG="FILL_ME_IN" # e.g. /Users/sample-user/.kube/config

####
#### DO NOT CHANGE ANYTHING BELOW THIS LINE
####

if [ $PROJECT -eq "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $ZONE -eq "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $LOCATION -eq "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $OAUTH -eq "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $BUCKET_NAME -eq "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $KEYRING_NAME -eq "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $ENCRYPTION_KEY_NAME -eq "FILL_ME_IN" ]; then
    echo "Please edit development/helper.sh and change variables with FILL_ME_IN to reflect your needs."
    exit 1
fi

if [ $KUBECONFIG -eq "FILL_ME_IN" ]; then
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
    echo $str
}