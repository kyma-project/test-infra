# Prow Secrets management

## Overview

Some jobs require using sensitive data. You need to encrypt data using Key Management Service (KMS) and store them in Google Cloud Storage (GCS).
This document shows the commands necessary to create a service account and store its encrypted key in a GCS bucket. This document assumes that you are logged in to the Google Cloud project with administrative rights.

## Prerequisites

 - [gcloud](https://cloud.google.com/sdk/gcloud/) 
 - Basic knowledge about [GCP key rings and keys](https://cloud.google.com/kms/docs/creating-keys).

Use the `export {VARIABLE}={value}` command to set up these variables, where:
 - **PROJECT_NAME** is a Google Cloud project.
 - **BUCKET_NAME** is a GCS bucket in the Google Cloud project that is used to store Prow Secrets.
 - **KEYRING_NAME** is the KMS key ring.
 - **ENCRYPTION_KEY_NAME** is the key name in the key ring that is used for data encryption.

## Secrets management

>**NOTE:** Before you follow this guide, check Prow Secrets setup for the Google Cloud project.

Execute this command to set the context to the Google Cloud project:
```
gcloud config set project $PROJECT_NAME
```

### Create a GCS bucket

The purpose of the bucket is to store encrypted credentials necessary for Prow jobs like provisioning clusters or virtual machines.
Run this command to create a bucket:
```
gsutil mb -p $PROJECT_NAME gs://$BUCKET_NAME/
```

### Create a Google service account

Export the variables, where:
 - **SA_NAME** is the name of the service account.
 - **SA_DISPLAY_NAME** is the display name of the service account.
 - **SECRET_FILE** is the path for the private key.
 - **ROLE** is the role bound to the service account.

Create a service account:
```
gcloud iam service-accounts create $SA_NAME --display-name $SA_DISPLAY_NAME
```

Create a private key for the service account:
```
gcloud iam service-accounts keys create $SECRET_FILE --iam-account=SA_NAME
```

Add a policy binding for the service account:
```
gcloud iam service-accounts add-iam-policy-binding $SA_NAME --member=serviceAccount:$SA_NAME@$PROJECT_NAME.iam.gserviceaccount.com --role=$ROLE
```

### Encrypt the Secret

Export the following:
 - SECRET_FILE - path to the file containing secret

Encrypt the Secret:
```
gcloud kms encrypt --location global --keyring $KEYRING_NAME --key $ENCRYPTION_KEY_NAME --plaintext-file $SECRET_FILE --ciphertext-file $SECRET_FILE.encrypted
```

### Upload the secret

Upload the encrypted Secret to GCP:
```
gsutil cp $SECRET_FILE.encrypted gs://$BUCKET_NAME/
```

Delete the file exported under the $SECRET_FILE variable.
