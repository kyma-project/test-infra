# Prow Secrets Management

## Overview

Some jobs require using sensitive data. Encrypt the data using Key Management Service (KMS) and store it in Google Cloud Storage (GCS).
This document shows the commands necessary to create a service account and store its encrypted key in a GCS bucket.

>**NOTE:** This document assumes that you are logged in to the Google Cloud project with administrative rights.

## Prerequisites

 - [gcloud](https://cloud.google.com/sdk/gcloud/) to communicate with Google Cloud Platform (GCP)
 - Basic knowledge of [GCP key rings and keys](https://cloud.google.com/kms/docs/creating-keys)

Use the `export {VARIABLE}={value}` command to set up these variables, where:
 - **PROJECT** is a Google Cloud project.
 - **BUCKET_NAME** is a GCS bucket in the Google Cloud project that stores Prow Secrets
 - **KEYRING_NAME** is the KMS key ring.
 - **ENCRYPTION_KEY_NAME** is the key name in the key ring that is used for data encryption.
 - **LOCATION** is the geographical location of the data center that handles requests for Cloud KMS regarding a given resource and stores the corresponding cryptographic keys. When set to `global`, your Cloud KMS resources are available from multiple data centres.

## Secrets management

>**NOTE:** Before you follow this guide, check Prow Secrets setup for the Google Cloud project.

When you communicate for the first time with Google Cloud, set the context to your Google Cloud project. Run this command:
```
gcloud config set project $PROJECT
```

### Create a GCS bucket

The purpose of the bucket is to store encrypted credentials necessary for Prow jobs like provisioning clusters or virtual machines.
Run this command to create a bucket:
```
gsutil mb -p $PROJECT gs://$BUCKET_NAME/
```

### Create a key ring

Use this command to create a key ring for the private keys:

```
gcloud kms keyrings create $KEYRING_NAME --location $LOCATION
```
### Create a key in the key ring

Create a key to encrypt your private key.

```
gcloud kms keys create $ENCRYPTION_KEY_NAME --location $LOCATION \
  --keyring $KEYRING_NAME --purpose encryption
  ```

### Create a Google service account

Follow these steps:

1. Export the variables, where:
   - **SA_NAME** is the name of the service account.
   - **SA_DISPLAY_NAME** is the display name of the service account.
   - **SECRET_FILE** is the path to the private key.
   - **ROLE** is the role bound to the service account.

   See an example of variables you must export for such an account:

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

1. Export the **SECRET_FILE** variable which is the path to the file which contains the Secret.

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
