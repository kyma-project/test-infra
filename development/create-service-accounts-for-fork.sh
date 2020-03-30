#!/bin/bash

set -o errexit


DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
# shellcheck source=development/helper.sh
. "$DIR/helper.sh"

if [ -z "$BUCKET_NAME" ]; then
      echo "\$BUCKET_NAME is empty"
      exit 1
fi

if [ -z "$KEYRING_NAME" ]; then
      echo "\$KEYRING_NAME is empty"
      exit 1
fi

if [ -z "$ENCRYPTION_KEY_NAME" ]; then
      echo "\$ENCRYPTION_KEY_NAME is empty"
      exit 1
fi

if [ -z "$PROJECT" ]; then
      echo "\$PROJECT is empty"
      exit 1
fi

if [ -z "$SECRET_FOLDER" ]; then
      echo "\$SECRET_FOLDER is empty"
      exit 1
fi

if [ -z "$LOCATION" ]; then
    LOCATION="global"
fi

mkdir -p "$SECRET_FOLDER"

####
#### Read short identifier from user input to go on with the script
####

ident_regexp="^[-0-9A-Za-z]{1,10}$"

while : ; do
    read -r -p "Enter a short handle for the cluster [10 characters, only alphanumeric, numbers and '-' allowed]: " SHORT_IDENTIFIER

    if [[ -z  $SHORT_IDENTIFIER ]]; then
    	read -r -i "n" -p "Short handle for the cluster is empty. Would you like to proceed without it?: [Y/n] " AGREEMENT
    	if [[ $AGREEMENT == "Y" ]]; then
    		break
    	else
    		continue
    	fi
    elif [[ ! $SHORT_IDENTIFIER =~ $ident_regexp ]]; then
        echo "Can only contain numbers, alphanumeric characters and '-', should be less than 10 characters"
    else
        break
    fi
done

gsutil mb -p ${PROJECT} gs://${BUCKET_NAME}/
gcloud kms keyrings create ${KEYRING_NAME} --location ${LOCATION}
gcloud kms keys create ${ENCRYPTION_KEY_NAME} --location ${LOCATION} --keyring ${KEYRING_NAME} --purpose encryption

####
#### Create sa-gcs-plank service account
####

function create_service_account() {
# positional arguments
# - service account name
# - role
	SA_REAL_NAME=$1
	if [[ -z $SHORT_IDENTIFIER ]]; then
		SA_NAME=${SA_REAL_NAME}
		SA_DISPLAY_NAME=${SA_REAL_NAME}
	else
		SA_NAME=$(echo "${SA_REAL_NAME}" | cut -c1-19)-$SHORT_IDENTIFIER
		SA_DISPLAY_NAME=$SA_REAL_NAME-$SHORT_IDENTIFIER
	fi
	SECRET_FILE=$SECRET_FOLDER/$SA_REAL_NAME

	SA_NAME=$(check_and_trim "$SA_NAME" 40)

	gcloud iam service-accounts create "$SA_NAME" --display-name "$SA_DISPLAY_NAME"
	gcloud iam service-accounts keys create "$SECRET_FILE" --iam-account="$SA_NAME@$PROJECT.iam.gserviceaccount.com"

	ROLE=$2
	gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

	gcloud kms encrypt --location "$LOCATION" --keyring "$KEYRING_NAME" --key "$ENCRYPTION_KEY_NAME" --plaintext-file "$SECRET_FILE" --ciphertext-file "$SECRET_FILE.encrypted"
	gsutil cp "$SECRET_FILE.encrypted" "gs://$BUCKET_NAME/"

	###
	read -n 1 -s -r -p "Press any key to continue"
}

# Arrays should be populated by reading required secrets file.
#declare -a SA_GCS_PLANK=("roles/storage.objectAdmin")
#declare -a SA_GKE_KYMA_INTEGRATION=("roles/compute.admin" "roles/container.admin" "roles/container.clusterAdmin" "roles/dns.admin" "roles/iam.serviceAccountUser" "roles/storage.admin")
#declare -a SA_VM_KYMA_INTEGRATION=("roles/compute.instanceAdmin" "roles/compute.osAdminLogin" "roles/iam.serviceAccountUser")
#declare -a SA_GCR_PUSH_KYMA_PROJECT=("roles/storage.objectCreator")
#declare -a SA_KYMA_ARTIFACTS=("roles/storage.objectAdmin")
#declare -A SERVICE_ACCOUNTS
#SERVICE_ACCOUNTS["sa-gcs-plank"]=$SA_GCS_PLANK
#SERVICE_ACCOUNTS["sa-gke-kyma-integration"]=$SA_GKE_KYMA_INTEGRATION
#SERVICE_ACCOUNTS["sa-vm-kyma-integration"]=$SA_VM_KYMA_INTEGRATION
#SERVICE_ACCOUNTS["sa-gcr-push-kyma-project"]=$SA_GCR_PUSH_KYMA_PROJECT
#SERVICE_ACCOUNTS["sa-kyma-artifacts"]=$SA_KYMA_ARTIFACTS


# TODO: This script should generate all mandatory secrets
## Create an HMAC token
#hmac_token="$(openssl rand -hex 32)"
#echo "$hmac_token" > hmac_token.txt
#echo "Token hmac stored in hmac_token.txt file"

#if [ "$OAUTH" == "" ]; then
#    echo -n "Enter OAuth2 token that has read and write access to the bot account, followed by [ENTER]: (input will not be printed)"
#    read -rs oauth_token
#else
#    oauth_token="$OAUTH"
#fi

#echo

#if [ ${#oauth_token} -lt 1 ]; then
#  echo "OAuth2 token not provided";
#  exit -1;
#fi

#kubectl create secret generic hmac-token --from-literal=hmac="$hmac_token"
#kubectl create secret generic oauth-token --from-literal=oauth="$oauth_token"

### Create sa-kms-storage account. This is needed and will be passed as GOOGLE_APPLICATION_CREDENTIALS for the installer
###

#SA_REAL_NAME=sa-kms-storage
#ROLE="roles/storage.objectAdmin"
#ROLE="roles/cloudkms.cryptoKeyDecrypter"

# This doesn't make sense as env variable will disappear when script finish.
#GOOGLE_APPLICATION_CREDENTIALS=$SECRET_FILE
#export GOOGLE_APPLICATION_CREDENTIALS

###
read -n 1 -s -r -p "Done. You can run the installer now. Make sure you delete the temp secret folder '${SECRET_FOLDER}' after you're done with the installer."
###
###
###
