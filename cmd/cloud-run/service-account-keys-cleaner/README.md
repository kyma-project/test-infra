# Cleanup of Service Account Secrets

## Overview

The Cloud Run service deletes old keys for a Google Cloud service account and updates the required secret data for all service account secrets stored in the Secret Manager. The service is triggered by a Cloud Scheduler job.

1. Cloud Scheduler calls the service-account-keys-cleaner service.
2. For each secret stored in Secret Manager, the service executes the following steps:
    1. Checks if the value of the **type** label is set to `service-account`. If not, it stops running.
    2. Checks if the value of the **skip-cleanup** label is set to `true`. If it is, the service stops running.
    3. Reads the name of the service account from the latest version of a secret.
    4. Checks if the latest secret version is older than the time in hours set in the **age** GET parameter. If not, it stops running.
    5. Removes old versions of keys for the service account.
    6. Removes old versions of a secret stored in Secret Manager.

## Cloud Run Service Deployment

ServiceAccountKeysCleaner is deployed to Cloud Run applying Terraform config stored
in [`./terraform` directory](../../../configs/terraform). `terraform apply` runs automatically on every PR changing
Terraform `.tf` files belonging to the application.

1. Create the `service-${PROJECT_NUMBER}@gcp-sa-secretmanager.iam.gserviceaccount.com` service account with the `roles/pubsub.publisher` role if it does not exist.
2. Merge your changes to test-infra main branch to trigger Terraform execution.


## GET Parameters

The Cloud Function accepts the following GET parameters:

| Name                           | Required | Description                                                           |
| :----------------------------- | :------: | :-------------------------------------------------------------------- |
| **project**                    |    Yes   | The name of the Google Cloud project with Secret Manager.|
| **age**                        |    No    | The age in hours that the latest version of a secret has to exist before old versions can be deleted. It defaults to `5`. |
| **dry_run**                    |    No    | The value controlling the `dry run` mode. It defaults to `false`.|
