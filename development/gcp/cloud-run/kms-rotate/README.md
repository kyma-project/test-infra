# Rotate KMS secrets using Cloud Run

## Overview

The Cloud Run application decrypts and encrypts files in a bucket with the latest version of a KMS key, and deletes old versions of a key. The function is triggered by a HTTP POST request sent by a Cloud Scheduler.

1. A job in Cloud Scheduler sends a POST request to the Cloud Run application.
2. The Cloud Run application checks if there is more than one enabled key version. If not, it stops running.
3. The Cloud Run application decrypts and encrypts files in the bucket with the latest version of the key.
4. The Cloud Run application marks the old versions of a KMS key for destruction.


## Cloud Run deployment

To deploy the Cloud Run application, follow these steps:

1. Use the following command to deploy the Cloud Run application:
```bash
gcloud run deploy rotate-kms-key \
--region europe-west1 \
--timeout 600 \
--max-instances 1 \
--memory 128Mi \
--service-account sa-kms-update@sap-kyma-prow.iam.gserviceaccount.com \
--ingress all \
--project sap-kyma-prow \
--image europe-docker.pkg.dev/kyma-project/prod/test-infra/kms-rotate:v20221025-47772933
```
2. Create a new job in Cloud Scheduler that calls the Cloud Run endpoint with JSON config passed as a POST body.


# JSON request parameters

See the list of JSON arguments for the function:
| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **project** | Yes | Name of the CGP project containing the KMS key. |
| **location** | Yes | Name of the CGP location where the KMS key is stored. |
| **bucketName** | Yes | Name of the CGP bucket containing the files to be re-encrypted. |
| **bucketPrefix** | No | Prefix of the files stored in the bucket used to filter them out. |
| **keyring** | Yes | Name of the keyring containing the KMS key. |
| **key** | Yes | Name of the KMS key. |


