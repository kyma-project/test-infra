# Prow Secrets Management

## Overview

Some jobs require using sensitive data. You need to encrypt data using Key Management Service (KMS) and store them in Google Cloud Storage (GCS).
This document shows the commands necessary to create a service account and store its encrypted key in a GCS bucket.

>**NOTE:** This document assumes that you are logged in to the Google Cloud project with administrative rights.

## Prerequisites

 - [gcloud](https://cloud.google.com/sdk/gcloud/) to communicate with Google Cloud Platform.
 - Basic knowledge of [GCP key rings and keys](https://cloud.google.com/kms/docs/creating-keys).

Use the `export {VARIABLE}={value}` command to set up these variables, where:
 - **PROJECT** is a Google Cloud project.
 - **BUCKET_NAME** is a GCS bucket in the Google Cloud project that is used to store Prow Secrets.
 - **KEYRING_NAME** is the KMS key ring.
 - **ENCRYPTION_KEY_NAME** is the key name in the key ring that is used for data encryption.
 - **LOCATION** is the geographical data center location where requests to Cloud KMS regarding a given resource are handled, and where the corresponding cryptographic keys are stored. When set to `global`, your Cloud KMS resources are available from multiple data centres.

## Secrets management

>**NOTE:** Before you follow this guide, check Prow Secrets setup for the Google Cloud project.

When you communicate for the first time with the Google Cloud, set the context to your Google Cloud project. Execute this command:
```
gcloud config set project $PROJECT
```

### Create a GCS bucket

The purpose of the bucket is to store encrypted credentials necessary for Prow jobs like provisioning clusters or virtual machines.
Run this command to create a bucket:
```
gsutil mb -p $PROJECT gs://$BUCKET_NAME/
```

### Create a keyring

Use this command to create a keyring for the private keys:

```
gcloud kms keyrings create $KEYRING_NAME --location $LOCATION
```

### Create a Google service account

Follow these steps:

1. Export the variables, where:
 - **SA_NAME** is the name of the service account.
 - **SA_DISPLAY_NAME** is the display name of the service account.
 - **SECRET_FILE** is the path for the private key.
 - **ROLE** is the role bound to the service account.

 See an example of variables you need to export for each account:

 ```
 export SA_NAME=sa-gcs-plank
 export SA_DISPLAY_NAME=sa-gcs-plank
 export SECRET_FILE=sa-gcs-plank
 export ROLE=roles/storage.objectAdmin

 ```

2. Create a service account:
```
gcloud iam service-accounts create $SA_NAME --display-name $SA_DISPLAY_NAME
```

3. Create a private key for the service account:
```
gcloud iam service-accounts keys create $SECRET_FILE --iam-account=$SA_NAME@$PROJECT.iam.gserviceaccount.com
```

4. Add a policy binding to the service account:
```
gcloud projects add-iam-policy-binding $PROJECT --member=serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com --role=$ROLE
```

### Encrypt the Secret

1. Export the **SECRET_FILE** variable which is the path to the file containing the Secret.

2. Encrypt the Secret:
```
gcloud kms encrypt --location global --keyring $KEYRING_NAME --key $ENCRYPTION_KEY_NAME --plaintext-file $SECRET_FILE --ciphertext-file $SECRET_FILE.encrypted
```

### Upload the Secret

Upload the encrypted Secret to GCP:
```
gsutil cp $SECRET_FILE.encrypted gs://$BUCKET_NAME/
```

### Delete the Secret

Delete the private key files:

```
rm {file-name}
rm {file-name}.encrypted
```
