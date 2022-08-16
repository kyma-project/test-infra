# Rotate service account cloud function


## Overview

This cloud function creates new key for GCP service account on Secret Manager Pub/Sub message and updates the requesting Secret data.

1. Secret in Secret mnager senda a Pub/Sub message to `secret-manager-notifications` Pub/Sub topic.
2. The Cloud Function is started.
3. The Cloud Function check if the value of the `eventType` attribute is set to `SECRET_ROTATE` and stops execution otherwise.
4. The Cloud Function check if the value of the `type` label is set to `service-account` and stops execution otherwise.
5. The Cloud Function reads the name of the service account from the latest version of a secret.
6. The Cloud Function generates new key for the service account.
7. The Cloud Function creates new secret version in Secret Manger, containing the newly created service account key.

## Cloud Function deployment

To deploy the Coud Function follow these steps:

1. Run `go mod vendor` inside the `development/gcp/cloud-functions/rotateserviceaccount/` directory.
2. Create `secret-manager-notifications` Pub/Sub topic.
2. Create `service-${PROJECT_ID_NUMBER}@gcp-sa-secretmanager.iam.gserviceaccount.com` service account with `roles/pubsub.publisher` role if it does not exist.
3. Use the following command to deploy the Cloud Function:
```bash
gcloud functions deploy rotate-secrets-service-account \
--region europe-west3 \
--trigger-topic secret-manager-notifications \
--runtime go113 \
--source ./ \
--timeout 10 \
--max-instances 10 \
--memory 128 \
--entry-point RotateServiceAccount
```

## Cloud Function usage

To setup secret for automatic rotation follow these steps:
1. Create new secret in Secret Manager with existing service account data.
2. Add `type: service-account` label to the secret.
3. Set `secret-manager-notifications` as a secret  Pub/Sub topic.
3. Set up rotation period for the secret.
