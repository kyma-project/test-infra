# Rotate service account secrets using Cloud Function

## Overview

Cloud Function creates a new key for a GCP service account and updates the required secret data. The function is triggered by a  Pub/Sub message sent by a secret stored in Secret Manager.

1. Secret in Secret Manager sends a Pub/Sub message to `secret-manager-notifications` Pub/Sub topic.
2. Cloud Function is started.
3. Cloud Function checks if the value of the **eventType** attribute is set to `SECRET_ROTATE`; if not, it stops its execution.
4. Cloud Function checks if the value of the **type** label is set to `service-account`; if not, it stops its execution.
5. Cloud Function reads the name of the service account from the latest version of a secret.
6. Cloud Function generates a new key for the service account.
7. Cloud Function creates a new secret version in Secret Manger, containing the newly created service account key.

## Cloud Function deployment

To deploy Cloud Function follow these steps:

1. Run `go mod vendor` inside the `development/gcp/cloud-functions/rotateserviceaccount/` directory.
2. Create the `secret-manager-notifications` Pub/Sub topic, if it does not exist.
3. Create the `service-${PROJECT_NUMBER}@gcp-sa-secretmanager.iam.gserviceaccount.com` service account with the `roles/pubsub.publisher` role, if it does not exist.
4. Use the following command from this directory to deploy Cloud Function:
```bash
gcloud functions deploy rotate-secrets-service-account \
--region europe-west3 \
--trigger-topic secret-manager-notifications \
--runtime go116 \
--source ./ \
--timeout 10 \
--max-instances 10 \
--memory 128 \
--entry-point RotateServiceAccount
```

## Cloud Function usage

To setup an automatic rotation for a Secret Manager secret follow these steps:
1. Create a new secret in Secret Manager with the existing service account data.
2. Add `type: service-account` label to the secret.
3. Set `secret-manager-notifications` as a secret Pub/Sub topic.
4. Set up rotation period for the secret.
