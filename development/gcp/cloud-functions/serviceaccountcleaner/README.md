# Cleanup of service account secrets using Cloud Function

## Overview

The Cloud Function deletes old keys for a GCP service account and updates the required secret data for all service account secrets stored in the Secret Manager. The function is triggered by a Cloud Scheduler job.

1. Cloud Scheduler starts the Cloud Function.
2. For each secret stored in Secret Manager:
    1. The Cloud Function checks if the value of the **type** label is set to `service-account`. If not, it stops running.
    2. The Cloud Function checks if the value of the **skip-cleanup** label is not set to `true`. If not, it stops running.
    3. The Cloud Function reads the name of the service account from the latest version of a secret.
    4. The Cloud function checks if the latest secret version is older than the time in hours set in the **age** GET parameter. If not, it stops running.
    5. The Cloud Function removes old versions of keys for the service account.
    6. The Cloud Function removes old versions of a secret stored in Secret Manager.

## Cloud Function deployment

To deploy Cloud Function follow these steps:

1. Run `go mod vendor` inside the `development/gcp/cloud-functions/rotateserviceaccount/` directory.
2. Create the `secret-manager-notifications` Pub/Sub topic if it does not exist.
3. Create the `service-${PROJECT_NUMBER}@gcp-sa-secretmanager.iam.gserviceaccount.com` service account with the `roles/pubsub.publisher` role if it does not exist.
4. Use the following command to deploy the Cloud Function:
```bash
gcloud functions deploy rotate-secrets-service-account-cleaner \
--region europe-west3 \
--trigger-http \
--runtime go116 \
--source ./ \
--timeout 10 \
--max-instances 10 \
--memory 128 \
--entry-point ServiceAccountCleaner
```
5. Set up a new job in Cloud Scheduler.


## GET parameters

The Cloud Function accepts the following GET parameters:

| Name                           | Required | Description                                                           |
| :----------------------------- | :------: | :-------------------------------------------------------------------- |
| **project**                    |    Yes   | The name of the GCP project with Secret Manager.|
| **age**                        |    No    | The age in hours that the latest version of a secret has to exist before old versions can be deleted. It defaults to `5`. |
| **dry_run**                    |    No    | The value controlling the `dry run` mode. It defaults to `false`.|
