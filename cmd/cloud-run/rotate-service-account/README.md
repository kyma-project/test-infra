# Rotate Service Account Secrets

## Overview

RotateServiceAccount creates a new key for a Google Cloud service account and updates the required secret data. It's triggered by a  Pub/Sub message sent by a secret stored in Secret Manager. It runs as a cloud run container.

1. Secret Manager sends a Pub/Sub message to `secret-manager-notifications` Pub/Sub topic.
3. RotateServiceAccount checks if the value of the **eventType** attribute is set to `SECRET_ROTATE`; if not, it stops its execution.
4. RotateServiceAccount checks if the value of the **type** label is set to `service-account`; if not, it stops its execution.
5. RotateServiceAccount reads the name of the service account from the latest version of a secret.
6. RotateServiceAccount generates a new key for the service account.
7. RotateServiceAccount creates a new secret version in Secret Manager, containing the newly created service account key.

## Cloud Run Deployment

RotateServiceAccount is deployed to Cloud Run applying Terraform config stored
in the [`./terraform` directory](../../../configs/terraform). `terraform apply` is executed automatically on every PR changing Terraform `.tf` files belonging to the application.

## RotateServiceAccount Usage

To setup an automatic rotation for a Secret Manager secret, follow these steps:
1. Create a new secret in Secret Manager with the existing service account data.
2. Add the `type: service-account` label to the secret.
3. Set `secret-manager-notifications` as the secret Pub/Sub topic.
4. Set up a rotation period for the secret.
