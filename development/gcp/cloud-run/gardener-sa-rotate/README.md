# Rotate Gardener service account secrets using Cloud Run

## Overview

Cloud Run app creates a new key for a GCP service account, updates the required secret data, and deletes old versions of a key. The function is triggered by a Pub/Sub message sent by a secret stored in Secret Manager.

1. Secret in Secret Manager sends a Pub/Sub message to `secret-manager-notifications` Pub/Sub topic.
2. Cloud Run app checks if the value of the **eventType** attribute is set to `SECRET_ROTATE`; if not, it stops its execution.
3. Cloud Run app checks if the value of the **type** label is set to `gardener-service-account`; if not, it stops its execution.
4. Cloud Run app checks if the values of the **kubeconfig-secret**, **gardener-secret**, and **gardener-secret-namespace** labels are set; if not, it stops its execution.
5. Cloud Run app authenticates to a cluster using the kubeconfig from the latest version of a secret provided in **kubeconfig-secret** label.
6. Cloud Run app reads the name of the service account from the latest version of a secret.
7. Cloud Run app generates a new key for the service account.
8. Cloud Run app creates a new secret version in Secret Manger, containing the newly created service account key.
9. Cloud Run app updates a secret in Gardener cluster, containing the newly created service account key.
10. Cloud Run app destroys old versions of a secret in Secret Manager.

## Cloud Run deployment

To deploy Cloud Run app follow these steps:

1. Create the `secret-manager-notifications` Pub/Sub topic, if it does not exist.
2. Create the `service-${PROJECT_NUMBER}@gcp-sa-secretmanager.iam.gserviceaccount.com` service account with the `roles/pubsub.publisher` role, if it does not exist.
3. Use the following command from this directory to deploy Cloud Run app:
```bash
gcloud run deploy rotate-gardener-secrets-service-account \
--region europe-west1 \
--timeout 10 \
--max-instances 1 \
--memory 128 \
--service-account sa-secret-update \
--image URL \
--ingress internal
```
4. Create Pub/Sub subscription 

## Cloud run usage

To setup an automatic rotation for a Secret Manager secret follow these steps:
1. Create a new secret in Secret Manager with the existing service account data.
2. Add `type: gardener-service-account` label to the secret.
3. Add `kubeconfig-secret` label containing name of the secret containing Gardener cluster kubeconfig to the secret.
4. Add `gardener-secret` label containing name of a Gardener secret containing service account credentials to the secret.
5. Add `gardener-secret-namespace` label containing name of a Gardener secret namespace to the secret.
6. Set `secret-manager-notifications` as a secret Pub/Sub topic.
7. Set up rotation period for the secret.


# Secret Manager secret labels

See the list of labels required for the function:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **type** | Yes | Type of secret, must be set to `gardener-service-account`. |
| **kubeconfig-secret** | Yes | Name of the Secret Manager secret containing kubeconfig. |
| **gardener-secret** | Yes | Name of the Gardener secret containing service account credentials. |
| **gardener-secret-namespace** | Yes | Name of the Gardener secret namespace containing service account credentials. |


# GET request parameters

See the list of GET arguments for the function:
| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **dry_run** | No | Enables dry run without updating secrets (defaults to false). |

