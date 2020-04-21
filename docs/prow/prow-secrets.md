
# Prow Secrets

## Overview

Some jobs require using sensitive data. Encrypt the data using Key Management Service (KMS) and store it in Google Cloud Storage (GCS).
This document lists all types of secrets used in kyma-prow cluster as well as workloads, where all the ProwJobs are executed.

>**NOTE:** All secrets are stored in the GCS bucket.



| Prow Secret   | Description | 
| :---------- | :---------------- | 
| `hmac-token`| It is used for validating webhooks. It is manually generated using `openssl rand -hex 20` command.| 
| `oauth-token`| It stores personal access token (called `Prow - Production`) used by the kyma-bot github user. | 
|`sap-slack-bot-token`| It is a token for publishing messages to the SAP CX workspace. Find more information [here](https://api.slack.com/docs/token-types#bot) |
|`workload-cluster`| It stores `workload-kyma-prow` cluster certificate. |


| Secret   | Description | 
| :---------- | :---------------- | 
| **whitesource** keys| Are directly copied from the bucket when executing job. Secrets are not stored on the cluster. | 
| **github-integration** secrets| Are used to authorize Github applications configured in [kyma-project organization.](https://github.com/organizations/kyma-project/settings/applications)|
| **sa-**| Service Accounts used in pipelines. You can find more information [here](/docs/prow/authorization.md).| 
| **kyma-website-bot-zenhub-token**| It stores the ZenHub token for the `kyma-website-bot` account.| 
| **kyma-bot-github-token**| It stores personal access token (called `Prow - Job`) used by the kyma-bot github user.| 
|**kyma-bot-npm-token** | It is a token for publishing npm packages, it is used by `post-master-varkes` job.| 
|**sap-slack-bot-token** | It is a token for publishing messages to the SAP CX workspace. Find more information [here](https://api.slack.com/docs/token-types#bot.|
| **gardener-kyma-prow-kubeconfig**| It is kubeconfig that allows connection to the Gardener `kyma-prow` project.| 
| **slack-nightly-token**| It is slack token that allows stability checker to push notifications to Slack. |
| | | 
                                                                                         |
