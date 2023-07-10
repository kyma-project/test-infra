# Rotate Gardener service account secrets using Cloud Run

## Overview

The Cloud Run application creates a new key for a GCP service account, updates the required secret data, and deletes old versions of a key. The function is triggered by a Pub/Sub message sent by a secret stored in Secret Manager.

1. A secret in Secret Manager sends a Pub/Sub message to the `secret-manager-notifications` Pub/Sub topic.
2. The Cloud Run application checks if the value of the **eventType** attribute is set to `SECRET_ROTATE`. If not, it stops running.
3. The Cloud Run application checks if the value of the **type** label is set to `gardener-service-account`. If not, it stops running.
4. The Cloud Run application checks if the values of the **kubeconfig-secret**, **gardener-secret**, and **gardener-secret-namespace** labels are set. If not, it stops running.
5. The Cloud Run application authenticates to a cluster using the kubeconfig from the latest version of a secret provided in the **kubeconfig-secret** label.
6. The Cloud Run application reads the name of the service account from the latest version of a secret.
7. The Cloud Run application generates a new key for the service account.
8. The Cloud Run application creates a new secret version containing the newly created service account key in Secret Manger.
9. The Cloud Run application updates a secret containing the newly created service account key in the Gardener cluster.
10. The Cloud Run application deletes old versions of a key in IAM.
11. The Cloud Run application destroys old versions of a secret in Secret Manager.

## Cloud Run deployment

To deploy the Cloud Run application, follow the following steps:

1. Create the `secret-manager-notifications` Pub/Sub topic if it does not exist.
2. Create the `service-${PROJECT_NUMBER}@gcp-sa-secretmanager.iam.gserviceaccount.com` service account with the `roles/pubsub.publisher` role if it does not exist.
3. Use the following command to deploy the Cloud Run application:
```bash
gcloud run deploy rotate-gardener-secrets-service-account \
--region europe-west1 \
--timeout 600 \
--max-instances 1 \
--memory 128Mi \
--service-account sa-secret-update@sap-kyma-prow.iam.gserviceaccount.com \
--ingress internal \
--project sap-kyma-prow \
--allow-unauthenticated \
--image eu.gcr.io/kyma-project/test-infra/gardener-sa-rotate:v20221006-6fd98cfd
```
4. Create the push `rotate-gardener-secrets-service-account` Pub/Sub subscription on `secret-manager-notifications` topic pointing to the Cloud Run application URL.


## Cloud Run usage

To setup an automatic rotation for a Secret Manager secret, follow these steps:
1. Create a new secret in Secret Manager with the existing service account data.
2. Add the `type: gardener-service-account` label to the secret.
3. Add the `kubeconfig-secret` label with the name of the secret containing the Gardener cluster kubeconfig to the secret.
4. Add the `gardener-secret` label with the name of a Gardener secret containing service account credentials to the secret.
5. Add the `gardener-secret-namespace` label containing the name of a Gardener secret namespace to the secret.
6. Set `secret-manager-notifications` as a secret Pub/Sub topic.
7. Set up a rotation period for the secret.


# Secret Manager secret labels

See the list of labels required for the function:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **type** | Yes | The type of secret. It must be set to `gardener-service-account`. |
| **kubeconfig-secret** | Yes | The name of the Secret Manager secret containing the kubeconfig. |
| **gardener-secret** | Yes | The name of the Gardener secret containing service account credentials. |
| **gardener-secret-namespace** | Yes | The name of the Gardener secret namespace containing service account credentials. |


# GET request parameters

See the list of GET arguments for the function:
| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **dry_run** | No | Enables a dry run without updating secrets (defaults to false). |

