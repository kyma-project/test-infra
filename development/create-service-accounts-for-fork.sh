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

    if [[ ! $SHORT_IDENTIFIER =~ $ident_regexp ]]; then
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

SA_REAL_NAME=sa-gcs-plank
SA_NAME=$(echo ${SA_REAL_NAME} | cut -c1-19)-$SHORT_IDENTIFIER
SA_DISPLAY_NAME=$SA_REAL_NAME-$SHORT_IDENTIFIER
SECRET_FILE=$SECRET_FOLDER/$SA_REAL_NAME
#SA_NAME=${SA_REAL_NAME}
#SA_DISPLAY_NAME=${SA_REAL_NAME}

SA_NAME=$(check_and_trim "$SA_NAME" 30)

gcloud iam service-accounts create "$SA_NAME" --display-name "$SA_DISPLAY_NAME"
gcloud iam service-accounts keys create "$SECRET_FILE" --iam-account="$SA_NAME@$PROJECT.iam.gserviceaccount.com"

ROLE="roles/storage.objectAdmin"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

gcloud kms encrypt --location "$LOCATION" --keyring "$KEYRING_NAME" --key "$ENCRYPTION_KEY_NAME" --plaintext-file "$SECRET_FILE" --ciphertext-file "$SECRET_FILE.encrypted"
gsutil cp "$SECRET_FILE.encrypted" "gs://$BUCKET_NAME/"

###
read -n 1 -s -r -p "Press any key to continue"
###
### Create sa-gke-kyma-integration
###

SA_REAL_NAME=sa-gke-kyma-integration
SA_NAME=$(echo ${SA_REAL_NAME} | cut -c1-19)-$SHORT_IDENTIFIER
SA_DISPLAY_NAME=$SA_REAL_NAME-$SHORT_IDENTIFIER
SECRET_FILE=$SECRET_FOLDER/$SA_REAL_NAME
#SA_NAME=${SA_REAL_NAME}
#SA_DISPLAY_NAME=${SA_REAL_NAME}

SA_NAME=$(check_and_trim "$SA_NAME" 30)

gcloud iam service-accounts create "$SA_NAME" --display-name "$SA_DISPLAY_NAME"
gcloud iam service-accounts keys create "$SECRET_FILE" --iam-account="$SA_NAME@$PROJECT.iam.gserviceaccount.com"

ROLE="roles/compute.admin"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

ROLE="roles/container.admin"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

ROLE="roles/container.clusterAdmin"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

ROLE="roles/dns.admin"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

ROLE="roles/iam.serviceAccountUser"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

ROLE="roles/storage.admin"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

gcloud kms encrypt --location "$LOCATION" --keyring "$KEYRING_NAME" --key "$ENCRYPTION_KEY_NAME" --plaintext-file "$SECRET_FILE" --ciphertext-file "$SECRET_FILE.encrypted"
gsutil cp "$SECRET_FILE.encrypted" "gs://$BUCKET_NAME/"

###
read -n 1 -s -r -p "Press any key to continue"
###
### Create sa-vm-kyma-integration
###

SA_REAL_NAME=sa-vm-kyma-integration
SA_NAME=$(echo ${SA_REAL_NAME} | cut -c1-19)-$SHORT_IDENTIFIER
SA_DISPLAY_NAME=$SA_REAL_NAME-$SHORT_IDENTIFIER
SECRET_FILE=$SECRET_FOLDER/$SA_REAL_NAME
#SA_NAME=${SA_REAL_NAME}
#SA_DISPLAY_NAME=${SA_REAL_NAME}

SA_NAME=$(check_and_trim "$SA_NAME" 30)

gcloud iam service-accounts create "$SA_NAME" --display-name "$SA_DISPLAY_NAME"
gcloud iam service-accounts keys create "$SECRET_FILE" --iam-account="$SA_NAME@$PROJECT.iam.gserviceaccount.com"

ROLE="roles/compute.instanceAdmin"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

ROLE="roles/compute.osAdminLogin"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

ROLE="roles/iam.serviceAccountUser"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

gcloud kms encrypt --location "$LOCATION" --keyring "$KEYRING_NAME" --key "$ENCRYPTION_KEY_NAME" --plaintext-file "$SECRET_FILE" --ciphertext-file "$SECRET_FILE.encrypted"
gsutil cp "$SECRET_FILE.encrypted" "gs://$BUCKET_NAME/"

###
read -n 1 -s -r -p "Press any key to continue"
###
### Create sa-gcr-push-kyma-project
###

SA_REAL_NAME=sa-gcr-push-kyma-project
SA_NAME=$(echo ${SA_REAL_NAME} | cut -c1-19)-$SHORT_IDENTIFIER
SA_DISPLAY_NAME=$SA_REAL_NAME-$SHORT_IDENTIFIER
SECRET_FILE=$SECRET_FOLDER/$SA_REAL_NAME
#SA_NAME=${SA_REAL_NAME}
#SA_DISPLAY_NAME=${SA_REAL_NAME}

SA_NAME=$(check_and_trim "$SA_NAME" 30)

gcloud iam service-accounts create "$SA_NAME" --display-name "$SA_DISPLAY_NAME"
gcloud iam service-accounts keys create "$SECRET_FILE" --iam-account="$SA_NAME@$PROJECT.iam.gserviceaccount.com"

ROLE="roles/storage.objectCreator"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

gcloud kms encrypt --location "$LOCATION" --keyring "$KEYRING_NAME" --key "$ENCRYPTION_KEY_NAME" --plaintext-file "$SECRET_FILE" --ciphertext-file "$SECRET_FILE.encrypted"
gsutil cp "$SECRET_FILE.encrypted" "gs://$BUCKET_NAME/"

###
read -n 1 -s -r -p "Press any key to continue"
###
### Create kyma-bot-npm-token
###

SA_REAL_NAME=kyma-bot-npm-token
SA_NAME=$(echo ${SA_REAL_NAME} | cut -c1-19)-$SHORT_IDENTIFIER
SA_DISPLAY_NAME=$SA_REAL_NAME-$SHORT_IDENTIFIER
SECRET_FILE=$SECRET_FOLDER/$SA_REAL_NAME
#SA_NAME=${SA_REAL_NAME}
#SA_DISPLAY_NAME=${SA_REAL_NAME}

SA_NAME=$(check_and_trim "$SA_NAME" 30)

gcloud iam service-accounts create "$SA_NAME" --display-name "$SA_DISPLAY_NAME"
gcloud iam service-accounts keys create "$SECRET_FILE" --iam-account="$SA_NAME@$PROJECT.iam.gserviceaccount.com"

gcloud kms encrypt --location "$LOCATION" --keyring "$KEYRING_NAME" --key "$ENCRYPTION_KEY_NAME" --plaintext-file "$SECRET_FILE" --ciphertext-file "$SECRET_FILE.encrypted"
gsutil cp "$SECRET_FILE.encrypted" "gs://$BUCKET_NAME/"

###
read -n 1 -s -r -p "Press any key to continue"
###
### Create sap-slack-bot-token
###

SA_REAL_NAME=sap-slack-bot-token
SA_NAME=$(echo ${SA_REAL_NAME} | cut -c1-19)-$SHORT_IDENTIFIER
SA_DISPLAY_NAME=$SA_REAL_NAME-$SHORT_IDENTIFIER
SECRET_FILE=$SECRET_FOLDER/$SA_REAL_NAME
#SA_NAME=${SA_REAL_NAME}
#SA_DISPLAY_NAME=${SA_REAL_NAME}

SA_NAME=$(check_and_trim "$SA_NAME" 30)

gcloud iam service-accounts create "$SA_NAME" --display-name "$SA_DISPLAY_NAME"
gcloud iam service-accounts keys create "$SECRET_FILE" --iam-account="$SA_NAME@$PROJECT.iam.gserviceaccount.com"

gcloud kms encrypt --location "$LOCATION" --keyring "$KEYRING_NAME" --key "$ENCRYPTION_KEY_NAME" --plaintext-file "$SECRET_FILE" --ciphertext-file "$SECRET_FILE.encrypted"
gsutil cp "$SECRET_FILE.encrypted" "gs://$BUCKET_NAME/"

###
read -n 1 -s -r -p "Press any key to continue"
###
### Create sa-kyma-artifacts
###

SA_REAL_NAME=sa-kyma-artifacts
SA_NAME=$(echo ${SA_REAL_NAME} | cut -c1-19)-$SHORT_IDENTIFIER
SA_DISPLAY_NAME=$SA_REAL_NAME-$SHORT_IDENTIFIER
SECRET_FILE=$SECRET_FOLDER/$SA_REAL_NAME
#SA_NAME=${SA_REAL_NAME}
#SA_DISPLAY_NAME=${SA_REAL_NAME}

SA_NAME=$(check_and_trim "$SA_NAME" 30)

gcloud iam service-accounts create "$SA_NAME" --display-name "$SA_DISPLAY_NAME"
gcloud iam service-accounts keys create "$SECRET_FILE" --iam-account="$SA_NAME@$PROJECT.iam.gserviceaccount.com"

ROLE="roles/storage.objectAdmin"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

gcloud kms encrypt --location "$LOCATION" --keyring "$KEYRING_NAME" --key "$ENCRYPTION_KEY_NAME" --plaintext-file "$SECRET_FILE" --ciphertext-file "$SECRET_FILE.encrypted"
gsutil cp "$SECRET_FILE.encrypted" "gs://$BUCKET_NAME/"

###
read -n 1 -s -r -p "Press any key to continue"
###
### Create sa-kms-storage account. This is needed and will be passed as GOOGLE_APPLICATION_CREDENTIALS for the installer
###

SA_REAL_NAME=sa-kms-storage
SA_NAME=$(echo ${SA_REAL_NAME} | cut -c1-19)-$SHORT_IDENTIFIER
SA_DISPLAY_NAME=$SA_REAL_NAME-$SHORT_IDENTIFIER
SECRET_FILE=$SECRET_FOLDER/$SA_REAL_NAME
#SA_NAME=${SA_REAL_NAME}
#SA_DISPLAY_NAME=${SA_REAL_NAME}

SA_NAME=$(check_and_trim "$SA_NAME" 30)

gcloud iam service-accounts create "$SA_NAME" --display-name "$SA_DISPLAY_NAME"
gcloud iam service-accounts keys create "$SECRET_FILE" --iam-account="$SA_NAME@$PROJECT.iam.gserviceaccount.com"

ROLE="roles/storage.objectAdmin"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

ROLE="roles/cloudkms.cryptoKeyDecrypter"
gcloud projects add-iam-policy-binding "$PROJECT" --member="serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com" --role="$ROLE"

gcloud kms encrypt --location "$LOCATION" --keyring "$KEYRING_NAME" --key "$ENCRYPTION_KEY_NAME" --plaintext-file "$SECRET_FILE" --ciphertext-file "$SECRET_FILE.encrypted"
gsutil cp "$SECRET_FILE.encrypted" "gs://$BUCKET_NAME/"


GOOGLE_APPLICATION_CREDENTIALS=$SECRET_FILE
export GOOGLE_APPLICATION_CREDENTIALS

###
read -n 1 -s -r -p "Done. You can run the installer now. Make sure you delete the temp secret folder '${SECRET_FOLDER}' after you're done with the installer."
###
###
###
