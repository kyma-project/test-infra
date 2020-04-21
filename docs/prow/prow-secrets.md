
# Prow Secrets

## Overview

Some jobs require using sensitive data. Encrypt the data using Key Management Service (KMS) and store it in Google Cloud Storage (GCS).
This document lists all types of secrets used in kyma-prow cluster as well as workloads cluster, where all the ProwJobs are executed.

>**NOTE:** All secrets are stored in the GCS bucket.



| Prow Secret   | Description | 
| :---------- | :---------------- | 
| **hmac-token**| It is used for validating webhooks. It is manually generated using `openssl rand -hex 20` command.| 
| **oauth-token**| It stores personal access token (called `Prow - Production`) used by the kyma-bot github user. | 
|**sap-slack-bot-token**| It is a token for publishing messages to the SAP CX workspace. Find more information [here](https://api.slack.com/docs/token-types#bot) |
|**workload-cluster**| It stores `workload-kyma-prow` cluster certificate. |


| Secret   | Description | 
| :---------- | :---------------- | 
| **whitesource** keys| are directly copied from the bucket when executing job. Secrets are not stored on the cluster. | 
| **github-integration** secrets| Are used to authorize Github applications configured in kyma-project organization (https://github.com/organizations/kyma-project/settings/applications).|
| **sa-***| Service Accounts used in pipelines. You can find more information [here](/docs/prow/authorization.md).| 
| **kyma-website-bot-zenhub-token**| stores the ZenHub token for the `kyma-website-bot` account.| 
| **kyma-bot-github-token**| stores personal access token (called `Prow - Job`) used by the kyma-bot github user.| 
| **kyma-guard-bot-github-token** | stores personal access token of the kyma-guard-bot Github account|
| **kyma-bot@sap.com**| stores credentials to the kyma-bot Github account. |
| **kyma-bot-npm-token** | it is a token for publishing npm packages in npmjs.com registry. Kyma-bot user credentials are uset to authenticate to the registry. The secret is used by `post-master-varkes` job. |
| **gardener-kyma-prow-kubeconfig**| it is a kubeconfig file, that allows connection to the Gardener `kyma-prow` project.| 
| **slack-nightly-token**| it is slack token that allows stability checker to push notifications to Slack. | 
| **sap-slack-bot-token** | it is a token for publishing messages to the SAP CX workspace. Find more information [here](https://api.slack.com/docs/token-types#bot.|
| **kyma-alerts-slack-api-url** | it is a token for publishing messages to the SAP CX workspace.  It is used by nightly and weekly ProwJobs.|
| **neighbors-alerts-slack-api-url** | publishes alerts to the private neighbors channel.|
| **kyma-azure-credential-*** | This set of secrets stores Azure subscription and service principal credentials. |
